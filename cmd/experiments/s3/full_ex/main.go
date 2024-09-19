package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	cfg "github.com/Alekseyt9/upscaler/internal/back/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	c, err := cfg.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	accessKey := c.S3AccessKeyID
	secretKey := c.S3SecretAccessKey
	bucketName := c.S3BucketName
	region := "ru-central1"
	endpoint := "https://storage.yandexcloud.net"

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
		panic("Error loading configuration: " + err.Error())
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(client)
	objectKey := "uploaded-file.jpg"

	presignedURL, err := presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		panic("Error generating presigned URL for upload: " + err.Error())
	}

	fmt.Println("Generated presigned URL for file upload:")
	fmt.Println(presignedURL.URL)

	err = uploadFileWithPresignedURL(presignedURL.URL, "input.jpg")
	if err != nil {
		panic("Error uploading file: " + err.Error())
	}
	fmt.Println("File uploaded successfully!")

	downloadURL, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		panic("Error generating presigned URL for download: " + err.Error())
	}

	fmt.Println("Generated presigned URL for file download:")
	fmt.Println(downloadURL.URL)

	err = downloadFileWithPresignedURL(downloadURL.URL, "downloaded-output.jpg")
	if err != nil {
		panic("Error downloading file: " + err.Error())
	}

	fmt.Println("File downloaded and saved as 'downloaded-output.jpg'")
}

func uploadFileWithPresignedURL(url string, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	req, err := http.NewRequest("PUT", url, file)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unsuccessful server response: %s\nResponse body: %s", resp.Status, string(bodyBytes))
	}

	return nil
}

func downloadFileWithPresignedURL(url string, outputPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error executing HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unsuccessful server response: %s\nResponse body: %s", resp.Status, string(bodyBytes))
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}
