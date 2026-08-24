package pool

import (
	"sync"
	"testing"

	"github.com/alexander-xyz/metrics/internal/audit"
	models "github.com/alexander-xyz/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// counted считает вызовы Reset, чтобы проверить сброс при возврате.
type counted struct {
	value  int
	resets int
}

func (c *counted) Reset() {
	c.value = 0
	c.resets++
}

func TestGetCreatesObject(t *testing.T) {
	p := New[counted, *counted]()

	value := p.Get()
	require.NotNil(t, value)
	assert.Zero(t, value.value)
}

func TestPutResetsObject(t *testing.T) {
	p := New[counted, *counted]()

	value := p.Get()
	value.value = 42

	p.Put(value)

	assert.Zero(t, value.value, "состояние сброшено при возврате")
	assert.Equal(t, 1, value.resets)
}

func TestPutIgnoresNil(t *testing.T) {
	p := New[counted, *counted]()

	assert.NotPanics(t, func() { p.Put(nil) })
}

func TestPoolReusesObject(t *testing.T) {
	p := New[counted, *counted]()

	first := p.Get()
	first.value = 7
	p.Put(first)

	second := p.Get()
	assert.Zero(t, second.value, "объект выдаётся уже сброшенным")
}

func TestPoolWithGeneratedResetMethods(t *testing.T) {
	metrics := New[models.Metrics, *models.Metrics]()

	metric := metrics.Get()
	metric.ID = "Alloc"
	metric.MType = models.Gauge
	metrics.Put(metric)

	assert.Empty(t, metric.ID, "метод Reset сгенерирован cmd/reset")
	assert.Empty(t, metric.MType)

	events := New[audit.Event, *audit.Event]()

	event := events.Get()
	event.IPAddress = "127.0.0.1"
	event.Metrics = append(event.Metrics, "Alloc")
	events.Put(event)

	assert.Empty(t, event.IPAddress)
	assert.Empty(t, event.Metrics)
}

func TestPoolIsSafeForConcurrentUse(t *testing.T) {
	p := New[counted, *counted]()

	var wg sync.WaitGroup

	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < 200; j++ {
				value := p.Get()
				value.value = j
				p.Put(value)
			}
		}()
	}

	wg.Wait()
}

func BenchmarkPoolGetPut(b *testing.B) {
	p := New[models.Metrics, *models.Metrics]()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		metric := p.Get()
		metric.ID = "Alloc"
		p.Put(metric)
	}
}

func BenchmarkWithoutPool(b *testing.B) {
	var total int

	for i := 0; i < b.N; i++ {
		metric := &models.Metrics{ID: "Alloc"}
		total += len(metric.ID)
	}

	if total == 0 {
		b.Fatal("метрики не создавались")
	}
}
