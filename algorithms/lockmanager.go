package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

type Mutex struct {
	redisClients []*redis.Client
	quorum       int
	locks        map[string]bool
	mutex        sync.Mutex
}

func NewRedlock(redisAddrs []string) *Mutex {
	clients := make([]*redis.Client, len(redisAddrs))
	for i, addr := range redisAddrs {
		clients[i] = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "", // Add password if required
			DB:       0,
		})
	}

	return &Mutex{
		redisClients: clients,
		quorum:       (len(redisAddrs) / 2) + 1,
		locks:        make(map[string]bool),
	}
}

func (r *Mutex) Lock(resource string, ttl time.Duration) bool {
	startTime := time.Now()
	expiration := startTime.Add(ttl)

	for time.Now().Before(expiration) {
		locksAcquired := 0

		var wg sync.WaitGroup
		wg.Add(len(r.redisClients))

		for _, client := range r.redisClients {
			go func(c *redis.Client) {
				defer wg.Done()
				if r.acquireLock(c, resource, ttl) {
					r.mutex.Lock()
					locksAcquired++
					r.mutex.Unlock()
				}
			}(client)
		}

		wg.Wait()

		r.mutex.Lock()
		if locksAcquired >= r.quorum && r.isResourceLocked(resource) {
			r.locks[resource] = true
			r.mutex.Unlock()
			return true
		}
		r.mutex.Unlock()

		r.releaseLocks()
		time.Sleep(time.Millisecond * 10) // Sleep for a small interval before retrying
	}

	return false
}

func (r *Mutex) isResourceLocked(resource string) bool {
	for _, client := range r.redisClients {
		if client.Exists(resource).Val() == 1 {
			return true
		}
	}

	return false
}

func (r *Mutex) releaseLocks() {
	for _, client := range r.redisClients {
		client.FlushDB()
	}
}

func (r *Mutex) Unlock(resource string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.locks[resource] {
		delete(r.locks, resource)

		for _, client := range r.redisClients {
			err := client.Del(resource).Err()
			if err != nil {
				return false
			}
		}

		return true
	}

	return false
}

func (r *Mutex) acquireLock(client *redis.Client, resource string, ttl time.Duration) bool {
	return client.SetNX(resource, "locked", ttl).Val()
}

func main() {
	redisAddrs := []string{
		"localhost:6379",
		"localhost:6380",
		"localhost:6381",
	}

	redlock := NewRedlock(redisAddrs)

	resource := "my-resource"
	ttl := 10 * time.Second

	if redlock.Lock(resource, ttl) {
		defer redlock.Unlock(resource)

		// Do some critical section operations here
		fmt.Println("Lock acquired!")
	} else {
		fmt.Println("Failed to acquire lock.")
	}
}
