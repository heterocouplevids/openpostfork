package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestYouTubeGenerateAuthURL(t *testing.T) {
	adapter := NewYouTubeAdapter("client-id", "client-secret", "https://app.example/api/v1/accounts/youtube/callback")

	authURL, _ := adapter.GenerateAuthURL("state-123")
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parsing auth url: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "accounts.google.com" || parsed.Path != "/o/oauth2/v2/auth" {
		t.Fatalf("unexpected auth url %s", authURL)
	}
	query := parsed.Query()
	if query.Get(oauthParamClientID) != "client-id" {
		t.Fatalf("expected client id, got %q", query.Get(oauthParamClientID))
	}
	if query.Get("access_type") != "offline" || query.Get("prompt") != "consent" {
		t.Fatalf("expected offline consent auth URL, got %s", authURL)
	}
	if !strings.Contains(query.Get("scope"), "youtube.upload") {
		t.Fatalf("expected youtube.upload scope, got %q", query.Get("scope"))
	}
}

func TestYouTubeExchangeRefreshAndSelectChannel(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	channelsCalls := 0
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/token":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading token body: %v", err)
			}
			values, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("parsing token body: %v", err)
			}
			if values.Get(grantType) == oauthGrantRefresh {
				return jsonResponse(req, `{"access_token":"refreshed-token","expires_in":3600,"token_type":"Bearer"}`), nil
			}
			return jsonResponse(req, `{"access_token":"access-token","refresh_token":"refresh-token","expires_in":3600,"token_type":"Bearer","scope":"https://www.googleapis.com/auth/youtube.upload"}`), nil
		case "/youtube/v3/channels":
			channelsCalls++
			if req.Header.Get(headerAuthorization) != "Bearer access-token" {
				t.Fatalf("unexpected auth header %q", req.Header.Get(headerAuthorization))
			}
			if req.URL.Query().Get("mine") != "true" {
				t.Fatalf("expected mine=true, got %s", req.URL.RawQuery)
			}
			return jsonResponse(req, `{"items":[{"id":"channel-1","snippet":{"title":"OpenPost Channel","customUrl":"@openpost","thumbnails":{"default":{"url":"https://yt.example/avatar.jpg"}}},"statistics":{"subscriberCount":"123"}}]}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})}

	adapter := NewYouTubeAdapter("client-id", "client-secret", "https://app.example/callback")
	token, err := adapter.ExchangeCode(context.Background(), "auth-code", nil)
	if err != nil {
		t.Fatalf("ExchangeCode returned error: %v", err)
	}
	if token.AccessToken != "access-token" || token.RefreshToken != "refresh-token" {
		t.Fatalf("unexpected token: %#v", token)
	}

	refreshed, err := adapter.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: "refresh-token"})
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}
	if refreshed.AccessToken != "refreshed-token" {
		t.Fatalf("unexpected refreshed token: %#v", refreshed)
	}

	options, err := adapter.ListAccountSelections(context.Background(), token)
	if err != nil {
		t.Fatalf("ListAccountSelections returned error: %v", err)
	}
	if len(options) != 1 || options[0].ID != "channel-1" || options[0].Username != "@openpost" {
		t.Fatalf("unexpected options: %#v", options)
	}

	selected, err := adapter.SelectAccount(context.Background(), token, "channel-1")
	if err != nil {
		t.Fatalf("SelectAccount returned error: %v", err)
	}
	if selected.AccountID != "channel-1" || selected.Token.RefreshToken != "refresh-token" {
		t.Fatalf("unexpected selected account: %#v", selected)
	}
	if selected.Token.Extra["channel_id"] != "channel-1" {
		t.Fatalf("expected selected token channel id, got %#v", selected.Token.Extra)
	}
	if channelsCalls != 2 {
		t.Fatalf("expected two channels calls, got %d", channelsCalls)
	}
}

func TestYouTubeUploadMediaWithMetadata(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	var metadata youtubeVideoInsertRequest
	var mediaBody string
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Path != "/upload/youtube/v3/videos" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
		}
		if req.URL.Query().Get("uploadType") != "multipart" || req.URL.Query().Get("part") != "snippet,status" {
			t.Fatalf("unexpected upload query %s", req.URL.RawQuery)
		}
		if req.Header.Get(headerAuthorization) != "Bearer access-token" {
			t.Fatalf("unexpected auth header %q", req.Header.Get(headerAuthorization))
		}
		_, params, err := mime.ParseMediaType(req.Header.Get(headerContentType))
		if err != nil {
			t.Fatalf("parsing content type: %v", err)
		}
		reader := multipart.NewReader(req.Body, params["boundary"])
		metaPart, err := reader.NextPart()
		if err != nil {
			t.Fatalf("reading metadata part: %v", err)
		}
		metaBytes, err := io.ReadAll(metaPart)
		if err != nil {
			t.Fatalf("reading metadata body: %v", err)
		}
		if err := json.Unmarshal(metaBytes, &metadata); err != nil {
			t.Fatalf("decoding metadata: %v", err)
		}
		mediaPart, err := reader.NextPart()
		if err != nil {
			t.Fatalf("reading media part: %v", err)
		}
		mediaBytes, err := io.ReadAll(mediaPart)
		if err != nil {
			t.Fatalf("reading media body: %v", err)
		}
		mediaBody = string(mediaBytes)
		return jsonResponse(req, `{"id":"youtube-video-1"}`), nil
	})}

	adapter := NewYouTubeAdapter("client-id", "client-secret", "https://app.example/callback")
	videoID, err := adapter.UploadMediaWithMetadata(context.Background(), "access-token", "channel-1", UploadMediaRequest{
		MimeType:    "video/mp4",
		Title:       "Launch Short",
		Description: "Launch Short\nDetailed description",
		Reader:      bytes.NewBufferString("video-bytes"),
	})
	if err != nil {
		t.Fatalf("UploadMediaWithMetadata returned error: %v", err)
	}
	if videoID != "youtube-video-1" {
		t.Fatalf("expected video id, got %q", videoID)
	}
	if metadata.Snippet.Title != "Launch Short" || metadata.Status.PrivacyStatus != "private" {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	if mediaBody != "video-bytes" {
		t.Fatalf("unexpected media body %q", mediaBody)
	}
}

func TestValidateMediaYouTubeRequiresOneVideo(t *testing.T) {
	RegisterAllMediaValidators()

	issues := ValidateMedia(providerYouTube, nil)
	if len(issues) != 1 {
		t.Fatalf("expected one missing-media issue, got %d", len(issues))
	}

	issues = ValidateMedia(providerYouTube, []MediaItem{{ID: "image", MimeType: "image/png"}})
	if len(issues) != 1 {
		t.Fatalf("expected one unsupported-media issue, got %d", len(issues))
	}

	issues = ValidateMedia(providerYouTube, []MediaItem{{ID: "video", MimeType: "video/mp4"}})
	if len(issues) != 0 {
		t.Fatalf("expected no issues for one video, got %#v", issues)
	}
}
