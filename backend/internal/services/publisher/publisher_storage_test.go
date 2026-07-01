package publisher

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/platform"
	"github.com/stretchr/testify/require"
)

type fakePublisherStorage struct {
	opened string
	body   string
}

func (f *fakePublisherStorage) Driver() string { return "s3" }
func (f *fakePublisherStorage) Save(string, io.Reader) (string, error) {
	return "", nil
}
func (f *fakePublisherStorage) Delete(string) error { return nil }
func (f *fakePublisherStorage) GetURL(string) string {
	return ""
}
func (f *fakePublisherStorage) Open(id string) (io.ReadCloser, error) {
	f.opened = id
	return io.NopCloser(strings.NewReader(f.body)), nil
}

type fakePublisherAdapter struct {
	uploadedBody string
	publishCalls int
	publishErr   error
	externalID   string
}

func (f *fakePublisherAdapter) GenerateAuthURL(string) (string, map[string]string) {
	return "", nil
}
func (f *fakePublisherAdapter) ExchangeCode(context.Context, string, map[string]string) (*platform.TokenResult, error) {
	return nil, nil
}
func (f *fakePublisherAdapter) RefreshCapability() platform.RefreshCapability {
	return platform.RefreshCapability{}
}
func (f *fakePublisherAdapter) RefreshToken(context.Context, platform.RefreshTokenInput) (*platform.TokenResult, error) {
	return nil, nil
}
func (f *fakePublisherAdapter) GetProfile(context.Context, string) (*platform.UserProfile, error) {
	return nil, nil
}
func (f *fakePublisherAdapter) UploadMedia(_ context.Context, _, _, _ string, reader io.Reader) (string, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	f.uploadedBody = string(body)
	return "platform-media-id", nil
}
func (f *fakePublisherAdapter) Publish(context.Context, string, string, *platform.PublishRequest) (string, error) {
	f.publishCalls++
	if f.publishErr != nil {
		return "", f.publishErr
	}
	if f.externalID != "" {
		return f.externalID, nil
	}
	return "external-post-id", nil
}

var errFakePublishFailed = errors.New("publish failed")

func TestUploadMediaToPlatformReadsFromBlobStorage(t *testing.T) {
	storage := &fakePublisherStorage{body: "stored-media"}
	adapter := &fakePublisherAdapter{}
	service := NewService(nil, nil)
	service.SetStorage(storage)

	got, err := service.uploadMediaToPlatform(
		context.Background(),
		&models.SocialAccount{Platform: "x", AccountID: "acct-1"},
		adapter,
		"token",
		models.MediaAttachment{FilePath: "media/example.png", MimeType: "image/png"},
	)

	require.NoError(t, err)
	require.Equal(t, "platform-media-id", got)
	require.Equal(t, "example.png", storage.opened)
	require.Equal(t, "stored-media", adapter.uploadedBody)
}

func TestUploadMediaToPlatformUsesPublicURLForTikTok(t *testing.T) {
	storage := &fakePublisherStorage{body: "stored-media"}
	adapter := &fakePublisherAdapter{}
	service := NewService(nil, nil)
	service.SetStorage(storage)
	service.SetPublicMediaURL("https://media.openpost.test/media")

	got, err := service.uploadMediaToPlatform(
		context.Background(),
		&models.SocialAccount{Platform: "tiktok", AccountID: "acct-1"},
		adapter,
		"token",
		models.MediaAttachment{ID: "media-1", FilePath: "media/example.mp4", MimeType: "video/mp4"},
	)

	require.NoError(t, err)
	require.Equal(t, "https://media.openpost.test/media/media-1.mp4", got)
	require.Empty(t, storage.opened)
	require.Empty(t, adapter.uploadedBody)
}
