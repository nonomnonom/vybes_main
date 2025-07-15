package storage

import (
	"context"
	"io"
	"vybes/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

// Client defines the interface for a file storage client.
type Client interface {
	UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (minio.UploadInfo, error)
	DeleteFile(ctx context.Context, bucketName, objectName string) error
}

type minioStorageClient struct {
	client *minio.Client
}

// NewClient creates a new MinIO client and ensures the required buckets exist.
func NewClient(ctx context.Context, cfg *config.Config) (Client, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	// Ensure all required buckets exist
	buckets := []string{cfg.MinioPostsBucket, cfg.MinioStoriesBucket}
	for _, bucket := range buckets {
		if err := ensureBucketExists(ctx, client, bucket); err != nil {
			return nil, err
		}
	}

	return &minioStorageClient{client: client}, nil
}

// ensureBucketExists checks if a bucket exists and creates it if it doesn't.
func ensureBucketExists(ctx context.Context, client *minio.Client, bucketName string) error {
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
		log.Info().Str("bucket", bucketName).Msg("Successfully created bucket")
	}
	return nil
}

// UploadFile uploads a file to the specified bucket.
func (c *minioStorageClient) UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (minio.UploadInfo, error) {
	return c.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: contentType})
}

// DeleteFile removes a file from the specified bucket.
func (c *minioStorageClient) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	return c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}