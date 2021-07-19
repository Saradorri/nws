package rd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"log"
	"nws/config"
	"time"
)

type Info struct {
	Client    *websocket.Conn
	Queue     string
	ClientID  string
	MessageID string
	Message   string
}

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func (rdc *RedisClient) connect() *redis.Client {
	rdc.ctx = context.Background()

	if rdc.client != nil {
		if _, e := rdc.client.Ping(rdc.ctx).Result(); e == nil {
			return rdc.client
		}
	}
	rdc.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.CNF.Redis.Host, config.CNF.Redis.Port),
		Password: config.CNF.Redis.Password,
		DB:       config.CNF.Redis.Name,
	})

	return rdc.client
}

func (rdc *RedisClient) Set(key string, value interface{}) {
	rdb := rdc.connect()

	v, _ := json.Marshal(value)

	err := rdb.Set(rdc.ctx, key, v, 0).Err()

	if err != nil {
		fmt.Printf("redis set data error: %s", err)
		time.Sleep(3 * time.Second)
		rdc.Set(key, value)
	}
}

func (rdc *RedisClient) Get(key string) *Info {
	rdb := rdc.connect()
	r, err := rdb.Get(rdc.ctx, key).Bytes()

	var dest *Info
	if !(err == redis.Nil || err != nil) {
		_ = json.Unmarshal(r, &dest)
		return dest
	}
	return nil
}

func (rdc *RedisClient) Delete(key string) {
	rdb := rdc.connect()
	rdb.Del(rdc.ctx, key)
}

func (rdc *RedisClient) GetAllKeys() []interface{} {

	rdb := rdc.connect()
	data, err := rdb.Do(rdc.ctx, "KEYS", "*").Result()
	if err != nil {
		log.Println(err)
	}
	return data.([]interface{})
}
