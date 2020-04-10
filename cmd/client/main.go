package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

const testConfigFilePath = "src/redis-proxy/config/test-config.json"
const testEntriesFilePath = "src/redis-proxy/config/test-entries.json"

func main() {
	start := time.Now()

	//add entries to redis
	entries := readEntries(testEntriesFilePath)
	redisClient := newRedis()

	// populate redis
	for key, value := range entries.Entries {
		_, err := redisClient.Set(key, value, 0).Result()
		if err != nil {
			log.Panic("Failed to set values in redis with error: ", err)
		}
	}

	// get each key from redis
	var wg sync.WaitGroup
	for key, value := range entries.Entries {
		wg.Add(1)
		go func(key, value string) {
			defer wg.Done()
			response, err := http.Get(fmt.Sprintf("http://localhost:8080/%s", key))
			if err != nil {
				log.Panic("Failed to receive response with error: ", err)
			}

			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {
				log.Panic("Got a bad status code from request: ", response.StatusCode)
			}
			blob, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Panic("No body returned error: ", err)
			}

			if value != string(blob) {
				log.Panic("Got the wrong value from the server")
			}

		}(key, value)
	}
	wg.Wait()

	delta := time.Since(start)
	log.Printf("We have succesfully gotten %d keys from the proxy concurrently in: %v\n", len(entries.Entries), delta)
}

type entries struct {
	Entries map[string]string `json:"entries"`
}

func readEntries(filePath string) *entries {
	testEntries := new(entries)
	entries, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic("Unable to read test entries with error: ", err)
	}
	err = json.Unmarshal(entries, testEntries)
	if err != nil {
		log.Panic("Unable to unmarshal test entries with error: ", err)
	}
	return testEntries
}

func newRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		DB:   0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Panic("Unable to create Redis client with error: ", err)
	}
	return client
}
