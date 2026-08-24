package logger

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	})
}

func TestTrustedSubnetAllowsAddressFromSubnet(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	srv := httptest.NewServer(TrustedSubnetMiddleware(subnet)(okHandler()))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(RealIPHeader, "192.168.1.42")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestTrustedSubnetRejectsForeignAddress(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	srv := httptest.NewServer(TrustedSubnetMiddleware(subnet)(okHandler()))
	defer srv.Close()

	testCases := []struct {
		name string
		ip   string
	}{
		{name: "чужая подсеть", ip: "10.0.0.1"},
		{name: "пустой заголовок", ip: ""},
		{name: "не IP-адрес", ip: "не адрес"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, reqErr := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, reqErr)

			if tc.ip != "" {
				req.Header.Set(RealIPHeader, tc.ip)
			}

			res, doErr := http.DefaultClient.Do(req)
			require.NoError(t, doErr)
			defer res.Body.Close()

			assert.Equal(t, http.StatusForbidden, res.StatusCode)
		})
	}
}

func TestTrustedSubnetSkippedWhenNotSet(t *testing.T) {
	srv := httptest.NewServer(TrustedSubnetMiddleware(nil)(okHandler()))
	defer srv.Close()

	res, err := http.Get(srv.URL)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode, "без подсети ограничений нет")
}
