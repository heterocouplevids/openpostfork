package mediastore

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	PublicBaseURL   string
	ForcePathStyle  bool
}

type s3ObjectClient interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type S3Storage struct {
	client        s3ObjectClient
	bucket        string
	publicBaseURL string
}

func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("OPENPOST_S3_BUCKET is required when OPENPOST_STORAGE_DRIVER=s3")
	}
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, fmt.Errorf("OPENPOST_S3_REGION is required when OPENPOST_STORAGE_DRIVER=s3")
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" {
		return nil, fmt.Errorf("OPENPOST_S3_ACCESS_KEY_ID is required when OPENPOST_STORAGE_DRIVER=s3")
	}
	if strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return nil, fmt.Errorf("OPENPOST_S3_SECRET_ACCESS_KEY is required when OPENPOST_STORAGE_DRIVER=s3")
	}

	awsCfg := aws.Config{
		Region:      cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.UsePathStyle = cfg.ForcePathStyle
		if cfg.Endpoint != "" {
			options.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	storage := newS3StorageWithClient(client, cfg)

	// Preserve the ctx parameter in the constructor signature so callers can
	// pass startup-scoped contexts when validation checks are added later.
	_ = ctx

	return storage, nil
}

func newS3StorageWithClient(client s3ObjectClient, cfg S3Config) *S3Storage {
	return &S3Storage{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}
}

func (s *S3Storage) Driver() string {
	return "s3"
}

func (s *S3Storage) Save(id string, reader io.Reader) (string, error) {
	key := cleanObjectKey(id)
	if _, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	}); err != nil {
		return "", err
	}
	return key, nil
}

func (s *S3Storage) Delete(id string) error {
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(cleanObjectKey(id)),
	})
	return err
}

func (s *S3Storage) GetURL(id string) string {
	key := cleanObjectKey(id)
	if s.publicBaseURL == "" {
		return key
	}
	return s.publicBaseURL + "/" + key
}

func (s *S3Storage) Open(id string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(cleanObjectKey(id)),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func cleanObjectKey(id string) string {
	return strings.TrimLeft(strings.TrimSpace(id), "/")
}
