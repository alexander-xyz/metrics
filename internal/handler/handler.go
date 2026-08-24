// Package handler содержит HTTP-хендлеры сервера метрик и сборку роутера.
package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/alexander-xyz/metrics/internal/repository"
	"github.com/go-chi/chi/v5"
)

// UpdateMetricHandler возвращает хендлер приёма метрики из параметров пути:
// POST /update/{type}/{id}/{value}. Поддерживаются типы gauge и counter.
func UpdateMetricHandler(store repository.Updater, auditor Auditor) http.HandlerFunc {
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

			if err := store.UpdateGauge(req.Context(), metricName, repository.Gauge(value)); err != nil {
				log.Print(err)
				http.Error(res, "cannot store metric", http.StatusInternalServerError)

				return
			}

			notifyAudit(auditor, req, []string{metricName})

			return
		}

		value, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := store.UpdateCounter(req.Context(), metricName, repository.Counter(value)); err != nil {
			log.Print(err)
			http.Error(res, "cannot store metric", http.StatusInternalServerError)

			return
		}

		notifyAudit(auditor, req, []string{metricName})
	}
}

// GetMetricHandler возвращает хендлер чтения значения метрики:
// GET /value/{type}/{id}. Если метрики нет, отвечает 404.
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
			value, err := store.GetGauge(req.Context(), metricName)

			if err != nil {
				res.WriteHeader(http.StatusNotFound)
				return
			}

			res.Write([]byte(strconv.FormatFloat(float64(value), 'f', -1, 64)))
			return
		}

		value, err := store.GetCounter(req.Context(), metricName)

		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Write([]byte(strconv.FormatInt(int64(value), 10)))
		return
	}
}

