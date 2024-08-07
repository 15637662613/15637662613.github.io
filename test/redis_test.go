package test

import (
	"context"
	"fmt"
	"gin-gorm-OJ/models"
	"github.com/go-redis/redis/v8"
	"testing"
	"time"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // no password set
	DB:       0,  // use default DB
})

func TestRedisSet(t *testing.T) {
	err := rdb.Set(ctx, "name", "mmc", time.Second*10).Err()
	if err != nil {
		panic(err)
	}
}

func TestRedisGet(t *testing.T) {
	val, err := rdb.Get(ctx, "name").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
}

// 测试封装方法
func TestRedisGetByModels(t *testing.T) {
	val, err := models.RDB.Get(ctx, "name").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
}
