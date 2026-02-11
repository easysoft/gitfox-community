package logs

import (
	"context"
	"io"
	"path"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var _ Interface = (*Store)(nil)

// Store 插件存储实现
type Store struct {
	bucket  string
	prefix  string
	session *session.Session
}

// NewS3Store 创建S3存储实例
func NewS3Store(s3cfg *config.S3Config) (*Store, error) {
	disableSSL := false

	if s3cfg.Endpoint != "" {
		disableSSL = !strings.HasPrefix(s3cfg.Endpoint, "https://")
	}

	region := s3cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	cfg := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(s3cfg.Endpoint),
		DisableSSL:       aws.Bool(disableSSL),
		S3ForcePathStyle: aws.Bool(s3cfg.PathStyle),
	}

	if s3cfg.AccessKey != "" && s3cfg.SecretKey != "" {
		cfg.Credentials = credentials.NewStaticCredentials(s3cfg.AccessKey, s3cfg.SecretKey, "")
	}

	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, errors.Newf("failed to create S3 session: %v", err)
	}

	// 验证 bucket 是否可达及凭证是否有效
	if _, err = s3.New(sess).HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(s3cfg.Bucket)}); err != nil {
		return nil, errors.Newf("failed to access bucket %q: %v", s3cfg.Bucket, err)
	}

	return &Store{
		bucket:  s3cfg.Bucket,
		prefix:  s3cfg.Prefix,
		session: sess,
	}, nil
}

func (s *Store) Find(_ context.Context, filePath string) (io.ReadCloser, error) {
	svc := s3.New(s.session)
	out, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(filePath)),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (s *Store) Create(_ context.Context, filePath string, r io.Reader) error {
	uploader := s3manager.NewUploader(s.session)
	input := &s3manager.UploadInput{
		ACL:    aws.String("private"),
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(filePath)),
		Body:   r,
	}
	_, err := uploader.Upload(input)
	return err
}

func (s *Store) Update(ctx context.Context, filePath string, r io.Reader) error {
	return s.Create(ctx, filePath, r)
}

func (s *Store) Delete(_ context.Context, filePath string) error {
	svc := s3.New(s.session)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key(filePath)),
	})
	return err
}

// key 生成 S3 对象路径: {prefix}/pipelineID/executionID/stageID/stepID.log
func (s *Store) key(filePath string) string {
	return path.Join("/", s.prefix, filePath)
}
