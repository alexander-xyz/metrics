package audit

import "sync"

type Event struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Observer interface {
	Update(event Event)
	ID() string
}

type Publisher struct {
	mu        sync.RWMutex
	observers map[string]Observer
}

func NewPublisher() *Publisher {
	return &Publisher{observers: make(map[string]Observer)}
}

func (p *Publisher) Register(observer Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.observers[observer.ID()] = observer
}

func (p *Publisher) Deregister(observer Observer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.observers, observer.ID())
}

func (p *Publisher) Notify(event Event) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, observer := range p.observers {
		observer.Update(event)
	}
}
