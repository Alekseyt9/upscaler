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
		log.Fatalf("Failed to set value: %v", err)
	}

	val, err := rdb.Get(ctx, "mykey").Result()
	if err != nil {
		log.Fatalf("Failed to get value: %v", err)
	}

	fmt.Printf("Value of key 'mykey': %s\n", val)

	val2, err := rdb.Get(ctx, "nonexistent").Result()
	if err == redis.Nil {
		fmt.Println("Key 'nonexistent' does not exist")
	} else if err != nil {
		log.Fatalf("Error getting key: %v", err)
	} else {
		fmt.Printf("Value of key 'nonexistent': %s\n", val2)
	}
}
