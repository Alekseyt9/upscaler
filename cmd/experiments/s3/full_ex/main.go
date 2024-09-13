package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"

	//"github.com/aws/aws-sdk-go-v2/aws/credentials"
	cfg "github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {

	c, err := cfg.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Замените на ваши значения
	accessKey := c.S3AccessKeyID
	secretKey := c.S3SecretAccessKey
	bucketName := c.S3BucketName
	region := "ru-central1"
	endpoint := "https://storage.yandexcloud.net"

	// Создаем конфигурацию AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           endpoint,
					SigningRegion: region,
				}, nil
			})),
	)
	if err != nil {
		panic("Ошибка загрузки конфигурации: " + err.Error())
	}

	// Создаем клиента S3
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Генерируем Presigned URL для загрузки файла
	presignClient := s3.NewPresignClient(client)

	// Укажите ключ объекта (имя файла в бакете)
	objectKey := "uploaded-file.jpg"

	// Генерируем Presigned URL
	presignedURL, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		// ContentType: aws.String("image/jpeg"), // Если требуется указать Content-Type
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		panic("Ошибка генерации Presigned URL: " + err.Error())
	}

	fmt.Println("Сгенерированный Presigned URL для загрузки файла:")
	fmt.Println(presignedURL.URL)

	// Загружаем файл с использованием сгенерированного URL
	err = uploadFileWithPresignedURL(presignedURL.URL, "input.jpg")
	if err != nil {
		panic("Ошибка загрузки файла: " + err.Error())
	}

	fmt.Println("Файл успешно загружен!")
}

// Функция для загрузки файла с использованием Presigned URL
func uploadFileWithPresignedURL(url string, filePath string) error {
	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл: %v", err)
	}
	defer file.Close()

	// Создаем HTTP-запрос PUT
	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return fmt.Errorf("не удалось создать HTTP-запрос: %v", err)
	}

	// Если вы указали ContentType при генерации Presigned URL, раскомментируйте следующую строку
	// req.Header.Set("Content-Type", "image/jpeg")

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения HTTP-запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("неудачный ответ сервера: %s\nТело ответа: %s", resp.Status, string(bodyBytes))
	}

	return nil
}
