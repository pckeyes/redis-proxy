package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"redis-proxy/pkg/config"
	"redis-proxy/pkg/server"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const testConfigFilePath = "../../config/test-config.json"
const testEntriesFilePath = "../../config/test-entries.json"

var testRedisClient *redis.Client
var testConfig *config.Config
var testApp *server.App
var testEntries *entries
var router *mux.Router

type entries struct {
	Entries map[string]string `json:"entries"`
}

func readTestEntries(filePath string) *entries {
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

func newConfig() *config.Config {
	config, err := config.ReadConfig(testConfigFilePath)
	if err != nil {
		log.Panic("Unable to read config file with error: ", err)
	}
	return config
}

func newRedis(config *config.Config) *redis.Client {
	redisConfig := testConfig.RedisConfig
	client := redis.NewClient(&redis.Options{
		Addr: redisConfig.Address,
		DB:   redisConfig.Database,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Panic("Unable to create Redis client with error: ", err)
	}
	return client
}

func newApp(config *config.Config) *server.App {
	app, err := server.NewApp(config)
	if err != nil {
		log.Panic("Unable to instantiate new App with error: ", err)
	}
	return app
}

func startServer() {
	r := mux.NewRouter()
	r.HandleFunc("/{key}", testApp.GetValue).Methods("GET")

	go testApp.Cache.ExpiryChecker(context.Background())

	port := fmt.Sprintf(":%d", testConfig.Port)
	go func() {
		if err := http.ListenAndServe(port, r); err != nil {
			log.Fatal(err)
		}
	}()
	router = r
}

func initTestEnv() {
	testEntries = readTestEntries(testEntriesFilePath)
	testConfig = newConfig()
	testRedisClient = newRedis(testConfig)
	testApp = newApp(testConfig)
	startServer()
}

func cleanUp() {
	testRedisClient.FlushDB()
	testApp.Cache.Clear()
}

func TestMain(m *testing.M) {
	//init global vars
	initTestEnv()

	exitValue := m.Run()

	//clean up
	cleanUp()

	os.Exit(exitValue)
}

// TestGetValue tests that a value can be retrived from the proxy
// and that requests for keys that aren't in the DB return a NotFound response code
func TestGetValue(t *testing.T) {
	key := "hello"
	value := "world"
	err := set(key, value)
	if err != nil {
		log.Panic("Failed to set values in redis with error: ", err)
	}

	// test good request
	goodRequest(key, t)

	// test bad request
	badRequest("badKey", t)

	cleanUp()
}

// TestValidReturnValue tests that the value that is returned
// by the proxy is the correct value
func TestValidReturnedValue(t *testing.T) {
	key := "hello"
	value := "world"
	err := set(key, value)
	if err != nil {
		log.Panic("Failed to set values in redis with error: ", err)
	}

	// test good request
	returnedValue := goodRequest(key, t)
	assert.Equal(t, value, returnedValue, fmt.Sprintf("Expected value is %s", value))

	cleanUp()
}

// TestCacheCapacity tests that the cache capacity is maintained
// across more get requests than the cache can hold. Furthermore,
// the concurrent requests indicate that multiple clients could call
// this api and be successfully served results
func TestCacheCapacity(t *testing.T) {
	// populate redis
	for key, value := range testEntries.Entries {
		err := set(key, value)
		if err != nil {
			log.Panic("Failed to set values in redis with error: ", err)
		}
	}

	// get each key from redis
	// note the assumption that the gets will complete
	// before any of the keys have expired
	submitConcurrentRequests(testEntries.Entries, t)

	assert.Equal(t, testConfig.CacheCapacity, testApp.Cache.Size(), fmt.Sprintf("Capacity of %d is expected", testConfig.CacheCapacity))

	cleanUp()
}

// TestCacheExpiry tests that cached keys will expire
// after the configured duration
// note that if the time it takes for the expired keys to be removed from the cache
// is longer than the expiry parameter, this test will fail
func TestCacheExpiry(t *testing.T) {
	// populate redis
	for key, value := range testEntries.Entries {
		err := set(key, value)
		if err != nil {
			log.Panic("Failed to set values in redis with error: ", err)
		}
	}

	// get each key from redis
	submitConcurrentRequests(testEntries.Entries, t)

	// the additional buffer affed to time.Sleep() is a hacky fix to deal with race conditions
	buffer := time.Millisecond * 10
	time.Sleep(testConfig.CacheExpiry + buffer)
	assert.Equal(t, 0, testApp.Cache.Size(), "Cache capacity of 0 is expected")
	cleanUp()
}

func set(key string, value string) error {
	_, err := testRedisClient.Set(key, value, 0).Result()
	return err
}

func newRequest(key string) *http.Request {
	request, _ := http.NewRequest("GET", fmt.Sprintf("/%s", key), nil)
	return request
}

func goodRequest(key string, t *testing.T) string {
	request := newRequest(key)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "OK response is expected")
	return response.Body.String()
}

func badRequest(key string, t *testing.T) {
	request := newRequest(key)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, 404, response.Code, "NotFound response is expected")
}

func submitConcurrentRequests(entries map[string]string, t *testing.T) {
	var wg sync.WaitGroup
	for key := range entries {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			request, _ := http.NewRequest("GET", fmt.Sprintf("/%s", key), nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			assert.Equal(t, 200, response.Code, "OK response is expected")
		}(key)
	}
	wg.Wait()
}
