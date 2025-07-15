package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"vybes/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// UploadInfo contains metadata about a successful file upload operation
type UploadInfo struct {
	URL      string    // Public URL of the uploaded file
	Key      string    // Object key in the storage bucket
	Size     int64     // File size in bytes
	Uploaded time.Time // Timestamp when upload completed
}

// Client defines the interface for file storage operations.
// Supports uploading, downloading, and deleting files from cloud storage.
type Client interface {
	// UploadFile uploads a file to the specified bucket and returns upload metadata
	UploadFile(ctx context.Context, bucket, key string, reader io.Reader) (*UploadInfo, error)
	// DeleteFile removes a file from the specified bucket
	DeleteFile(ctx context.Context, bucket, key string) error
}

// r2Client implements the Client interface using Cloudflare R2 as the backend
type r2Client struct {
	s3Client *s3.Client
	cfg      *config.Config
}

// NewClient creates and initializes a new R2 storage client with the provided configuration.
// It ensures all required buckets exist before returning the client.
//
// Parameters:
//   - ctx: Context for the operation
//   - cfg: Configuration containing R2 credentials and settings
//
// Returns:
//   - Client: A configured storage client ready for use
//   - error: Any error that occurred during client initialization
func NewClient(ctx context.Context, cfg *config.Config) (Client, error) {
	// Create custom endpoint resolver for R2
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID),
		}, nil
	})

	// Load AWS config with R2 credentials
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithEndpointResolverWithOptions(customResolver),
		awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.R2AccessKeyID,
				SecretAccessKey: cfg.R2SecretAccessKey,
			}, nil
		})),
		awsconfig.WithRegion("auto"), // R2 uses "auto" as region
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(sdkConfig)

	// Ensure all required buckets exist
	requiredBuckets := []string{cfg.R2BucketName}
	for _, bucket := range requiredBuckets {
		if err := ensureBucketExists(ctx, s3Client, bucket); err != nil {
			return nil, fmt.Errorf("failed to ensure bucket %s exists: %w", bucket, err)
		}
	}

	return &r2Client{s3Client: s3Client, cfg: cfg}, nil
}

// ensureBucketExists checks if a bucket exists and creates it if it doesn't.
// This ensures the storage client can operate without manual bucket setup.
//
// Parameters:
//   - ctx: Context for the operation
//   - s3Client: S3 client instance
//   - bucketName: Name of the bucket to check/create
//
// Returns:
//   - error: Any error that occurred during bucket verification/creation
func ensureBucketExists(ctx context.Context, s3Client *s3.Client, bucketName string) error {
	_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucketName)})
	if err == nil {
		return nil // Bucket exists
	}

	// Bucket doesn't exist, create it
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
	}

	return nil
}

// UploadFile uploads a file to the specified bucket and returns metadata about the upload.
// The file is uploaded with public read access and appropriate content type detection.
//
// Parameters:
//   - ctx: Context for the operation
//   - bucket: Target bucket name
//   - key: Object key (file path) in the bucket
//   - reader: Reader containing the file data
//
// Returns:
//   - *UploadInfo: Metadata about the uploaded file
//   - error: Any error that occurred during upload
func (c *r2Client) UploadFile(ctx context.Context, bucket, key string, reader io.Reader) (*UploadInfo, error) {
	// Determine content type based on file extension
	contentType := "application/octet-stream"
	if ext := strings.ToLower(filepath.Ext(key)); ext != "" {
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".webp":
			contentType = "image/webp"
		case ".mp4":
			contentType = "video/mp4"
		case ".mov":
			contentType = "video/quicktime"
		}
	}

	// Upload file to R2
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead, // Make file publicly accessible
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Construct public URL
	url := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", c.cfg.R2AccountID, key)

	return &UploadInfo{
		URL:      url,
		Key:      key,
		Uploaded: time.Now(),
	}, nil
}

// DeleteFile removes a file from the specified bucket.
// This operation is irreversible and should be used with caution.
//
// Parameters:
//   - ctx: Context for the operation
//   - bucket: Source bucket name
//   - key: Object key (file path) to delete
//
// Returns:
//   - error: Any error that occurred during deletion
func (c *r2Client) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file %s from bucket %s: %w", key, bucket, err)
	}
	return nil
}
