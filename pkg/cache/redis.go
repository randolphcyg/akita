package cache

import (
	"context"
	"encoding/json"
	"time"

	"gitee.com/RandolphCYG/akita/pkg/log"
	"github.com/go-redis/redis/v8"
)

/*
TODO:序列化与反序列化需要优化并测试其他类型的数据
需要增加删除和修改封装，暂时没需要
*/

// 最新版本的redis需要传上下文参数
var ctx = context.Background()

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
)

// 初始化连接
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

// func serializer(value interface{}) ([]byte, error) {
// 	var buffer bytes.Buffer
// 	enc := gob.NewEncoder(&buffer)
// 	storeValue := item{
// 		Value: value,
// 	}
// 	err := enc.Encode(storeValue)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buffer.Bytes(), nil
// }

// Serializer 序列化 将go结构体对象转换为字节流
func Serializer(value interface{}) (result []byte, err error) {
	result, err = json.Marshal(value)
	return
}

// func deserializer(value []byte) (interface{}, error) {
// 	var res item
// 	buffer := bytes.NewReader(value)
// 	dec := gob.NewDecoder(buffer)
// 	err := dec.Decode(&res)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return res.Value, nil
// }

// Deserializer 反序列化 将字节流转换为go结构体对象
func Deserializer(value []byte, result interface{}) (interface{}, error) {
	err := json.Unmarshal(value, &result)
	if err != nil {
		log.Log().Error("get data failed, err:%v\n", err)
		return nil, err
	}
	return result, nil
}

// Set 覆盖更新缓存中的键值
func Set(key string, value interface{}) (err error) {
	err = RedisClient.Set(ctx, key, value, 0).Err()
	if err != nil {
		log.Log().Error("cache data failed, err:%v\n", err)
		return
	}
	return
}

// Get 从缓存取数据
func Get(key string, value interface{}) (result interface{}, err error) {
	strCmd := RedisClient.Get(ctx, key)
	byteValue, _ := strCmd.Bytes()
	result, err = Serializer(byteValue)
	if err != nil {
		log.Log().Error("deserializer data failed, err:%v\n", err)
		return
	}
	return
}
