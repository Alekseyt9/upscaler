package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func main() {
	// Подключение к Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Адрес Redis сервера
		Password: "",               // Пароль, если он установлен (по умолчанию "")
		DB:       0,                // Используемая база данных (по умолчанию 0)
	})

	// Установить значение ключа
	err := rdb.Set(ctx, "mykey", "hello, redis!", 0).Err()
	if err != nil {
		log.Fatalf("Не удалось установить значение: %v", err)
	}

	// Получить значение ключа
	val, err := rdb.Get(ctx, "mykey").Result()
	if err != nil {
		log.Fatalf("Не удалось получить значение: %v", err)
	}

	fmt.Printf("Значение ключа 'mykey': %s\n", val)

	// Попытка получить несуществующий ключ
	val2, err := rdb.Get(ctx, "nonexistent").Result()
	if err == redis.Nil {
		fmt.Println("Ключ 'nonexistent' не существует")
	} else if err != nil {
		log.Fatalf("Ошибка при получении ключа: %v", err)
	} else {
		fmt.Printf("Значение ключа 'nonexistent': %s\n", val2)
	}
}
