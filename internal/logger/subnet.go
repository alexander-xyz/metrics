package logger

import (
	"net"
	"net/http"
)

// RealIPHeader — заголовок, в котором клиент передаёт свой IP-адрес.
const RealIPHeader = "X-Real-IP"

// TrustedSubnetMiddleware пропускает только запросы с IP-адресом
// из доверенной подсети. Адрес берётся из заголовка X-Real-IP.
// При пустой подсети запросы проходят без ограничений.
func TrustedSubnetMiddleware(subnet *net.IPNet) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == nil {
				h.ServeHTTP(w, r)
				return
			}

			ip := net.ParseIP(r.Header.Get(RealIPHeader))
			if ip == nil || !subnet.Contains(ip) {
				http.Error(w, "untrusted client address", http.StatusForbidden)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
