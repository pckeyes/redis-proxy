# redix-proxy
A simple proxy for getting cached values backed by Redis

Architectural overview:
  - Router that concurrently responds to incoming GET requests
  - App struct that contains a Config, Cache, and Redis client
    - Config holds relevant fields imported on startup (e.g. ports, cache params)
    - Cache holds a map, capacity, TTL on keys, a queue that tracks TTLs, and a mutex. The internal queue also contains a mutex
  - Code is built within a docker container

Functionality
  - Server main:
    - Creates the relevant structs and fires off a go routine that checks for expiring keys in the Cache's queue every 10 milliseconds
    - Creates a router that continually listens and serves for GET requests
    - Upon receiving a GET request, the proxy checks if the URI path exists in the cache
      - If it does, the value is returned
      - If it does not:
        - Checks if the value is in Redis
          - If it is not, a 404 status is returned
          - If it is, the value is added to the Cache. If the number of keys in the Cache exceeds the Cache capacity, the key at the head of the Cache's queue (i.e. the oldest key) gets popped off and that key is removed from the Cache's map. The new key is pushed into the queue.
  - Client main:
    - Reads in 13 test key value pairs
    - Creates a Redis client and stores the pairs within the backing Redis instance
    - Concurrently submits GET requests for each key and ensures that a 200 response was received and that the value returned is the same as the one that was submitted
    - Prints the outcome of this test
  - Key expiration routine:
    - At an interval of 10 milliseconds, the routine checks if the queue is not empty and if so, loops until all keys at the head of the queue that are expired have been popped off and removed from the map
  - Additional testing in server/main_test.go:
    - Starts up a second server on a separate port to preserve the state of the primary server
    - Tests the getting of both good and bad requests
    - Ensures that the Cache's capacity is maintained at the configured value
    - Ensures that keys are expired correctly

Algorithmic complexity of Cache operations:
  - Cache's map:
    - Getting keys that already exist in the map: constant
      - Storing a key gotten from Redis: 
        - Getting the length of the map is constant assuming the length is caches, as it is with slices
        - Popping the oldest key from the queue is constant due to head pointer
        - Deleting the oldest key from the map is constant
        - Setting the key in the map is constant
        - Pushing the new key into the queue is constant due to tail pointer
  - Cache's expiration checker routine:
     - Upon each loop, if an expired key is found, the complexity will be O(number expired keyes)

Running the proxy and test:
   - Pull this repo
   - Within redis-proxy, run make test
   - Client test output will be printed
   - Extra integration tests will only print if they fail
   
Time allocation:
  - Spent some personal time familiarizing myself with Docker and Go before beginning
  - Three hours spent on building the proxy
  - One hour spent writing tests

Functionality not implemented:
  - In its current state, the proxy assumes that all values stored in the Cache will be strings. This is not ideal, and with more time I would have used the interface{} type to allow for more flexible usage.
      
