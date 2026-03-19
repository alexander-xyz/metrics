package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
)

func UpdateMetricHandler(store repository.Updater) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-type", "text/plain")

		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "id")
		metricValue := chi.URLParam(req, "value")

		if metricType != "gauge" && metricType != "counter" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		if metricName == "" || metricValue == "" {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		if metricType == "gauge" {
			value, err := strconv.ParseFloat(metricValue, 64)

			if err != nil {
				res.WriteHeader(http.StatusBadRequest)
				return
			}

			store.UpdateGauge(metricName, repository.Gauge(value))
			return
		}

		value, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		store.UpdateCounter(metricName, repository.Counter(value))
	}
}

func GetMetricHandler(store repository.Getter) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-type", "text/plain")

		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "id")

		if metricType != "gauge" && metricType != "counter" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		if metricName == "" {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		if metricType == "gauge" {
			value, ok := store.GetGauge(metricName)

			if !ok {
				res.WriteHeader(http.StatusNotFound)
				return
			}

			res.Write([]byte(strconv.FormatFloat(float64(value), 'f', -1, 64)))
			return
		}

		value, ok := store.GetCounter(metricName)

		if !ok {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Write([]byte(strconv.FormatInt(int64(value), 10)))
		return
	}
}

func GetMetricsHandler(store repository.Getter) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-type", "text/html")

		res.Write([]byte("<h1>Gauges:</h1>"))
		for t, v := range store.GetGauges() {
			text := fmt.Sprintf("<p>%s:%g</p>", t, v)
			res.Write([]byte(text))
		}

		res.Write([]byte("<h1>Counters:</h1>"))
		for t, v := range store.GetCounters() {
			text := fmt.Sprintf("<p>%s:%d</p>", t, v)
			res.Write([]byte(text))
		}
	}
}
