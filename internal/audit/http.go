package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPObserver struct {
	url    string
	client *http.Client

	onError func(error)
}

func NewHTTPObserver(url string, onError func(error)) *HTTPObserver {
	return &HTTPObserver{
		url:     url,
		client:  &http.Client{Timeout: 5 * time.Second},
		onError: onError,
	}
}

func (o *HTTPObserver) ID() string {
	return "url:" + o.url
}

func (o *HTTPObserver) Update(event Event) {
	if err := o.send(event); err != nil && o.onError != nil {
		o.onError(err)
	}
}

func (o *HTTPObserver) send(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	response, err := o.client.Post(o.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("send audit event: %w", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("audit receiver returned %d", response.StatusCode)
	}

	return nil
}
