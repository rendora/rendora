/*
Copyright 2018 George Badawi.
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

package service

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/go-redis/redis"

	cache "github.com/patrickmn/go-cache"
)

//Store represents the cache store
type Store struct {
	DefaultTimeout time.Duration
	Type           uint8
	redis          *redis.Client
	gocache        *cache.Cache
}

//StoreConfig cache config
type StoreConfig struct {
	Type    string
	Timeout uint32
	Redis   struct {
		Address   string
		Password  string
		DB        int
		KeyPrefix string
	}
}

const (
	typeLocal = 0
	typeRedis = 1
	typeNone  = 2
)

//InitCacheStore initializes the cache store
func InitCacheStore(cfg *StoreConfig) *Store {
	cs := &Store{
		DefaultTimeout: time.Duration(cfg.Timeout) * time.Second,
	}

	switch cfg.Type {
	case "local":
		cs.Type = typeLocal
		cs.gocache = cache.New(cs.DefaultTimeout, 4*time.Minute)
	case "redis":
		cs.Type = typeRedis
		cs.redis = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Address,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
	case "none":
		cs.Type = typeNone
	default:
		cs.Type = typeLocal
	}

	return cs
}

//Set stores HeadlessResponse in the cache with the key cKey (i.e. request path)
func (c *Store) Set(cKey string, d *HeadlessResponse) error {

	switch c.Type {
	case typeLocal:
		c.gocache.Set(cKey, d, cache.DefaultExpiration)
	case typeRedis:
		op := &bytes.Buffer{}
		enc := json.NewEncoder(op)
		enc.SetEscapeHTML(false)
		err := enc.Encode(d)
		if err != nil {
			return err
		}
		err = c.redis.SetNX(cKey, string(op.Bytes()), c.DefaultTimeout).Err()
		if err != nil {
			return err
		}
	case typeNone:
	default:
	}
	return nil
}

//Get gets HeadlessResponse from the cache with the key cKey (i.e. request path)
func (c *Store) Get(cKey string) (*HeadlessResponse, bool, error) {

	switch c.Type {
	case typeLocal:
		if x, found := c.gocache.Get(cKey); found {
			foo := x.(*HeadlessResponse)

			return foo, true, nil
		}
		return nil, false, nil
	case typeRedis:
		val, err := c.redis.Get(cKey).Result()
		if err == redis.Nil {
			return nil, false, nil
		} else if err != nil {
			return nil, false, err
		} else {
			var dt HeadlessResponse
			json.Unmarshal([]byte(val), &dt)

			return &dt, true, nil
		}
	case typeNone:
		return nil, false, nil
	default:
		return nil, false, nil
	}
}
