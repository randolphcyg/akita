package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// Config redis 配置
type Config struct {
	Addr         string
	Password     string
	DB           int
	MinIdleConn  int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	PoolTimeout  time.Duration
}

var (
	RedisClient *redis.Client
	ctx         = context.Background() // 最新版本的redis需要传上下文参数
)

// Init 初始化redis连接池
func Init(c *Config) (err error) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         c.Addr,
		Password:     c.Password,
		DB:           c.DB,
		MinIdleConns: c.MinIdleConn,
		DialTimeout:  c.DialTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		PoolSize:     c.PoolSize,
		PoolTimeout:  c.PoolTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RedisClient.Ping(ctx).Result()
	return
}

// Serializer 序列化 将go结构体对象转换为字节流
func Serializer(value interface{}) (res []byte, err error) {
	res, err = json.Marshal(value)
	return
}

// Deserializer 反序列化 将字节流转换为go结构体对象
func Deserializer(value []byte, res interface{}) (interface{}, error) {
	err := json.Unmarshal(value, &res)
	if err != nil {
		err = errors.New("Fail to get deserializer data, err: " + err.Error())
		return nil, err
	}
	return res, nil
}

// Set 存string
func Set(key string, value interface{}) (err error) {
	err = RedisClient.Set(ctx, key, value, 0).Err()
	if err != nil {
		err = errors.New("Fail to cache data, err: " + err.Error())
		return
	}
	return
}

// Get 取string
func Get(key string) (res interface{}, err error) {
	strCmd := RedisClient.Get(ctx, key)
	byteValue, _ := strCmd.Bytes()
	res, err = Serializer(byteValue)
	if err != nil {
		err = errors.New("Fail to deserializer data, err: " + err.Error())
		return
	}
	return
}

/*
以下是对hash操作的封装 将上下文参数隐藏 错误上抛
*/

// HSet hash 存储
func HSet(key string, field string, value interface{}) (res bool, err error) {
	res, err = RedisClient.HMSet(ctx, key, map[string]interface{}{field: value}).Result()
	if err != nil {
		err = errors.New("Fail to batch set a filed, err: " + err.Error())
		return
	}
	return
}

// HMSet hash 批量存储
func HMSet(key string, data map[string]interface{}) (res bool, err error) {
	res, err = RedisClient.HMSet(ctx, key, data).Result()
	if err != nil {
		err = errors.New("Fail to batch set fileds, err: " + err.Error())
		return
	}
	return
}

// HDel hash 删除
func HDel(key string) (res int64, err error) {
	res, err = RedisClient.Del(ctx, key).Result()
	if err != nil {
		err = errors.New("Fail to delete key, err: " + err.Error())
		return
	}
	return
}

// HGet hash 获取某个元素
func HGet(key string, field string) (res string, err error) {
	res, _ = RedisClient.HGet(ctx, key, field).Result()
	return
}

// HGetAll hash 获取全部元素
func HGetAll(key string) (data map[string]string, err error) {
	data, err = RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		err = errors.New("Fail to get all elements, err: " + err.Error())
		return
	}
	return
}

// HExists 判断哈希表key是否存在
func HExists(key string, field string) (res bool, err error) {
	res, err = RedisClient.HExists(ctx, key, field).Result()
	if err != nil {
		err = errors.New("Fail to determine whether the element exists in hash table, err: " + err.Error())
		return
	}
	return
}

// HLen hash 获取长度
func HLen(key string) (res int64, err error) {
	res, err = RedisClient.HLen(ctx, key).Result()
	if err != nil {
		err = errors.New("Fail to get length of hash table, err: " + err.Error())
		return
	}
	return
}

// Exists 判断缓存项是否存在
func Exists(key string) (res bool, err error) {
	result, err := RedisClient.Exists(ctx, key).Result()
	if err != nil {
		err = errors.New("Fail to determine whether the element exists, err: " + err.Error())
		return
	}
	if result > 0 {
		return true, nil
	} else {
		return false, nil
	}
}
