package incognia

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func periodicTestClient(t *testing.T, handler http.HandlerFunc) *TokenClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client := NewTokenClient(&TokenClientConfig{Timeout: 3 * time.Second})
	client.tokenEndpoint = server.URL
	return client
}

func writePeriodicToken(w http.ResponseWriter, value string, lifetime int) {
	fmt.Fprintf(w, `{"access_token":%q,"token_type":"Bearer","expires_in":"%d"}`, value, lifetime)
}

func TestPeriodicProviderInitialTokenAndClose(t *testing.T) {
	var calls atomic.Int32
	client := periodicTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		writePeriodicToken(w, "initial", 3600)
	})
	p, err := NewPeriodicRefreshTokenProvider(client)
	require.NoError(t, err)
	defer p.Close()
	token, err := p.GetToken()
	require.NoError(t, err)
	require.Equal(t, "initial", token.(*accessToken).AccessToken)
	require.EqualValues(t, 1, calls.Load())

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Close()
		}()
	}
	wg.Wait()
	_, err = p.GetToken()
	require.NoError(t, err)
	require.EqualValues(t, 1, calls.Load())
}

func TestPeriodicProviderInitialFailure(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			client := periodicTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			})
			p, err := NewPeriodicRefreshTokenProvider(client)
			require.Error(t, err)
			require.Nil(t, p)
			if status == http.StatusUnauthorized {
				require.Equal(t, ErrInvalidCredentials, err)
			}
		})
	}
	client := periodicTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writePeriodicToken(w, "expired", 0)
	})
	p, err := NewPeriodicRefreshTokenProvider(client)
	require.Equal(t, ErrTokenExpired, err)
	require.Nil(t, p)
}

func TestPeriodicProviderRefreshDelay(t *testing.T) {
	now := time.Unix(100, 0)
	p := &PeriodicRefreshTokenProvider{config: PeriodicRefreshTokenProviderConfig{
		RefreshBefore: time.Minute,
		RetryInterval: 5 * time.Second,
	}}
	for _, tt := range []struct {
		lifetime int64
		want     time.Duration
	}{
		{3600, 59 * time.Minute},
		{30, 15 * time.Second},
		{60, 30 * time.Second},
		{0, 5 * time.Second},
		{-1, 5 * time.Second},
	} {
		token := &accessToken{CreatedAt: now.Unix(), ExpiresIn: tt.lifetime}
		require.Equal(t, tt.want, p.refreshDelay(token, now))
	}
}

func TestPeriodicProviderRefreshDoesNotBlockReadsAndCloseCancels(t *testing.T) {
	var calls atomic.Int32
	started := make(chan struct{})
	canceled := make(chan struct{})
	client := periodicTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			writePeriodicToken(w, "initial", 2)
			return
		}
		close(started)
		<-r.Context().Done()
		close(canceled)
	})
	p, err := NewPeriodicRefreshTokenProvider(client)
	require.NoError(t, err)
	defer p.Close()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("background refresh did not start")
	}
	read := make(chan error, 1)
	go func() {
		_, err := p.GetToken()
		read <- err
	}()
	select {
	case err := <-read:
		require.NoError(t, err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("cached read blocked on refresh")
	}
	closed := make(chan struct{})
	go func() {
		p.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close did not cancel the refresh")
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("token request was not canceled")
	}
	require.EqualValues(t, 2, calls.Load())
}

func TestPeriodicProviderRetriesAndRetainsToken(t *testing.T) {
	var calls atomic.Int32
	var recoverRequests atomic.Bool
	attempts := make(chan time.Time, 100)
	client := periodicTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			writePeriodicToken(w, "initial", 2)
			return
		}
		attempts <- time.Now()
		if recoverRequests.Load() {
			writePeriodicToken(w, "recovered", 3600)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})
	p, err := NewPeriodicRefreshTokenProviderWithConfig(client, PeriodicRefreshTokenProviderConfig{
		RetryInterval: 50 * time.Millisecond,
	})
	require.NoError(t, err)
	defer p.Close()
	initial, err := p.GetToken()
	require.NoError(t, err)
	nextAttempt := func() time.Time {
		t.Helper()
		select {
		case at := <-attempts:
			return at
		case <-time.After(3 * time.Second):
			t.Fatal("refresh attempt did not arrive")
			return time.Time{}
		}
	}
	first := nextAttempt()
	token, err := p.GetToken()
	require.NoError(t, err)
	require.Same(t, initial, token)
	second := nextAttempt()
	require.True(t, second.Sub(first) >= 50*time.Millisecond, "refresh retried too early")

	require.Eventually(t, func() bool {
		_, err := p.GetToken()
		return err == ErrTokenExpired
	}, 3*time.Second, 10*time.Millisecond)
	// Reads of an expired token must not synchronously contact the token endpoint.
	p.manual.tokenMutex.RLock()
	cached := p.manual.token
	p.manual.tokenMutex.RUnlock()
	require.Same(t, initial, cached)

	recoverRequests.Store(true)
	require.Eventually(t, func() bool {
		token, err := p.GetToken()
		return err == nil && token.(*accessToken).AccessToken == "recovered"
	}, time.Second, 10*time.Millisecond)
	require.NoError(t, p.Close())
	count := calls.Load()
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, count, calls.Load())
}
