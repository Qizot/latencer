package hls

import "encoding/json"

type TraceWriteEntry struct {
	Label             string  `json:"label"`
	Url               string  `json:"url"`
	Status            int     `json:"status"`
	ServerWaitingTime float64 `json:"serverWaitingTime"`
	BodyDownloadTime  float64 `json:"bodyDownloadTime"`
	TotalDuration     float64 `json:"totalDuration"`
}

type TraceWriter struct {
	Entries []TraceWriteEntry `json:"entries"`
}

func (t *TraceWriter) Write(label string, trace *RequestTraceResult) {
	t.Entries = append(t.Entries, TraceWriteEntry{
		Label:             label,
		Url:               trace.Url,
		Status:            trace.Status,
		ServerWaitingTime: trace.ServerWaitingTime.Seconds(),
		BodyDownloadTime:  trace.BodyDownloadTime.Seconds(),
		TotalDuration:     trace.TotalDuration.Seconds(),
	})
}

func (t *TraceWriter) Serialize() ([]byte, error) {
	return json.Marshal(t)
}
