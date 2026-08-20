package handler

import (
	"html/template"
	"log"
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
			value, err := store.GetGauge(metricName)

			if err != nil {
				res.WriteHeader(http.StatusNotFound)
				return
			}

			res.Write([]byte(strconv.FormatFloat(float64(value), 'f', -1, 64)))
			return
		}

		value, err := store.GetCounter(metricName)

		if err != nil {
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

		const tpl = `
			<!DOCTYPE html>
			<html>
				<head>
					<meta charset="UTF-8">
					<title>{{.Title}}</title>

					<style type="text/css">
					body {
						background: #000;
						color: #fff;
					}
					</style>
				</head>
				<body>
					<h1>Gauges:</h1>
					{{range $key, $value := .Gauges}}<div>{{ $key }}: {{ printf "%f" $value }}</div>{{end}}
					<h1>Counters:</h1>
					{{range $key, $value := .Counters}}<div>{{ $key }}: {{ printf "%d" $value }}</div>{{end}}
				</body>
			</html>`

		data := struct {
			Title    string
			Gauges   map[string]repository.Gauge
			Counters map[string]repository.Counter
		}{
			Title:    "Metrics",
			Gauges:   store.GetGauges(),
			Counters: store.GetCounters(),
		}

		t, err := template.New("webpage").Parse(tpl)

		if err != nil {
			log.Fatal(err)
		}

		err = t.Execute(res, data)
		if err != nil {
			log.Fatal(err)
		}
	}
}
