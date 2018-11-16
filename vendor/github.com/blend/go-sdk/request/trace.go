package request

import (
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

// HTTPTrace is timing information for the full http call.
type HTTPTrace struct {
	GetConn     time.Time `json:"getConn"`
	GotConn     time.Time `json:"gotConn"`
	PutIdleConn time.Time `json:"putIdleConn"`

	DNSStart time.Time `json:"dnsStart"`
	DNSDone  time.Time `json:"dnsDone"`

	ConnectStart time.Time `json:"connectStart"`
	ConnectDone  time.Time `json:"connectDone"`

	TLSHandshakeStart time.Time `json:"tlsHandshakeStart"`
	TLSHandshakeDone  time.Time `json:"tlsHandshakeDone"`

	WroteHeaders         time.Time `json:"wroteHeaders"`
	WroteRequest         time.Time `json:"wroteRequest"`
	GotFirstResponseByte time.Time `json:"gotFirstResponseByte"`

	DNSElapsed          time.Duration `json:"dnsElapsed"`
	TLSHandshakeElapsed time.Duration `json:"tlsHandshakeElapsed"`
	DialElapsed         time.Duration `json:"dialElapsed"`
	RequestElapsed      time.Duration `json:"requestElapsed"`
	ServerElapsed       time.Duration `json:"severElapsed"`
}

// Trace returns the trace binder.
func (ht *HTTPTrace) Trace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(_ string) {
			ht.GetConn = now()
		},
		GotConn: func(_ httptrace.GotConnInfo) {
			ht.GotConn = now()
		},
		PutIdleConn: func(_ error) {
			ht.PutIdleConn = now()
		},
		GotFirstResponseByte: func() {
			ht.GotFirstResponseByte = now()
			ht.ServerElapsed = ht.GotFirstResponseByte.Sub(ht.WroteRequest)
		},
		DNSStart: func(_ httptrace.DNSStartInfo) {
			ht.DNSStart = now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			ht.DNSDone = now()
			ht.DNSElapsed = ht.DNSDone.Sub(ht.DNSStart)
		},
		ConnectStart: func(_, _ string) {
			ht.ConnectStart = now()
		},
		ConnectDone: func(_, _ string, _ error) {
			ht.ConnectDone = now()
			ht.DialElapsed = ht.ConnectDone.Sub(ht.ConnectStart)
		},
		TLSHandshakeStart: func() {
			ht.TLSHandshakeStart = now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			ht.TLSHandshakeDone = now()
			ht.TLSHandshakeElapsed = ht.TLSHandshakeDone.Sub(ht.TLSHandshakeStart)
		},
		WroteHeaders: func() {
			ht.WroteHeaders = now()
		},
		WroteRequest: func(_ httptrace.WroteRequestInfo) {
			ht.WroteRequest = now()
			if !ht.ConnectDone.IsZero() {
				ht.RequestElapsed = ht.WroteRequest.Sub(ht.ConnectDone)
			} else if !ht.GetConn.IsZero() {
				ht.RequestElapsed = ht.WroteRequest.Sub(ht.GetConn)
			} else if !ht.GotConn.IsZero() {
				ht.RequestElapsed = ht.WroteRequest.Sub(ht.GotConn)
			}
		},
	}
}
