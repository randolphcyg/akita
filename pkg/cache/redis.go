package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

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

// // Set 存储值
// func Set(key string, value interface{}) error {

// 	serialized, err := serializer(value)
// 	if err != nil {
// 		return err
// 	}
// 	err = RedisClient.Set(ctx, key, serialized, 0).Err()

// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // Get 取值
// func Get(key string) (interface{}, bool) {
// 	vl := RedisClient.Get(ctx, key)
// 	// bytesVl, err := vl.Bytes()

// 	// if err != nil {
// 	// 	return nil, false
// 	// }
// 	// finalValue, err := deserializer(bytesVl)
// 	// if err != nil {
// 	// 	return nil, false
// 	// }

// 	return vl, true

// }

// type item struct {
// 	Value interface{}
// }

// func ser2(value interface{}) {
// 	//将数据进行gob序列化
// 	var buffer bytes.Buffer
// 	ecoder := gob.NewEncoder(&buffer)

// 	storeValue := item{
// 		Value: value,
// 	}
// 	ecoder.Encode(storeValue)
// 	//reids缓存数据
// 	RedisClient.Set(ctx, "HrUsers", buffer.Bytes(), 0)

// }

// func deser2(key string) (interface{}, error) {
// 	//redis读取缓存
// 	strCmd := RedisClient.Get(ctx, key)
// 	// rebytes, _ := redis.Bytes(conn.Do("get", "struct2"))
// 	//进行gob序列化
// 	rrr, _ := strCmd.Bytes()
// 	reader := bytes.NewReader(rrr)
// 	dec := gob.NewDecoder(reader)
// 	// object := &HrUser{}
// 	// fmt.Printf("%T", object)
// 	var res item
// 	dec.Decode(&res)
// 	// fmt.Println(object)
// 	return res.Value, nil
// }

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
