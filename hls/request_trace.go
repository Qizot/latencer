// TODO: this could be in common package instead
package hls

import (
	"net/http"
	"net/http/httptrace"
	"time"
)

type RequestTrace struct {
	url              string
	status           int
	requestStartedAt time.Time
	connectAt        time.Time
	connectedAt      time.Time
	firstByteAt      time.Time
	downloadedAt     time.Time
	clientTrace      *httptrace.ClientTrace
}

type RequestTraceResult struct {
	Url               string
	Status            int
	ServerWaitingTime time.Duration
	BodyDownloadTime  time.Duration
	TotalDuration     time.Duration
}

func (trace *RequestTrace) GetClientTrace() *httptrace.ClientTrace {
	return trace.clientTrace
}

func (trace *RequestTrace) SetDownloaded() {
	trace.downloadedAt = time.Now()
}
func (trace *RequestTrace) SetStatus(status int) {
	trace.status = status
}

func (trace *RequestTrace) SetRequestStart() {
	trace.requestStartedAt = time.Now()
}

func (trace *RequestTrace) ServerWaitingTime() time.Duration {
	if trace.connectAt.IsZero() {
		return trace.firstByteAt.Sub(trace.requestStartedAt)
	} else {
		return trace.firstByteAt.Sub(trace.connectAt)
	}
}

func (trace *RequestTrace) BodyDownloadTime() time.Duration {
	return time.Now().Sub(trace.firstByteAt)
}

func (trace *RequestTrace) TotalTime() time.Duration {
	if trace.connectAt.IsZero() {
		return trace.downloadedAt.Sub(trace.requestStartedAt)
	} else {
		return trace.downloadedAt.Sub(trace.connectAt)
	}

}

func (trace *RequestTrace) ToRequestTraceResult() *RequestTraceResult {
	return &RequestTraceResult{
		Url:               trace.url,
		Status:            trace.status,
		ServerWaitingTime: trace.ServerWaitingTime(),
		BodyDownloadTime:  trace.BodyDownloadTime(),
		TotalDuration:     trace.TotalTime(),
	}
}

func newRequestTrace(req *http.Request) *RequestTrace {
	requestTrace := &RequestTrace{url: req.URL.String()}

	clientTrace := &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) {
			requestTrace.connectAt = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if err != nil {
				requestTrace.connectedAt = time.Now()
			}
		},
		GotFirstResponseByte: func() {
			requestTrace.firstByteAt = time.Now()
		},
	}

	requestTrace.clientTrace = clientTrace

	return requestTrace
}
