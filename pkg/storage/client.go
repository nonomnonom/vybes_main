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

// NewClient creates a new MinIO client and ensures the bucket exists.
func NewClient(ctx context.Context, cfg *config.Config) (Client, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	// Check if the bucket already exists.
	exists, err := client.BucketExists(ctx, cfg.MinioBucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Create the bucket if it does not exist.
		err = client.MakeBucket(ctx, cfg.MinioBucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		log.Info().Str("bucket", cfg.MinioBucketName).Msg("Successfully created bucket")
	}

	return &minioStorageClient{client: client}, nil
}

// UploadFile uploads a file to the specified bucket.
func (c *minioStorageClient) UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, contentType string) (minio.UploadInfo, error) {
	return c.client.PutObject(ctx, bucketName, objectName, reader, objectSize, minio.PutObjectOptions{ContentType: contentType})
}

// DeleteFile removes a file from the specified bucket.
func (c *minioStorageClient) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	return c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}