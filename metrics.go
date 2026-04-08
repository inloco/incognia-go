package incognia

import (
	"encoding/base64"
	"encoding/json"
)

const metricsHeader = "ICG-API-METRICS"

type libMetrics struct {
	RequestID string `json:"rid,omitempty"`
	Endpoint  string `json:"ed"`
	Latency   int64  `json:"lt"`
}

func (m *libMetrics) encode() string {
	b, _ := json.Marshal(m)
	return base64.URLEncoding.EncodeToString(b)
}

func decodeLibMetrics(s string) (*libMetrics, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var m libMetrics
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
