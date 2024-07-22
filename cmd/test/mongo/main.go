package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("imageDB").Collection("previews")

	// Вставка документов для примера
	imageIDs := []string{"image1", "image2", "image3"}
	previews := [][]byte{
		[]byte("image data 1"),
		[]byte("image data 2"),
		[]byte("image data 3"),
	}

	for i, imageID := range imageIDs {
		document := bson.M{
			"image_id":  imageID,
			"preview":   previews[i],
			"createdAt": time.Now(),
		}
		_, err = collection.InsertOne(context.Background(), document)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Чтение документов батчем
	filter := bson.M{
		"image_id": bson.M{
			"$in": imageIDs,
		},
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var result struct {
			ImageID string `bson:"image_id"`
			Preview []byte `bson:"preview"`
		}
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		// Сохранение извлеченных изображений в файлы
		err = os.WriteFile("retrieved_"+result.ImageID+".jpg", result.Preview, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}
}
