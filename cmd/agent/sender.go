package main

import (
	"fmt"
	"net/http"
)

func sendMetrics(store MemStorage) error {
	for t, v := range store.gauges {
		url := fmt.Sprintf("%s/update/gauge/%s/%g", server, t, v)
		_, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
	}

	for t, v := range store.counters {
		url := fmt.Sprintf("%s/update/counter/%s/%d", server, t, v)
		_, err := http.Post(url, "text/plain", nil)
		if err != nil {
			return err
		}
	}

	return nil
}
