package config

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

// Config a struct holding all server configuration information
type Config struct {
	RedisConfig   RedisConfig   `json:"redis_config"`
	CacheExpiry   time.Duration `json:"expiry"` //in seconds
	CacheCapacity int           `json:"capacity"`
	Port          int           `json:"port"`
}

// RedisConfig a struct holding Redis connection information
type RedisConfig struct {
	Address  string `json:"address"`
	Database int    `json:"db"`
}

// ReadConfig reads in the provided config file and marshals it into a Config struct
// note that expiry is stored as seconds
func ReadConfig(filePath string) (*Config, error) {
	config := new(Config)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	config.CacheExpiry = time.Duration(config.CacheExpiry * time.Second)
	return config, nil
}
