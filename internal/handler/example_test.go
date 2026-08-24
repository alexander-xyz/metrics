package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/alexander-xyz/metrics/internal/handler"
	"github.com/alexander-xyz/metrics/internal/repository"
)

// exampleServer поднимает сервер метрик с хранилищем в памяти.
func exampleServer() *httptest.Server {
	router, err := handler.GetRouter(repository.NewMemStorage(), nil, "", nil, nil, nil)
	if err != nil {
		panic(err)
	}

	return httptest.NewServer(router)
}

// post отправляет запрос и печатает код ответа вместе с телом.
func post(url, contentType, body string) {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	res, err := http.Post(url, contentType, reader)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(strings.TrimSpace(fmt.Sprintf("%d %s", res.StatusCode, data)))
}

// Метрику можно отправить параметрами пути: тип, имя и значение.
func Example_updateMetric() {
	srv := exampleServer()
	defer srv.Close()

	post(srv.URL+"/update/gauge/Alloc/12.5", "text/plain", "")
	post(srv.URL+"/update/counter/PollCount/3", "text/plain", "")

	// Метрика неизвестного типа не принимается.
	post(srv.URL+"/update/unknown/Alloc/1", "text/plain", "")

	// Output:
	// 200
	// 200
	// 400
}

// Значение сохранённой метрики читается по её типу и имени.
func Example_getMetric() {
	srv := exampleServer()
	defer srv.Close()

	post(srv.URL+"/update/gauge/Alloc/12.5", "text/plain", "")

	res, err := http.Get(srv.URL + "/value/gauge/Alloc")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	fmt.Println(strings.TrimSpace(fmt.Sprintf("%d %s", res.StatusCode, data)))

	// Output:
	// 200
	// 200 12.5
}

// Метрику можно отправить и в формате JSON — в ответе придёт сохранённое значение.
func Example_updateMetricJSON() {
	srv := exampleServer()
	defer srv.Close()

	post(srv.URL+"/update", "application/json", `{"id":"Alloc","type":"gauge","value":12.5}`)

	// Counter суммируется с предыдущим значением.
	post(srv.URL+"/update", "application/json", `{"id":"PollCount","type":"counter","delta":2}`)
	post(srv.URL+"/update", "application/json", `{"id":"PollCount","type":"counter","delta":3}`)

	// Output:
	// 200 {"id":"Alloc","type":"gauge","value":12.5}
	// 200 {"id":"PollCount","type":"counter","delta":2}
	// 200 {"id":"PollCount","type":"counter","delta":5}
}

// Значение метрики можно запросить в формате JSON.
func Example_getMetricJSON() {
	srv := exampleServer()
	defer srv.Close()

	post(srv.URL+"/update", "application/json", `{"id":"Alloc","type":"gauge","value":12.5}`)
	post(srv.URL+"/value", "application/json", `{"id":"Alloc","type":"gauge"}`)

	// Неизвестная метрика не находится.
	post(srv.URL+"/value", "application/json", `{"id":"Missing","type":"gauge"}`)

	// Output:
	// 200 {"id":"Alloc","type":"gauge","value":12.5}
	// 200 {"id":"Alloc","type":"gauge","value":12.5}
	// 404 metric not found
}

// Несколько метрик отправляются одним пакетом.
func Example_updateMetrics() {
	srv := exampleServer()
	defer srv.Close()

	batch := `[{"id":"Alloc","type":"gauge","value":12.5},{"id":"PollCount","type":"counter","delta":3}]`
	post(srv.URL+"/updates", "application/json", batch)

	// Пустой пакет не принимается.
	post(srv.URL+"/updates", "application/json", `[]`)

	// Output:
	// 200 {"status":"ok"}
	// 400 empty batch
}
