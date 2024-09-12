package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	err := rdb.Set(ctx, "mykey", "hello, redis!", 0).Err()
	if err != nil {
		log.Fatalf("Не удалось установить значение: %v", err)
	}

	val, err := rdb.Get(ctx, "mykey").Result()
	if err != nil {
		log.Fatalf("Не удалось получить значение: %v", err)
	}

	fmt.Printf("Значение ключа 'mykey': %s\n", val)

	val2, err := rdb.Get(ctx, "nonexistent").Result()
	if err == redis.Nil {
		fmt.Println("Ключ 'nonexistent' не существует")
	} else if err != nil {
		log.Fatalf("Ошибка при получении ключа: %v", err)
	} else {
		fmt.Printf("Значение ключа 'nonexistent': %s\n", val2)
	}
}
