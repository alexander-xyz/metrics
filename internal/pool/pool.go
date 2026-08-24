// Package pool содержит пул переиспользуемых объектов с обязательным
// сбросом состояния при возврате.
package pool

import "sync"

// Resettable ограничивает параметр пула указателями на тип T,
// у которых есть метод Reset.
type Resettable[T any] interface {
	*T
	Reset()
}

// Pool хранит объекты одного типа и выдаёт их повторно, избавляя
// от повторных выделений памяти. Перед возвратом в пул состояние
// объекта сбрасывается методом Reset.
//
// Пул безопасен для конкурентного использования.
type Pool[T any, P Resettable[T]] struct {
	pool sync.Pool
}

// New создаёт пул объектов типа T. Объекты, которых не хватает,
// пул создаёт сам.
//
// Тип указывается двумя параметрами — значением и указателем на него:
//
//	p := pool.New[models.Metrics, *models.Metrics]()
func New[T any, P Resettable[T]]() *Pool[T, P] {
	return &Pool[T, P]{
		pool: sync.Pool{
			New: func() any {
				return P(new(T))
			},
		},
	}
}

// Get возвращает объект из пула. Если свободных объектов нет,
// создаётся новый.
func (p *Pool[T, P]) Get() P {
	return p.pool.Get().(P)
}

// Put сбрасывает состояние объекта и возвращает его в пул.
// Значение nil игнорируется.
func (p *Pool[T, P]) Put(value P) {
	if value == nil {
		return
	}

	value.Reset()
	p.pool.Put(value)
}
