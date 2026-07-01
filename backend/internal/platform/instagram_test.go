package platform

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestInstagramGenerateAuthURL(t *testing.T) {
	t.Setenv("META_GRAPH_API_VERSION", "v25.0")
	adapter := NewInstagramAdapter("client-id", "client-secret", "https://app.example/api/v1/accounts/instagram/callback")

	authURL, _ := adapter.GenerateAuthURL("state-123")
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parsing auth url: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "www.facebook.com" || parsed.Path != "/v25.0/dialog/oauth" {
		t.Fatalf("unexpected auth url %s", authURL)
	}
	query := parsed.Query()
	if query.Get(oauthParamClientID) != "client-id" {
		t.Fatalf("expected client id, got %q", query.Get(oauthParamClientID))
	}
	if !strings.Contains(query.Get("scope"), "instagram_content_publish") {
		t.Fatalf("expected instagram_content_publish scope, got %q", query.Get("scope"))
	}
}

func TestInstagramExchangeAndSelectBusinessAccount(t *testing.T) {
	t.Setenv("META_GRAPH_API_VERSION", "v25.0")
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	accountsCalls := 0
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v25.0/oauth/access_token":
			query := req.URL.Query()
			if query.Get(grantType) == "fb_exchange_token" {
				return jsonResponse(req, `{"access_token":"long-token","token_type":"bearer","expires_in":5184000}`), nil
			}
			return jsonResponse(req, `{"access_token":"short-token","token_type":"bearer","expires_in":3600}`), nil
		case "/v25.0/me/accounts":
			accountsCalls++
			if req.URL.Query().Get(oauthParamAccessToken) != "long-token" {
				t.Fatalf("unexpected accounts token %q", req.URL.Query().Get(oauthParamAccessToken))
			}
			return jsonResponse(req, `{"data":[{"id":"page-1","name":"OpenPost Page","access_token":"page-token","picture":{"data":{"url":"https://cdn.example/page.png"}},"instagram_business_account":{"id":"ig-1","username":"openpost","name":"OpenPost IG","profile_picture_url":"https://cdn.example/ig.png"}}]}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})}

	adapter := NewInstagramAdapter("client-id", "client-secret", "https://app.example/callback")
	token, err := adapter.ExchangeCode(context.Background(), "auth-code", nil)
	if err != nil {
		t.Fatalf("ExchangeCode returned error: %v", err)
	}
	if token.AccessToken != "long-token" {
		t.Fatalf("unexpected token: %#v", token)
	}

	options, err := adapter.ListAccountSelections(context.Background(), token)
	if err != nil {
		t.Fatalf("ListAccountSelections returned error: %v", err)
	}
	if len(options) != 1 || options[0].ID != "ig-1" || options[0].Username != "openpost" {
		t.Fatalf("unexpected options: %#v", options)
	}

	selected, err := adapter.SelectAccount(context.Background(), token, "ig-1")
	if err != nil {
		t.Fatalf("SelectAccount returned error: %v", err)
	}
	if selected.AccountID != "ig-1" || selected.Token.AccessToken != "page-token" || selected.Token.ExpiresIn != 0 {
		t.Fatalf("unexpected selected account: %#v", selected)
	}
	if accountsCalls != 2 {
		t.Fatalf("expected two accounts calls, got %d", accountsCalls)
	}
}

func TestInstagramPublishImageFromPublicURL(t *testing.T) {
	t.Setenv("META_GRAPH_API_VERSION", "v25.0")
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	var createForm url.Values
	var publishForm url.Values
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v25.0/ig-1/media":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading create body: %v", err)
			}
			createForm, err = url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("parsing create form: %v", err)
			}
			return jsonResponse(req, `{"id":"container-1"}`), nil
		case "/v25.0/container-1":
			return jsonResponse(req, `{"status_code":"FINISHED"}`), nil
		case "/v25.0/ig-1/media_publish":
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("reading publish body: %v", err)
			}
			publishForm, err = url.ParseQuery(string(body))
			if err != nil {
				t.Fatalf("parsing publish form: %v", err)
			}
			return jsonResponse(req, `{"id":"ig-media-1"}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})}

	adapter := NewInstagramAdapter("client-id", "client-secret", "https://app.example/callback")
	externalID, err := adapter.Publish(context.Background(), "page-token", "ig-1", &PublishRequest{
		Content:          "Launch image",
		PlatformMediaIDs: []string{"https://media.example/image.jpg"},
		Media:            []MediaItem{{ID: "media-1", MimeType: "image/jpeg"}},
	})
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	if externalID != "ig-media-1" {
		t.Fatalf("expected media id, got %q", externalID)
	}
	if createForm.Get("image_url") != "https://media.example/image.jpg" || createForm.Get("caption") != "Launch image" {
		t.Fatalf("unexpected create form: %s", createForm.Encode())
	}
	if publishForm.Get("creation_id") != "container-1" || publishForm.Get(oauthParamAccessToken) != "page-token" {
		t.Fatalf("unexpected publish form: %s", publishForm.Encode())
	}
}

func TestInstagramPublishRejectsNonHTTPSMediaURL(t *testing.T) {
	adapter := NewInstagramAdapter("client-id", "client-secret", "https://app.example/callback")
	_, err := adapter.Publish(context.Background(), "page-token", "ig-1", &PublishRequest{
		Content:          "Launch image",
		PlatformMediaIDs: []string{"http://media.example/image.jpg"},
		Media:            []MediaItem{{ID: "media-1", MimeType: "image/jpeg"}},
	})
	if err == nil || !strings.Contains(err.Error(), "publicly-accessible HTTPS") {
		t.Fatalf("expected HTTPS URL error, got %v", err)
	}
}
