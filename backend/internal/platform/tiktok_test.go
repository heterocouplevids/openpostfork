package platform

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestTikTokGenerateAuthURL(t *testing.T) {
	adapter := NewTikTokAdapter("client-key", "client-secret", "https://app.example/api/v1/accounts/tiktok/callback")

	authURL, _ := adapter.GenerateAuthURL("state-123")
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parsing auth url: %v", err)
	}

	if parsed.Scheme != "https" || parsed.Host != "www.tiktok.com" || parsed.Path != "/v2/auth/authorize/" {
		t.Fatalf("unexpected auth url %s", authURL)
	}
	query := parsed.Query()
	if query.Get("client_key") != "client-key" {
		t.Fatalf("expected client_key, got %q", query.Get("client_key"))
	}
	if query.Get("redirect_uri") != "https://app.example/api/v1/accounts/tiktok/callback" {
		t.Fatalf("unexpected redirect uri %q", query.Get("redirect_uri"))
	}
	if query.Get("response_type") != "code" {
		t.Fatalf("unexpected response_type %q", query.Get("response_type"))
	}
	if query.Get("state") != "state-123" {
		t.Fatalf("unexpected state %q", query.Get("state"))
	}
	if !strings.Contains(query.Get("scope"), "video.publish") {
		t.Fatalf("expected video.publish scope, got %q", query.Get("scope"))
	}
}

func TestTikTokExchangeCodeAndProfile(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case tiktokTokenURL:
			if req.Method != http.MethodPost {
				t.Fatalf("unexpected token method %s", req.Method)
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading token body: %v", err)
			}
			form, err := url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("parsing token form: %v", err)
			}
			if form.Get("client_key") != "client-key" || form.Get(oauthParamClientSecret) != "client-secret" {
				t.Fatalf("unexpected client credentials in form: %s", string(body))
			}
			if form.Get(grantType) != oauthGrantAuthCode || form.Get(oauthParamCode) != "auth-code" {
				t.Fatalf("unexpected grant/code in form: %s", string(body))
			}
			return jsonResponse(req, `{"access_token":"access","refresh_token":"refresh","expires_in":86400,"token_type":"Bearer","scope":"user.info.basic,video.publish","open_id":"open-1"}`), nil
		case tiktokUserInfoURL:
			if req.Header.Get(headerAuthorization) != bearerPrefix+"access" {
				t.Fatalf("unexpected profile auth header %q", req.Header.Get(headerAuthorization))
			}
			return jsonResponse(req, `{"data":{"user":{"open_id":"open-1","display_name":"Creator","username":"creator"}},"error":{"code":"ok","message":"","log_id":"log"}}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})}

	adapter := NewTikTokAdapter("client-key", "client-secret", "https://app.example/callback")
	token, err := adapter.ExchangeCode(context.Background(), "auth-code", nil)
	if err != nil {
		t.Fatalf("ExchangeCode returned error: %v", err)
	}
	if token.AccessToken != "access" || token.RefreshToken != "refresh" || token.Extra["open_id"] != "open-1" {
		t.Fatalf("unexpected token result: %#v", token)
	}

	profile, err := adapter.GetProfile(context.Background(), token.AccessToken)
	if err != nil {
		t.Fatalf("GetProfile returned error: %v", err)
	}
	if profile.ID != "open-1" || profile.Username != "creator" || profile.DisplayName != "Creator" {
		t.Fatalf("unexpected profile: %#v", profile)
	}
}

func TestTikTokPublishDirectVideoFromPublicURL(t *testing.T) {
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	var initPayload map[string]any
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get(headerAuthorization) != bearerPrefix+"access" {
			t.Fatalf("unexpected auth header %q", req.Header.Get(headerAuthorization))
		}
		switch req.URL.String() {
		case tiktokCreatorInfoURL:
			return jsonResponse(req, `{"data":{"privacy_level_options":["SELF_ONLY","PUBLIC_TO_EVERYONE"]},"error":{"code":"ok"}}`), nil
		case tiktokVideoInitURL:
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading init body: %v", err)
			}
			if err := json.Unmarshal(body, &initPayload); err != nil {
				t.Fatalf("decoding init payload: %v", err)
			}
			return jsonResponse(req, `{"data":{"publish_id":"publish-1"},"error":{"code":"ok"}}`), nil
		case tiktokPublishStatusURL:
			return jsonResponse(req, `{"data":{"status":"PUBLISH_COMPLETE","publicly_available_post_id":["video-1"]},"error":{"code":"ok"}}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})}

	adapter := NewTikTokAdapter("client-key", "client-secret", "https://app.example/callback")
	externalID, err := adapter.Publish(context.Background(), "access", "open-1", &PublishRequest{
		Content:          "Launch video",
		PlatformMediaIDs: []string{"https://media.example/video.mp4"},
		Media:            []MediaItem{{ID: "media-1", MimeType: "video/mp4"}},
	})
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	if externalID != "video-1" {
		t.Fatalf("expected video id, got %q", externalID)
	}

	postInfo, ok := initPayload["post_info"].(map[string]any)
	if !ok {
		t.Fatalf("missing post_info payload: %#v", initPayload)
	}
	if postInfo["privacy_level"] != "PUBLIC_TO_EVERYONE" || postInfo["title"] != "Launch video" {
		t.Fatalf("unexpected post_info: %#v", postInfo)
	}
	sourceInfo, ok := initPayload["source_info"].(map[string]any)
	if !ok {
		t.Fatalf("missing source_info payload: %#v", initPayload)
	}
	if sourceInfo["source"] != "PULL_FROM_URL" || sourceInfo["video_url"] != "https://media.example/video.mp4" {
		t.Fatalf("unexpected source_info: %#v", sourceInfo)
	}
}

func TestTikTokPublishRequiresHTTPSVideoURL(t *testing.T) {
	adapter := NewTikTokAdapter("client-key", "client-secret", "https://app.example/callback")
	_, err := adapter.Publish(context.Background(), "access", "open-1", &PublishRequest{
		Content:          "Launch video",
		PlatformMediaIDs: []string{"http://media.example/video.mp4"},
		Media:            []MediaItem{{ID: "media-1", MimeType: "video/mp4"}},
	})
	if err == nil || !strings.Contains(err.Error(), "publicly-accessible HTTPS") {
		t.Fatalf("expected HTTPS URL error, got %v", err)
	}
}

func jsonResponse(req *http.Request, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{contentTypeJSON}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}
