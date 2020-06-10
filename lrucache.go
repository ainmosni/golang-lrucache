/*
Copyright 2020 Daniël Franke

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package lrucache contains a simple LRU cache, designed to wrap a function.
package lrucache

import (
	"container/list"
	"log"
	"sync"
)

// CachedFunc represents a function that we want to cache the responses from.
type CachedFunc = func(a uint64) uint64

// Entry is a list Entry, we're adding both key and value here so that we can easily remove it from
// the lookup table.
type Entry struct {
	Key   uint64
	Value uint64
}

// Cache represents an instance of an LRU cache.
type Cache struct {
	function        CachedFunc
	ResponsesList   *list.List
	ResponsesLookup map[uint64]*list.Element
	size            int
	sync.Mutex
}

// New returns a new Cache instance.
func New(size int, f CachedFunc) Cache {
	return Cache{
		function:        f,
		ResponsesList:   list.New(),
		ResponsesLookup: make(map[uint64]*list.Element),
		size:            size,
	}
}

// Call checks the lookup table to see if there is already a cached response for this input.
// If there already is, move the cache entry to the front of the list.
// If not, check if our cache is full. If full, delete the least recently used entry (at the end of the list)
// and insert the new one in the front.
func (c *Cache) Call(a uint64) uint64 {
	el, ok := c.ResponsesLookup[a]
	if ok {
		log.Println("Cache hit")
		c.Lock()
		c.ResponsesList.MoveToFront(el)
		c.Unlock()
		return el.Value.(*Entry).Value
	}

	log.Println("Cache miss")

	log.Println("Calling function")
	resp := c.function(a)

	if c.ResponsesList.Len() >= c.size {
		log.Println("Cache is full, evicting least recently used.")
		el := c.ResponsesList.Back()
		if el != nil {
			c.Lock()
			delete(c.ResponsesLookup, el.Value.(*Entry).Key)
			c.ResponsesList.Remove(el)
			c.Unlock()
		}
	}

	log.Println("Adding entry to cache.")
	e := &Entry{
		Key:   a,
		Value: resp,
	}
	c.Lock()
	el = c.ResponsesList.PushFront(e)
	c.ResponsesLookup[e.Key] = el
	c.Unlock()
	return resp
}
