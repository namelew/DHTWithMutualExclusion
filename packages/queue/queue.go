package queue

import (
	"sync"
)

type Queue[T any] struct {
	slice []T
	lock  sync.Mutex
	cond  sync.Cond
}

func New[T any]() *Queue[T] {
	q := Queue[T]{
		slice: make([]T, 0),
	}

	q.cond = *sync.NewCond(&q.lock)

	return &q
}

func (q *Queue[T]) Enqueue(i T) {
	q.lock.Lock()
	q.slice = append(q.slice, i)
	q.lock.Unlock()
	q.cond.Signal()
}

func (q *Queue[T]) Dequeue() T {
	for q.Empty() {
		q.cond.Wait()
	}
	q.lock.Lock()
	defer q.lock.Unlock()
	result := q.slice[0]
	q.slice = q.slice[1:len(q.slice)]
	return result
}

func (q *Queue[T]) Empty() bool {
	return len(q.slice) == 0
}