const metricsPageTpl = `
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

// GetMetricsHandler возвращает хендлер главной страницы со списком всех метрик
// в виде HTML-таблицы. Шаблон разбирается один раз при создании хендлера.
func GetMetricsHandler(store repository.Getter) (http.HandlerFunc, error) {
	t, err := template.New("webpage").Parse(metricsPageTpl)
	if err != nil {
		return nil, fmt.Errorf("parse metrics page template: %w", err)
	}

	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-type", "text/html")

		gauges, err := store.GetGauges(req.Context())
		if err != nil {
			log.Print(err)
			http.Error(res, "cannot read metrics", http.StatusInternalServerError)

			return
		}

		counters, err := store.GetCounters(req.Context())
		if err != nil {
			log.Print(err)
			http.Error(res, "cannot read metrics", http.StatusInternalServerError)

			return
		}

		data := struct {
			Title    string
			Gauges   map[string]repository.Gauge
			Counters map[string]repository.Counter
		}{
			Title:    "Metrics",
			Gauges:   gauges,
			Counters: counters,
		}

		if err := t.Execute(res, data); err != nil {
			log.Print(err)
			http.Error(res, "internal server error", http.StatusInternalServerError)
		}
	}, nil
}

func decodeMetric(res http.ResponseWriter, req *http.Request) (models.Metrics, bool) {
	var metric models.Metrics

	if err := json.NewDecoder(req.Body).Decode(&metric); err != nil {
		http.Error(res, "cannot decode request body", http.StatusBadRequest)
		return metric, false
	}

	if metric.MType != models.Gauge && metric.MType != models.Counter {
		http.Error(res, "unsupported metric type", http.StatusBadRequest)
		return metric, false
	}

	if metric.ID == "" {
		http.Error(res, "empty metric name", http.StatusNotFound)
		return metric, false
	}

	return metric, true
}

func writeMetric(res http.ResponseWriter, metric models.Metrics) {
	res.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(res).Encode(metric); err != nil {
		log.Print(err)
	}
}

// UpdateMetricJSONHandler возвращает хендлер приёма одной метрики в формате JSON:
// POST /update. В ответе отдаётся сохранённое значение метрики.
func UpdateMetricJSONHandler(store Storage, auditor Auditor) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metric, ok := decodeMetric(res, req)
		if !ok {
			return
		}

		if metric.MType == models.Gauge {
			if metric.Value == nil {
				http.Error(res, "empty gauge value", http.StatusBadRequest)
				return
			}

			if err := store.UpdateGauge(req.Context(), metric.ID, repository.Gauge(*metric.Value)); err != nil {
				log.Print(err)
				http.Error(res, "cannot store metric", http.StatusInternalServerError)

				return
			}

			value, err := store.GetGauge(req.Context(), metric.ID)
			if err != nil {
				http.Error(res, "metric not found", http.StatusNotFound)
				return
			}

			stored := float64(value)
			metric.Value = &stored
			writeMetric(res, metric)
			notifyAudit(auditor, req, []string{metric.ID})

			return
		}

		if metric.Delta == nil {
			http.Error(res, "empty counter value", http.StatusBadRequest)
			return
		}

		if err := store.UpdateCounter(req.Context(), metric.ID, repository.Counter(*metric.Delta)); err != nil {
			log.Print(err)
			http.Error(res, "cannot store metric", http.StatusInternalServerError)

			return
		}

		delta, err := store.GetCounter(req.Context(), metric.ID)
		if err != nil {
			http.Error(res, "metric not found", http.StatusNotFound)
			return
		}

		stored := int64(delta)
		metric.Delta = &stored
		writeMetric(res, metric)
		notifyAudit(auditor, req, []string{metric.ID})
	}
}

// GetMetricJSONHandler возвращает хендлер чтения метрики в формате JSON:
// POST /value. Тело запроса содержит идентификатор и тип метрики.
func GetMetricJSONHandler(store repository.Getter) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		metric, ok := decodeMetric(res, req)
		if !ok {
			return
		}

		if metric.MType == models.Gauge {
			value, err := store.GetGauge(req.Context(), metric.ID)
			if err != nil {
				http.Error(res, "metric not found", http.StatusNotFound)
				return
			}

			stored := float64(value)
			metric.Value = &stored
			writeMetric(res, metric)

			return
		}

		delta, err := store.GetCounter(req.Context(), metric.ID)
		if err != nil {
			http.Error(res, "metric not found", http.StatusNotFound)
			return
		}

		stored := int64(delta)
		metric.Delta = &stored
		writeMetric(res, metric)
	}
}

// PingHandler возвращает хендлер проверки соединения с базой данных: GET /ping.
func PingHandler(db *sql.DB) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if db == nil {
			http.Error(res, "database is not configured", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			log.Print(err)
			http.Error(res, "database is unavailable", http.StatusInternalServerError)

			return
		}

		res.WriteHeader(http.StatusOK)
	}
}

// UpdateMetricsJSONHandler возвращает хендлер приёма пакета метрик:
// POST /updates. Метрики сохраняются одной транзакцией.
func UpdateMetricsJSONHandler(store Storage, auditor Auditor) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var metrics []models.Metrics

		if err := json.NewDecoder(req.Body).Decode(&metrics); err != nil {
			http.Error(res, "cannot decode request body", http.StatusBadRequest)
			return
		}

		if len(metrics) == 0 {
			http.Error(res, "empty batch", http.StatusBadRequest)
			return
		}

		for _, metric := range metrics {
			if metric.MType != models.Gauge && metric.MType != models.Counter {
				http.Error(res, "unsupported metric type", http.StatusBadRequest)
				return
			}

			if metric.ID == "" {
				http.Error(res, "empty metric name", http.StatusNotFound)
				return
			}
		}

		if err := store.UpdateBatch(req.Context(), metrics); err != nil {
			log.Print(err)
			http.Error(res, "cannot store metrics", http.StatusInternalServerError)

			return
		}

		names := make([]string, 0, len(metrics))
		for _, metric := range metrics {
			names = append(names, metric.ID)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)

		if _, err := res.Write([]byte(`{"status":"ok"}`)); err != nil {
			log.Print(err)
		}

		notifyAudit(auditor, req, names)
	}
}
