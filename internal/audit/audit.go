// Package audit реализует аудит запросов по паттерну «Наблюдатель»:
// издатель рассылает события аудита всем подписанным приёмникам.
package audit

import "sync"

// Event — событие аудита обработанного запроса.
type Event struct {
	IPAddress string   `json:"ip_address"`
	Metrics   []string `json:"metrics"`
	Timestamp int64    `json:"ts"`
}

// Observer — приёмник событий аудита.
type Observer interface {
	Update(event Event)
	ID() string
}

// Publisher рассылает события аудита подписанным приёмникам
// и безопасен для конкурентного использования.
type Publisher struct {
	observers map[string]Observer
	mu        sync.RWMutex
}

// NewPublisher создаёт издателя без подписчиков.
func NewPublisher() *Publisher {
	return &Publisher{observers: make(map[string]Observer)}
}

// Register подписывает приёмник на события аудита.
func (p *Publisher) Register(observer Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.observers[observer.ID()] = observer
}

// Deregister отписывает приёмник от событий аудита.
func (p *Publisher) Deregister(observer Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.observers, observer.ID())
}

// Notify рассылает событие всем подписанным приёмникам.
func (p *Publisher) Notify(event Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, observer := range p.observers {
		observer.Update(event)
	}
}
