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

type Summary struct {
	AvgServerWaitingTime float64 `json:"avgServerWaitingTime"`
	AvgBodyDownloadTime  float64 `json:"avgBodyDownloadTime"`
	AvgTotalDuration     float64 `json:"avgTotalDuration"`
}

type TraceWriter struct {
	Entries   []TraceWriteEntry  `json:"entries"`
	Summaries map[string]Summary `json:"summaries"`
}

func NewTraceWriter() *TraceWriter {
	return &TraceWriter{
		Entries:   make([]TraceWriteEntry, 0, 10),
		Summaries: make(map[string]Summary),
	}
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

func (t *TraceWriter) CalculateSummaries() {
	groupedEntries := make(map[string][]TraceWriteEntry)

	for _, entry := range t.Entries {
		if entries, ok := groupedEntries[entry.Label]; ok {
			entries = append(entries, entry)
			groupedEntries[entry.Label] = entries
		} else {
			groupedEntries[entry.Label] = []TraceWriteEntry{entry}
		}
	}

	for label, entries := range groupedEntries {
		t.Summaries[label] = calculateSummary(entries)
	}
}

func calculateSummary(entries []TraceWriteEntry) Summary {
	avgWait := 0.0
	avgDownload := 0.0
	avgTotal := 0.0

	for _, entry := range entries {
		avgWait += entry.ServerWaitingTime
		avgDownload += entry.BodyDownloadTime
		avgTotal += entry.TotalDuration
	}

	avgWait /= float64(len(entries))
	avgDownload /= float64(len(entries))
	avgTotal /= float64(len(entries))

	return Summary{
		AvgServerWaitingTime: avgWait,
		AvgBodyDownloadTime:  avgDownload,
		AvgTotalDuration:     avgTotal,
	}

}

func (t *TraceWriter) Serialize() ([]byte, error) {
	return json.Marshal(t)
}
