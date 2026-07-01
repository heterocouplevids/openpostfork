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

func TestFacebookGenerateAuthURL(t *testing.T) {
	t.Setenv("META_GRAPH_API_VERSION", "v25.0")
	adapter := NewFacebookAdapter("client-id", "client-secret", "https://app.example/api/v1/accounts/facebook/callback")

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
	if query.Get("response_type") != oauthResponseType {
		t.Fatalf("unexpected response_type %q", query.Get("response_type"))
	}
	if !strings.Contains(query.Get("scope"), "pages_manage_posts") {
		t.Fatalf("expected pages_manage_posts scope, got %q", query.Get("scope"))
	}
}

func TestFacebookExchangeAndSelectPage(t *testing.T) {
	t.Setenv("META_GRAPH_API_VERSION", "v25.0")
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	accountsCalls := 0
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/v25.0/oauth/access_token":
			query := req.URL.Query()
			switch query.Get(grantType) {
			case "fb_exchange_token":
				if query.Get("fb_exchange_token") != "short-token" {
					t.Fatalf("unexpected fb_exchange_token %q", query.Get("fb_exchange_token"))
				}
				return jsonResponse(req, `{"access_token":"long-token","token_type":"bearer","expires_in":5184000}`), nil
			default:
				if query.Get(oauthParamCode) != "auth-code" {
					t.Fatalf("unexpected code %q", query.Get(oauthParamCode))
				}
				return jsonResponse(req, `{"access_token":"short-token","token_type":"bearer","expires_in":3600}`), nil
			}
		case "/v25.0/me/accounts":
			accountsCalls++
			if req.URL.Query().Get(oauthParamAccessToken) != "long-token" {
				t.Fatalf("unexpected accounts token %q", req.URL.Query().Get(oauthParamAccessToken))
			}
			return jsonResponse(req, `{"data":[{"id":"page-1","name":"OpenPost Page","username":"openpost","access_token":"page-token","picture":{"data":{"url":"https://cdn.example/page.png"}}}]}`), nil
		default:
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
			return nil, nil
		}
	})}

	adapter := NewFacebookAdapter("client-id", "client-secret", "https://app.example/callback")
	token, err := adapter.ExchangeCode(context.Background(), "auth-code", nil)
	if err != nil {
		t.Fatalf("ExchangeCode returned error: %v", err)
	}
	if token.AccessToken != "long-token" || token.ExpiresIn != 5184000 {
		t.Fatalf("unexpected token: %#v", token)
	}

	options, err := adapter.ListAccountSelections(context.Background(), token)
	if err != nil {
		t.Fatalf("ListAccountSelections returned error: %v", err)
	}
	if len(options) != 1 || options[0].ID != "page-1" || options[0].Username != "openpost" {
		t.Fatalf("unexpected options: %#v", options)
	}

	selected, err := adapter.SelectAccount(context.Background(), token, "page-1")
	if err != nil {
		t.Fatalf("SelectAccount returned error: %v", err)
	}
	if selected.AccountID != "page-1" || selected.Token.AccessToken != "page-token" || selected.Token.ExpiresIn != 0 {
		t.Fatalf("unexpected selected account: %#v", selected)
	}
	if accountsCalls != 2 {
		t.Fatalf("expected two accounts calls, got %d", accountsCalls)
	}
}

func TestFacebookPublishPhotoFromPublicURL(t *testing.T) {
	t.Setenv("META_GRAPH_API_VERSION", "v25.0")
	originalClient := httpClient
	defer func() { httpClient = originalClient }()

	var form url.Values
	httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/v25.0/page-1/photos" {
			t.Fatalf("unexpected request %s %s", req.Method, req.URL.String())
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("reading publish body: %v", err)
		}
		form, err = url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("parsing publish form: %v", err)
		}
		return jsonResponse(req, `{"id":"photo-1","post_id":"page-1_post-1"}`), nil
	})}

	adapter := NewFacebookAdapter("client-id", "client-secret", "https://app.example/callback")
	externalID, err := adapter.Publish(context.Background(), "page-token", "page-1", &PublishRequest{
		Content:          "Launch photo",
		PlatformMediaIDs: []string{"https://media.example/photo.jpg"},
		Media:            []MediaItem{{ID: "media-1", MimeType: "image/jpeg"}},
	})
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}
	if externalID != "page-1_post-1" {
		t.Fatalf("expected post id, got %q", externalID)
	}
	if form.Get("url") != "https://media.example/photo.jpg" || form.Get("caption") != "Launch photo" || form.Get(oauthParamAccessToken) != "page-token" {
		t.Fatalf("unexpected publish form: %s", form.Encode())
	}
}

func TestFacebookPublishRejectsNonHTTPSMediaURL(t *testing.T) {
	adapter := NewFacebookAdapter("client-id", "client-secret", "https://app.example/callback")
	_, err := adapter.Publish(context.Background(), "page-token", "page-1", &PublishRequest{
		Content:          "Launch photo",
		PlatformMediaIDs: []string{"http://media.example/photo.jpg"},
		Media:            []MediaItem{{ID: "media-1", MimeType: "image/jpeg"}},
	})
	if err == nil || !strings.Contains(err.Error(), "publicly-accessible HTTPS") {
		t.Fatalf("expected HTTPS URL error, got %v", err)
	}
}

func TestFacebookPublishedIDRejectsGraphError(t *testing.T) {
	_, err := facebookPublishedID("facebook publish", []byte(`{"error":{"message":"missing permission"}}`))
	if err == nil || !strings.Contains(err.Error(), "missing permission") {
		t.Fatalf("expected graph error, got %v", err)
	}
}

func TestFacebookPublishedIDParsesID(t *testing.T) {
	body, err := json.Marshal(map[string]string{"id": "post-1"})
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	got, err := facebookPublishedID("facebook publish", body)
	if err != nil {
		t.Fatalf("facebookPublishedID returned error: %v", err)
	}
	if got != "post-1" {
		t.Fatalf("expected post-1, got %q", got)
	}
}
