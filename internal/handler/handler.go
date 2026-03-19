package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/alexander-xyz/metrics/internal/repository"
)

func UpdateMetricHandler(store repository.Updater) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-type", "text/plain")

		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		path := req.URL.Path
		parts := strings.Split(path, "/")

		if len(parts) != 5 {
			res.WriteHeader(http.StatusNotFound)

			return
		}

		metricType := parts[2]  // "gauge"
		metricName := parts[3]  // "temperature"
		metricValue := parts[4] // "1.5"

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
			res.WriteHeader(http.StatusOK)
			return
		}

		value, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		store.UpdateCounter(metricName, repository.Counter(value))

		res.WriteHeader(http.StatusOK)
		return
	}
}
