package mediastore

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

type fakeS3Client struct {
	putBucket    string
	putKey       string
	putBody      string
	deleteBucket string
	deleteKey    string
	getBucket    string
	getKey       string
	getBody      string
}

func (f *fakeS3Client) PutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	f.putBucket = aws.ToString(input.Bucket)
	f.putKey = aws.ToString(input.Key)
	f.putBody = string(body)
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeS3Client) DeleteObject(_ context.Context, input *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	f.deleteBucket = aws.ToString(input.Bucket)
	f.deleteKey = aws.ToString(input.Key)
	return &s3.DeleteObjectOutput{}, nil
}

func (f *fakeS3Client) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	f.getBucket = aws.ToString(input.Bucket)
	f.getKey = aws.ToString(input.Key)
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewBufferString(f.getBody))}, nil
}

func TestLocalStorageReportsDriver(t *testing.T) {
	storage := NewLocalStorage(t.TempDir(), "/media")

	require.Equal(t, "local", storage.Driver())
}

func TestNewStorageRejectsUnsupportedDriver(t *testing.T) {
	storage, err := New(context.Background(), Config{Driver: "gcs"})

	require.Nil(t, storage)
	require.ErrorContains(t, err, "unsupported storage driver")
}

func TestS3StorageUsesBlobStorageContract(t *testing.T) {
	client := &fakeS3Client{getBody: "stored-content"}
	storage := newS3StorageWithClient(client, S3Config{
		Bucket:        "openpost-media",
		PublicBaseURL: "https://media.openpost.social/",
	})

	savedPath, err := storage.Save("media/example.png", bytes.NewBufferString("uploaded-content"))
	require.NoError(t, err)
	require.Equal(t, "media/example.png", savedPath)
	require.Equal(t, "openpost-media", client.putBucket)
	require.Equal(t, "media/example.png", client.putKey)
	require.Equal(t, "uploaded-content", client.putBody)
	require.Equal(t, "s3", storage.Driver())
	require.Equal(t, "https://media.openpost.social/media/example.png", storage.GetURL("media/example.png"))

	reader, err := storage.Open("media/example.png")
	require.NoError(t, err)
	defer reader.Close()
	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "stored-content", string(body))
	require.Equal(t, "openpost-media", client.getBucket)
	require.Equal(t, "media/example.png", client.getKey)

	require.NoError(t, storage.Delete("media/example.png"))
	require.Equal(t, "openpost-media", client.deleteBucket)
	require.Equal(t, "media/example.png", client.deleteKey)
}
