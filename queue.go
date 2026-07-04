package main

import "sync"

type queue struct {
	mu       sync.Mutex
	elements []string
}

func (q *queue) enqueue(element string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.elements = append(q.elements, element)
}

func (q *queue) empty() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.elements = nil
}

func (q *queue) peekAndEmpty() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.elements) == 0 {
		return ""
	}
	element := q.elements[0]
	q.elements = nil
	return element
}

func (q *queue) get() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.elements) == 0 {
		return ""
	}
	return q.elements[len(q.elements)-1]
}

func (q *queue) length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.elements)
}
