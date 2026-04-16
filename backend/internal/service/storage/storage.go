package storage

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// StorageService 对象存储抽象接口
type StorageService interface {
	Upload(ctx context.Context, folder, filename string, body io.Reader, contentType string) (string, error)
}

// S3Storage S3 兼容存储实现
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage 创建 S3 存储服务
func NewS3Storage(endpoint, accessKey, secretKey, bucket, region string) (*S3Storage, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("S3 配置不完整: endpoint/accessKey/secretKey 不能为空")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("加载 S3 配置失败: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &S3Storage{client: client, bucket: bucket}, nil
}

// Upload 上传文件到 S3，返回对象 URL
func (s *S3Storage) Upload(ctx context.Context, folder, filename string, body io.Reader, contentType string) (string, error) {
	ext := path.Ext(filename)
	objectKey := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %w", err)
	}

	return objectKey, nil
}

// NoopStorage 空实现，S3 未配置时使用
type NoopStorage struct{}

func (n *NoopStorage) Upload(_ context.Context, _, _ string, _ io.Reader, _ string) (string, error) {
	return "", fmt.Errorf("对象存储未配置，无法上传文件")
}

// NewStorageService 根据配置决定创建 S3 或 Noop 存储
func NewStorageService(endpoint, accessKey, secretKey, bucket, region string) StorageService {
	s3svc, err := NewS3Storage(endpoint, accessKey, secretKey, bucket, region)
	if err != nil {
		return &NoopStorage{}
	}
	return s3svc
}
