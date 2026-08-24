package handler

import (
	"net"
	"net/http"
	"time"

	"github.com/alexander-xyz/metrics/internal/audit"
)

type Auditor interface {
	Notify(event audit.Event)
}

func notifyAudit(auditor Auditor, req *http.Request, metrics []string) {
	if auditor == nil {
		return
	}

	auditor.Notify(audit.Event{
		Timestamp: time.Now().Unix(),
		Metrics:   metrics,
		IPAddress: requestIP(req),
	})
}

func requestIP(req *http.Request) string {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}

	return host
}
