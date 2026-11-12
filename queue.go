package main

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type Item struct {
	link      string
	movies    []string
	expiresAt time.Time
	index     int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].expiresAt.Before(pq[j].expiresAt) // earliest expires first.
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)

}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

type Cache struct {
	items   map[string]*Item
	queue   PriorityQueue
	maxSize int
	mu      sync.Mutex
}

func NewCache(maxSize int) *Cache {
	return &Cache{
		items:   make(map[string]*Item),
		queue:   make(PriorityQueue, 0),
		maxSize: maxSize,
	}
}

func (c *Cache) Add(link string, movies []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Now().Add(1 * time.Hour)
	if item, exists := c.items[link]; exists {
		item.movies = movies
		item.expiresAt = expiresAt
		heap.Fix(&c.queue, item.index)
	} else {
		if len(c.items) >= c.maxSize {
			oldest := heap.Pop(&c.queue).(*Item)
			delete(c.items, oldest.link)
		}
		item = &Item{
			link:      link,
			movies:    movies,
			expiresAt: expiresAt,
		}
		c.items[link] = item
		heap.Push(&c.queue, item)
	}
}

func (c *Cache) Get(link string) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[link]; exists {
		fmt.Println("Link,", item.link)
		if time.Now().Before(item.expiresAt) {
			return item.movies, true
		}

		// item expired, remove
		delete(c.items, link)
		heap.Remove(&c.queue, item.index)
	}
	return nil, false
}
