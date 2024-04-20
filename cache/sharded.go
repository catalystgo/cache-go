package cache

import (
	"errors"
	"fmt"
	"sync"

	"github.com/serialx/hashring"
)

var (
	// ErrEmptyNodes is the TTL error if it is less than 0.
	ErrEmptyNodes = errors.New("empty nodes list for sharded cache client")
)

type shardedCache struct {
	ring *hashring.HashRing

	nodes      []NamedCache
	nodeByName map[string]NamedCache
}

func NewShardedCache(nodes ...NamedCache) (Cache, error) {
	if len(nodes) == 0 {
		return nil, ErrEmptyNodes
	}

	var (
		names      = make([]string, len(nodes))
		nodeByName = make(map[string]NamedCache, len(nodes))
	)

	for i, node := range nodes {
		names[i] = node.Name()
		nodeByName[node.Name()] = node
	}

	ring := hashring.New(names)

	return &shardedCache{
		ring: ring,

		nodes:      nodes,
		nodeByName: nodeByName,
	}, nil
}

func (sc *shardedCache) Cap() int {
	return sc.doInParallel(func(nc NamedCache) int { return nc.Cap() })
}

func (sc *shardedCache) Len() int {
	return sc.doInParallel(func(nc NamedCache) int { return nc.Len() })
}

func (sc *shardedCache) Clear() {
	_ = sc.doInParallel(func(nc NamedCache) int { nc.Clear(); return 0 })
}

func (sc *shardedCache) Contains(key interface{}) bool {
	node, ok := sc.keyToNode(key)
	if !ok {
		return false
	}
	return node.Contains(key)
}

func (sc *shardedCache) Get(key interface{}) (value interface{}, ok bool) {
	node, ok := sc.keyToNode(key)
	if !ok {
		return nil, false
	}
	return node.Get(key)
}

func (sc *shardedCache) Peek(key interface{}) (value interface{}, ok bool) {
	node, ok := sc.keyToNode(key)
	if !ok {
		return nil, false
	}
	return node.Peek(key)
}

func (sc *shardedCache) Put(key, value interface{}) {
	node, ok := sc.keyToNode(key)
	if !ok {
		return
	}
	node.Put(key, value)
}

func (sc *shardedCache) Remove(key interface{}) {
	node, ok := sc.keyToNode(key)
	if !ok {
		return
	}
	node.Remove(key)
}

func (sc *shardedCache) keyToNode(key interface{}) (Cache, bool) {
	name, ok := sc.ring.GetNode(fmt.Sprintf("%+v", key))
	if !ok {
		return nil, false
	}
	return sc.nodeByName[name], true
}

func (sc *shardedCache) doInParallel(f func(NamedCache) int) int {
	output := make(chan int)

	wg := sync.WaitGroup{}
	for _, node := range sc.nodes {
		wg.Add(1)
		go func(node NamedCache) {
			defer wg.Done()
			output <- f(node)
		}(node)
	}

	wg.Wait()
	close(output)

	var sum int
	for v := range output {
		sum += v
	}

	return sum
}
