package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

const pathKey = "key"

// GetValue a handler for getting a value from redis
func (a *App) GetValue(w http.ResponseWriter, r *http.Request) {

	//get key from path
	key, exists := mux.Vars(r)[pathKey]
	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//get value from cache or redis
	value, err := getValue(a, key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}
	w.Write([]byte(value))
}

func getValue(a *App, key string) (string, error) {
	value, exists := a.Cache.Get(key)
	if !exists {
		//try to get value from redis
		var err error
		value, err = a.client.Get(key).Result()
		if err != nil {
			return "", err
		}
		//put it in the cache
		a.Cache.Put(key, value)
	}
	return value, nil
}
