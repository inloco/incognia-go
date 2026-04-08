package incognia

type libMetrics struct {
	RequestID string `json:"rid,omitempty"`
	Endpoint  string `json:"ed"`
	Latency   int64  `json:"lt"`
}
