package audit

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testObserver struct {
	id     string
	events []Event
}

func (o *testObserver) ID() string { return o.id }

func (o *testObserver) Update(event Event) {
	o.events = append(o.events, event)
}

func TestPublisherNotifiesObservers(t *testing.T) {
	publisher := NewPublisher()
	first := &testObserver{id: "first"}
	second := &testObserver{id: "second"}

	publisher.Register(first)
	publisher.Register(second)

	event := Event{Timestamp: 100, Metrics: []string{"Alloc"}, IPAddress: "192.168.0.42"}
	publisher.Notify(event)

	assert.Equal(t, []Event{event}, first.events)
	assert.Equal(t, []Event{event}, second.events)
}

func TestPublisherDeregister(t *testing.T) {
	publisher := NewPublisher()
	observer := &testObserver{id: "first"}

	publisher.Register(observer)
	publisher.Deregister(observer)
	publisher.Notify(Event{Timestamp: 1})

	assert.Empty(t, observer.events)
}

func TestFileObserverAppendsEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	observer := NewFileObserver(path, func(err error) {
		t.Errorf("unexpected error: %v", err)
	})

	observer.Update(Event{Timestamp: 1, Metrics: []string{"Alloc"}, IPAddress: "127.0.0.1"})
	observer.Update(Event{Timestamp: 2, Metrics: []string{"Frees"}, IPAddress: "127.0.0.2"})

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.Equal(t,
		"{\"ts\":1,\"metrics\":[\"Alloc\"],\"ip_address\":\"127.0.0.1\"}\n"+
			"{\"ts\":2,\"metrics\":[\"Frees\"],\"ip_address\":\"127.0.0.2\"}\n",
		string(data))
}

func TestFileObserverReportsError(t *testing.T) {
	var got error

	observer := NewFileObserver(filepath.Join(t.TempDir(), "missing", "audit.log"), func(err error) {
		got = err
	})
	observer.Update(Event{Timestamp: 1})

	assert.Error(t, got)
}

func TestHTTPObserverSendsEvents(t *testing.T) {
	var (
		body   []byte
		method string
	)

	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		method = req.Method
		body, _ = io.ReadAll(req.Body)
		res.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	observer := NewHTTPObserver(server.URL, func(err error) {
		t.Errorf("unexpected error: %v", err)
	})
	observer.Update(Event{Timestamp: 5, Metrics: []string{"Alloc", "Frees"}, IPAddress: "10.0.0.1"})

	assert.Equal(t, http.MethodPost, method)

	var event Event
	require.NoError(t, json.Unmarshal(body, &event))
	assert.Equal(t, Event{Timestamp: 5, Metrics: []string{"Alloc", "Frees"}, IPAddress: "10.0.0.1"}, event)
}

func TestHTTPObserverReportsBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	var got error

	observer := NewHTTPObserver(server.URL, func(err error) {
		got = err
	})
	observer.Update(Event{Timestamp: 1})

	assert.Error(t, got)
}
