package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/alexander-xyz/metrics/internal/audit"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type auditorStub struct {
	events []audit.Event
}

func (a *auditorStub) Notify(event audit.Event) {
	a.events = append(a.events, event)
}

func TestAuditOnMetricUpdates(t *testing.T) {
	auditor := &auditorStub{}
	router, err := GetRouter(repository.NewMemStorage(), nil, "", auditor)
	require.NoError(t, err)

	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New().SetBaseURL(srv.URL)

	_, err = client.R().Post("/update/gauge/Alloc/12.5")
	require.NoError(t, err)

	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"id":"PollCount","type":"counter","delta":3}`).
		Post("/update")
	require.NoError(t, err)

	_, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`[{"id":"Frees","type":"gauge","value":1},{"id":"Sys","type":"gauge","value":2}]`).
		Post("/updates")
	require.NoError(t, err)

	require.Len(t, auditor.events, 3)
	assert.Equal(t, []string{"Alloc"}, auditor.events[0].Metrics)
	assert.Equal(t, []string{"PollCount"}, auditor.events[1].Metrics)
	assert.Equal(t, []string{"Frees", "Sys"}, auditor.events[2].Metrics)

	for _, event := range auditor.events {
		assert.NotZero(t, event.Timestamp)
		assert.NotEmpty(t, event.IPAddress)
	}
}

func TestAuditSkippedOnFailedUpdate(t *testing.T) {
	auditor := &auditorStub{}
	router, err := GetRouter(repository.NewMemStorage(), nil, "", auditor)
	require.NoError(t, err)

	srv := httptest.NewServer(router)
	defer srv.Close()

	client := resty.New().SetBaseURL(srv.URL)

	_, err = client.R().Post("/update/gauge/Alloc/nonsense")
	require.NoError(t, err)

	_, err = client.R().Get("/value/gauge/Alloc")
	require.NoError(t, err)

	assert.Empty(t, auditor.events)
}
