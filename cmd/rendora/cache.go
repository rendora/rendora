package main

import (
	"bytes"
	"time"

	"github.com/go-redis/redis"

	cache "github.com/patrickmn/go-cache"
)

//CacheStore represents the cache store
type CacheStore struct {
	DefaultTimeout time.Duration
	Type           uint8
	redis          *redis.Client
	gocache        *cache.Cache
}

const (
	typeLocal = 0
	typeRedis = 1
	typeNone  = 2
)

//InitCacheStore initializes the cache store
func InitCacheStore() *CacheStore {
	cs := &CacheStore{
		DefaultTimeout: time.Duration(Rendora.C.Cache.Timeout) * time.Second,
	}

	switch Rendora.C.Cache.Type {
	case "local":
		cs.Type = typeLocal
		cs.gocache = cache.New(cs.DefaultTimeout, 4*time.Minute)
	case "redis":
		cs.Type = typeRedis
		cs.redis = redis.NewClient(&redis.Options{
			Addr:     Rendora.C.Cache.Redis.Address,
			Password: Rendora.C.Cache.Redis.Password,
			DB:       Rendora.C.Cache.Redis.DB,
		})
	case "none":
		cs.Type = typeNone
	default:
		cs.Type = typeLocal
	}

	return cs
}

//Set stores HeadlessResponse in the cache with the key cKey (i.e. request path)
func (c *CacheStore) Set(cKey string, d *HeadlessResponse) error {

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
func (c *CacheStore) Get(cKey string) (*HeadlessResponse, bool, error) {

	switch c.Type {
	case typeLocal:
		if x, found := c.gocache.Get(cKey); found {
			foo := x.(*HeadlessResponse)
			if Rendora.C.Server.Enable {
				Rendora.M.CountSSRCached.Inc()
			}
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
			if Rendora.C.Server.Enable {
				Rendora.M.CountSSRCached.Inc()
			}
			return &dt, true, nil
		}
	case typeNone:
		return nil, false, nil
	default:
		return nil, false, nil
	}
}
