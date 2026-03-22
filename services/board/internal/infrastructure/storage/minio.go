package storage

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	bucketName        = "yammi-files"
	uploadURLExpiry   = 15 * time.Minute
	downloadURLExpiry = 1 * time.Hour
)

// MinIOStorage реализует usecase.FileStorage.
// URL генерируются для internal host, затем подменяются на public.
// Nginx proxy на public host проксирует к MinIO с правильным Host header,
// что сохраняет валидность подписи.
type MinIOStorage struct {
	client       *minio.Client
	internalHost string // minio:9000 (Docker internal)
	publicHost   string // localhost:9090 (Nginx proxy → minio:9000 с Host: minio:9000)
}

func NewMinIOStorage(endpoint, publicEndpoint, accessKey, secretKey string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("check bucket existence: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket %s: %w", bucketName, err)
		}
		log.Printf("minio bucket %q created", bucketName)
	}

	if publicEndpoint == "" {
		publicEndpoint = endpoint
	}

	log.Printf("minio storage connected to %s (public: %s)", endpoint, publicEndpoint)
	return &MinIOStorage{
		client:       client,
		internalHost: endpoint,
		publicHost:   publicEndpoint,
	}, nil
}

// replaceHost заменяет internal хост на публичный в presigned URL.
// Nginx proxy на publicHost проксирует к MinIO с Host: internalHost,
// поэтому подпись остаётся валидной.
func (s *MinIOStorage) replaceHost(rawURL string) string {
	if s.internalHost == s.publicHost {
		return rawURL
	}
	return strings.Replace(rawURL, s.internalHost, s.publicHost, 1)
}

func (s *MinIOStorage) GenerateUploadURL(ctx context.Context, key, contentType string, size int64) (string, error) {
	presignedURL, err := s.client.PresignedPutObject(ctx, bucketName, key, uploadURLExpiry)
	if err != nil {
		return "", fmt.Errorf("generate upload url: %w", err)
	}
	return s.replaceHost(presignedURL.String()), nil
}

func (s *MinIOStorage) GenerateDownloadURL(ctx context.Context, key string) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, bucketName, key, downloadURLExpiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("generate download url: %w", err)
	}
	return s.replaceHost(presignedURL.String()), nil
}

func (s *MinIOStorage) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, bucketName, key, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, bucketName, key, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("stat object %s: %w", key, err)
	}
	return true, nil
}
