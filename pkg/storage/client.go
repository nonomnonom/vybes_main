package storage

import (
	"context"
	"io"
	"vybes/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

// UploadInfo represents the result of a file upload
type UploadInfo struct {
	Location string
	ETag     string
	Version  string
}

// Client defines the interface for a file storage client.
type Client interface {
	UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (*UploadInfo, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
}

type r2StorageClient struct {
	client *s3.Client
}

// NewClient creates a new R2 client and ensures the required buckets exist.
func NewClient(ctx context.Context, cfg *config.Config) (Client, error) {
	// Create custom endpoint resolver for R2
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: cfg.R2Endpoint,
		}, nil
	})

	// Load AWS config with R2 credentials
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2AccessKeyID,
			cfg.R2SecretAccessKey,
			"",
		)),
		awsconfig.WithRegion("auto"), // R2 uses "auto" as region
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg)

	// Ensure all required buckets exist
	buckets := []string{cfg.R2PostsBucket, cfg.R2StoriesBucket}
	for _, bucket := range buckets {
		if err := ensureBucketExists(ctx, client, bucket); err != nil {
			return nil, err
		}
	}

	return &r2StorageClient{client: client}, nil
}

// ensureBucketExists checks if a bucket exists and creates it if it doesn't.
func ensureBucketExists(ctx context.Context, client *s3.Client, bucketName string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Bucket doesn't exist, create it
		_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			return err
		}
		log.Info().Str("bucket", bucketName).Msg("Successfully created bucket")
	}
	return nil
}

// UploadFile uploads a file to the specified bucket.
func (c *r2StorageClient) UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (*UploadInfo, error) {
	result, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectName),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, err
	}

	return &UploadInfo{
		Location: objectName,
		ETag:     *result.ETag,
		Version:  "",
	}, nil
}

// DeleteFile removes a file from the specified bucket.
func (c *r2StorageClient) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectName),
	})
	return err
}
