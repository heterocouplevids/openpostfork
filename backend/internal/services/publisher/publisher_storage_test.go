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
	uploadCalls  int
	publishCalls int
	publishErr   error
	externalID   string
	lastRequest  *platform.PublishRequest
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
	f.uploadCalls++
	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	f.uploadedBody = string(body)
	return "platform-media-id", nil
}
func (f *fakePublisherAdapter) Publish(_ context.Context, _, _ string, req *platform.PublishRequest) (string, error) {
	f.publishCalls++
	f.lastRequest = req
	if f.publishErr != nil {
		return "", f.publishErr
	}
	if f.externalID != "" {
		return f.externalID, nil
	}
	return "external-post-id", nil
}

type fakeMetadataPublisherAdapter struct {
	fakePublisherAdapter
	uploadReq platform.UploadMediaRequest
}

func (f *fakeMetadataPublisherAdapter) UploadMediaWithMetadata(_ context.Context, _, _ string, req platform.UploadMediaRequest) (string, error) {
	f.uploadCalls++
	body, err := io.ReadAll(req.Reader)
	if err != nil {
		return "", err
	}
	f.uploadedBody = string(body)
	req.Reader = nil
	f.uploadReq = req
	return "metadata-media-id", nil
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
		"Launch\nDescription",
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
		"Launch video",
	)

	require.NoError(t, err)
	require.Equal(t, "https://media.openpost.test/media/media-1.mp4", got)
	require.Empty(t, storage.opened)
	require.Empty(t, adapter.uploadedBody)
}

func TestUploadMediaToPlatformUsesPublicURLForFacebook(t *testing.T) {
	storage := &fakePublisherStorage{body: "stored-media"}
	adapter := &fakePublisherAdapter{}
	service := NewService(nil, nil)
	service.SetStorage(storage)
	service.SetPublicMediaURL("https://media.openpost.test/media")

	got, err := service.uploadMediaToPlatform(
		context.Background(),
		&models.SocialAccount{Platform: "facebook", AccountID: "acct-1"},
		adapter,
		"token",
		models.MediaAttachment{ID: "media-1", FilePath: "media/example.jpg", MimeType: "image/jpeg"},
		"Launch image",
	)

	require.NoError(t, err)
	require.Equal(t, "https://media.openpost.test/media/media-1.jpg", got)
	require.Empty(t, storage.opened)
	require.Empty(t, adapter.uploadedBody)
}

func TestUploadMediaToPlatformUsesPublicURLForInstagram(t *testing.T) {
	storage := &fakePublisherStorage{body: "stored-media"}
	adapter := &fakePublisherAdapter{}
	service := NewService(nil, nil)
	service.SetStorage(storage)
	service.SetPublicMediaURL("https://media.openpost.test/media")

	got, err := service.uploadMediaToPlatform(
		context.Background(),
		&models.SocialAccount{Platform: "instagram", AccountID: "acct-1"},
		adapter,
		"token",
		models.MediaAttachment{ID: "media-1", FilePath: "media/example.jpg", MimeType: "image/jpeg"},
		"Launch image",
	)

	require.NoError(t, err)
	require.Equal(t, "https://media.openpost.test/media/media-1.jpg", got)
	require.Empty(t, storage.opened)
	require.Empty(t, adapter.uploadedBody)
}

func TestUploadMediaToPlatformUsesMetadataUploader(t *testing.T) {
	storage := &fakePublisherStorage{body: "stored-video"}
	adapter := &fakeMetadataPublisherAdapter{}
	service := NewService(nil, nil)
	service.SetStorage(storage)

	got, err := service.uploadMediaToPlatform(
		context.Background(),
		&models.SocialAccount{Platform: "youtube", AccountID: "channel-1"},
		adapter,
		"token",
		models.MediaAttachment{ID: "media-1", FilePath: "media/example.mp4", MimeType: "video/mp4", OriginalFilename: "example.mp4"},
		"Launch title\nLonger description",
	)

	require.NoError(t, err)
	require.Equal(t, "metadata-media-id", got)
	require.Equal(t, "example.mp4", storage.opened)
	require.Equal(t, "stored-video", adapter.uploadedBody)
	require.Equal(t, "video/mp4", adapter.uploadReq.MimeType)
	require.Equal(t, "example.mp4", adapter.uploadReq.Filename)
	require.Equal(t, "Launch title", adapter.uploadReq.Title)
	require.Equal(t, "Launch title\nLonger description", adapter.uploadReq.Description)
}
