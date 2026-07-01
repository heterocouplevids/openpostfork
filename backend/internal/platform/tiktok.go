package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

const (
	tiktokAuthURL          = "https://www.tiktok.com/v2/auth/authorize/"
	tiktokTokenURL         = "https://open.tiktokapis.com/v2/oauth/token/"
	tiktokUserInfoURL      = "https://open.tiktokapis.com/v2/user/info/?fields=open_id,union_id,avatar_url,display_name,username"
	tiktokCreatorInfoURL   = "https://open.tiktokapis.com/v2/post/publish/creator_info/query/"
	tiktokVideoInitURL     = "https://open.tiktokapis.com/v2/post/publish/video/init/"
	tiktokPublishStatusURL = "https://open.tiktokapis.com/v2/post/publish/status/fetch/"
	tiktokTitleMaxRunes    = 2000
)

type TikTokAdapter struct {
	clientKey    string
	clientSecret string
	redirectURI  string
}

func NewTikTokAdapter(clientKey, clientSecret, redirectURI string) *TikTokAdapter {
	return &TikTokAdapter{
		clientKey:    clientKey,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

func (t *TikTokAdapter) GenerateAuthURL(state string) (string, map[string]string) {
	params := url.Values{}
	params.Set("client_key", t.clientKey)
	params.Set(oauthParamRedirectURI, t.redirectURI)
	params.Set("response_type", oauthResponseType)
	params.Set("scope", strings.Join(tiktokScopes(), ","))
	params.Set("state", state)
	return tiktokAuthURL + "?" + params.Encode(), nil
}

func (t *TikTokAdapter) ExchangeCode(ctx context.Context, code string, _ map[string]string) (*TokenResult, error) {
	values := map[string]string{
		"client_key":           t.clientKey,
		oauthParamClientSecret: t.clientSecret,
		oauthParamCode:         code,
		grantType:              oauthGrantAuthCode,
		oauthParamRedirectURI:  t.redirectURI,
	}
	tokenResp, err := t.exchangeToken(ctx, values, "tiktok token exchange")
	if err != nil {
		return nil, err
	}
	return tokenResp, nil
}

func (t *TikTokAdapter) RefreshCapability() RefreshCapability {
	return RefreshCapability{
		Supported:        true,
		CredentialSource: RefreshCredentialRefreshToken,
	}
}

func (t *TikTokAdapter) RefreshToken(ctx context.Context, input RefreshTokenInput) (*TokenResult, error) {
	if input.RefreshToken == "" {
		return nil, fmt.Errorf("tiktok refresh requires a refresh token")
	}
	values := map[string]string{
		"client_key":                          t.clientKey,
		oauthParamClientSecret:                t.clientSecret,
		grantType:                             oauthGrantRefresh,
		string(RefreshCredentialRefreshToken): input.RefreshToken,
	}
	return t.exchangeToken(ctx, values, "tiktok token refresh")
}

func (t *TikTokAdapter) exchangeToken(ctx context.Context, values map[string]string, label string) (*TokenResult, error) {
	respBody, err := DoFormURLEncoded(ctx, "POST", tiktokTokenURL, values, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		OpenID       string `json:"open_id"`
		Error        string `json:"error"`
		Description  string `json:"error_description"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", label, err)
	}
	if tokenResp.Error != "" {
		return nil, fmt.Errorf("%s: %s", label, firstNonEmptyString(tokenResp.Description, tokenResp.Error))
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%s: missing access token", label)
	}

	extra := map[string]string{}
	if tokenResp.OpenID != "" {
		extra["open_id"] = tokenResp.OpenID
	}
	if tokenResp.Scope != "" {
		extra["scope"] = tokenResp.Scope
	}

	return &TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		TokenType:    firstNonEmptyString(tokenResp.TokenType, tokenTypeBearer),
		Extra:        extra,
	}, nil
}

func (t *TikTokAdapter) GetProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	respBody, err := DoRequest(ctx, "GET", tiktokUserInfoURL, nil, map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("tiktok profile: %w", err)
	}

	var profileResp struct {
		Data struct {
			User struct {
				OpenID      string `json:"open_id"`
				DisplayName string `json:"display_name"`
				Username    string `json:"username"`
				AvatarURL   string `json:"avatar_url"`
			} `json:"user"`
		} `json:"data"`
		Error tiktokAPIError `json:"error"`
	}
	if err := json.Unmarshal(respBody, &profileResp); err != nil {
		return nil, fmt.Errorf("decoding tiktok profile: %w", err)
	}
	if err := profileResp.Error.err("tiktok profile"); err != nil {
		return nil, err
	}
	if profileResp.Data.User.OpenID == "" {
		return nil, fmt.Errorf("tiktok profile: missing open_id")
	}

	username := firstNonEmptyString(profileResp.Data.User.Username, profileResp.Data.User.DisplayName)
	return &UserProfile{
		ID:          profileResp.Data.User.OpenID,
		Username:    username,
		DisplayName: profileResp.Data.User.DisplayName,
	}, nil
}

func (t *TikTokAdapter) UploadMedia(_ context.Context, _ string, _ string, _ string, _ io.Reader) (string, error) {
	return "", fmt.Errorf("tiktok requires publicly accessible HTTPS media URLs for the initial adapter")
}

func (t *TikTokAdapter) Publish(ctx context.Context, accessToken, _ string, req *PublishRequest) (string, error) {
	if len(req.PlatformMediaIDs) != 1 {
		return "", fmt.Errorf("tiktok video publishing requires exactly one media URL")
	}
	if len(req.Media) != 1 || !isVideoMime(req.Media[0].MimeType) {
		return "", fmt.Errorf("tiktok initial adapter supports one video attachment")
	}
	mediaURL := req.PlatformMediaIDs[0]
	if !strings.HasPrefix(mediaURL, "https://") {
		return "", fmt.Errorf("tiktok requires a publicly-accessible HTTPS media URL. Set OPENPOST_MEDIA_URL to your public media base URL")
	}

	privacyLevel, err := t.defaultPrivacyLevel(ctx, accessToken)
	if err != nil {
		return "", err
	}

	payload := map[string]any{
		"post_info": map[string]any{
			"title":           tiktokTitle(req.Content),
			"privacy_level":   privacyLevel,
			"disable_duet":    false,
			"disable_comment": false,
			"disable_stitch":  false,
		},
		"source_info": map[string]any{
			"source":    "PULL_FROM_URL",
			"video_url": mediaURL,
		},
	}

	respBody, err := DoJSON(ctx, "POST", tiktokVideoInitURL, payload, map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
	})
	if err != nil {
		return "", fmt.Errorf("tiktok video init: %w", err)
	}

	var initResp struct {
		Data struct {
			PublishID string `json:"publish_id"`
		} `json:"data"`
		Error tiktokAPIError `json:"error"`
	}
	if err := json.Unmarshal(respBody, &initResp); err != nil {
		return "", fmt.Errorf("decoding tiktok video init: %w", err)
	}
	if err := initResp.Error.err("tiktok video init"); err != nil {
		return "", err
	}
	if initResp.Data.PublishID == "" {
		return "", fmt.Errorf("tiktok video init: missing publish_id")
	}

	return t.waitForPublishID(ctx, accessToken, initResp.Data.PublishID)
}

func (t *TikTokAdapter) defaultPrivacyLevel(ctx context.Context, accessToken string) (string, error) {
	respBody, err := DoJSON(ctx, "POST", tiktokCreatorInfoURL, map[string]any{}, map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
	})
	if err != nil {
		return "", fmt.Errorf("tiktok creator info: %w", err)
	}

	var creatorResp struct {
		Data struct {
			PrivacyLevelOptions []string `json:"privacy_level_options"`
		} `json:"data"`
		Error tiktokAPIError `json:"error"`
	}
	if err := json.Unmarshal(respBody, &creatorResp); err != nil {
		return "", fmt.Errorf("decoding tiktok creator info: %w", err)
	}
	if err := creatorResp.Error.err("tiktok creator info"); err != nil {
		return "", err
	}

	for _, option := range creatorResp.Data.PrivacyLevelOptions {
		if option == "PUBLIC_TO_EVERYONE" {
			return option, nil
		}
	}
	for _, option := range creatorResp.Data.PrivacyLevelOptions {
		if option == "SELF_ONLY" {
			return option, nil
		}
	}
	if len(creatorResp.Data.PrivacyLevelOptions) > 0 {
		return creatorResp.Data.PrivacyLevelOptions[0], nil
	}
	return "", fmt.Errorf("tiktok creator info: no privacy level options returned")
}

func (t *TikTokAdapter) waitForPublishID(ctx context.Context, accessToken, publishID string) (string, error) {
	const maxAttempts = 6
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		respBody, err := DoJSON(ctx, "POST", tiktokPublishStatusURL, map[string]any{
			"publish_id": publishID,
		}, map[string]string{
			headerAuthorization: bearerPrefix + accessToken,
		})
		if err != nil {
			return "", fmt.Errorf("tiktok publish status: %w", err)
		}

		var statusResp struct {
			Data struct {
				Status                   string   `json:"status"`
				PubliclyAvailablePostID  []string `json:"publicly_available_post_id"`
				PublicalyAvailablePostID []string `json:"publicaly_available_post_id"` //nolint:misspell
				FailReason               string   `json:"fail_reason"`
			} `json:"data"`
			Error tiktokAPIError `json:"error"`
		}
		if err := json.Unmarshal(respBody, &statusResp); err != nil {
			return "", fmt.Errorf("decoding tiktok publish status: %w", err)
		}
		if err := statusResp.Error.err("tiktok publish status"); err != nil {
			return "", err
		}

		switch statusResp.Data.Status {
		case "PUBLISH_COMPLETE":
			if ids := firstNonEmptyStringSlice(statusResp.Data.PubliclyAvailablePostID, statusResp.Data.PublicalyAvailablePostID); len(ids) > 0 {
				return ids[0], nil
			}
			return publishID, nil
		case "SEND_TO_USER_INBOX":
			return publishID, nil
		case platformStatusFailed:
			return "", fmt.Errorf("tiktok publish failed: %s", firstNonEmptyString(statusResp.Data.FailReason, "unknown reason"))
		}

		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(10 * time.Second):
			}
		}
	}

	return publishID, nil
}

func validateTikTokMedia(media []MediaItem) []MediaValidationIssue {
	if len(media) != 1 {
		return []MediaValidationIssue{{
			Provider: providerTikTok,
			Severity: severityError,
			Message:  "TikTok publishing currently requires exactly one video attachment.",
		}}
	}
	if !isVideoMime(media[0].MimeType) {
		return []MediaValidationIssue{{
			Provider: providerTikTok,
			MediaID:  media[0].ID,
			Severity: severityError,
			Message:  "TikTok publishing currently supports video attachments only.",
		}}
	}
	return nil
}

func tiktokScopes() []string {
	return []string{
		"user.info.basic",
		"user.info.profile",
		"user.info.stats",
		"video.list",
		"video.publish",
		"video.upload",
	}
}

type tiktokAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	LogID   string `json:"log_id"`
}

func (e tiktokAPIError) err(label string) error {
	if e.Code == "" || e.Code == "ok" {
		return nil
	}
	message := firstNonEmptyString(e.Message, e.Code)
	if e.LogID != "" {
		return fmt.Errorf("%s: %s (log_id=%s)", label, message, e.LogID)
	}
	return fmt.Errorf("%s: %s", label, message)
}

func tiktokTitle(content string) string {
	title := strings.TrimSpace(content)
	if title == "" {
		return "#OpenPost"
	}
	return truncateRunes(title, tiktokTitleMaxRunes)
}

func truncateRunes(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonEmptyStringSlice(values ...[]string) []string {
	for _, value := range values {
		if len(value) > 0 {
			return value
		}
	}
	return nil
}
