package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	defaultMetaGraphAPIVersion = "v25.0"
	facebookOAuthBaseURL       = "https://www.facebook.com"
	facebookGraphBaseURL       = "https://graph.facebook.com"
)

type FacebookAdapter struct {
	clientID     string
	clientSecret string
	redirectURI  string
	graphVersion string
}

func NewFacebookAdapter(clientID, clientSecret, redirectURI string) *FacebookAdapter {
	return &FacebookAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		graphVersion: metaGraphAPIVersion(),
	}
}

func metaGraphAPIVersion() string {
	if version := strings.TrimSpace(os.Getenv("META_GRAPH_API_VERSION")); version != "" {
		return strings.TrimPrefix(version, "/")
	}
	return defaultMetaGraphAPIVersion
}

func (f *FacebookAdapter) graphURL(path string) string {
	return facebookGraphBaseURL + "/" + f.graphVersion + "/" + strings.TrimPrefix(path, "/")
}

func (f *FacebookAdapter) GenerateAuthURL(state string) (string, map[string]string) {
	params := url.Values{}
	params.Set(oauthParamClientID, f.clientID)
	params.Set(oauthParamRedirectURI, f.redirectURI)
	params.Set("response_type", oauthResponseType)
	params.Set("scope", strings.Join(facebookScopes(), ","))
	params.Set("state", state)
	return facebookOAuthBaseURL + "/" + f.graphVersion + "/dialog/oauth?" + params.Encode(), nil
}

func (f *FacebookAdapter) ExchangeCode(ctx context.Context, code string, _ map[string]string) (*TokenResult, error) {
	return exchangeMetaAuthCode(ctx, f.graphURL, f.clientID, f.clientSecret, f.redirectURI, "facebook", code)
}

func exchangeMetaAuthCode(ctx context.Context, graphURL func(string) string, clientID, clientSecret, redirectURI, providerName, code string) (*TokenResult, error) {
	params := url.Values{}
	params.Set(oauthParamClientID, clientID)
	params.Set(oauthParamClientSecret, clientSecret)
	params.Set(oauthParamRedirectURI, redirectURI)
	params.Set(oauthParamCode, code)

	label := providerName + " token exchange"
	respBody, err := DoRequest(ctx, http.MethodGet, graphURL("oauth/access_token")+"?"+params.Encode(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}

	tokenResp, err := decodeFacebookToken(label, respBody)
	if err != nil {
		return nil, err
	}

	longLived, err := exchangeMetaLongLivedToken(ctx, graphURL, clientID, clientSecret, providerName, tokenResp.AccessToken)
	if err != nil {
		return tokenResp, nil
	}
	return longLived, nil
}

func exchangeMetaLongLivedToken(ctx context.Context, graphURL func(string) string, clientID, clientSecret, providerName, accessToken string) (*TokenResult, error) {
	params := url.Values{}
	params.Set(grantType, "fb_exchange_token")
	params.Set(oauthParamClientID, clientID)
	params.Set(oauthParamClientSecret, clientSecret)
	params.Set("fb_exchange_token", accessToken)

	label := providerName + " long-lived token exchange"
	respBody, err := DoRequest(ctx, http.MethodGet, graphURL("oauth/access_token")+"?"+params.Encode(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	return decodeFacebookToken(label, respBody)
}

func decodeFacebookToken(label string, respBody []byte) (*TokenResult, error) {
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Error       struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", label, err)
	}
	if tokenResp.Error.Message != "" {
		return nil, fmt.Errorf("%s: %s", label, tokenResp.Error.Message)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("%s: missing access token", label)
	}
	return &TokenResult{
		AccessToken: tokenResp.AccessToken,
		ExpiresIn:   tokenResp.ExpiresIn,
		TokenType:   firstNonEmptyString(tokenResp.TokenType, tokenTypeBearer),
	}, nil
}

func (f *FacebookAdapter) RefreshCapability() RefreshCapability {
	return RefreshCapability{}
}

func (f *FacebookAdapter) RefreshToken(_ context.Context, _ RefreshTokenInput) (*TokenResult, error) {
	return nil, fmt.Errorf("facebook page tokens do not support OpenPost refresh yet")
}

func (f *FacebookAdapter) GetProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	respBody, err := DoRequest(ctx, http.MethodGet, f.graphURL("me?fields=id,name&access_token="+url.QueryEscape(accessToken)), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("facebook profile: %w", err)
	}

	var profile struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &profile); err != nil {
		return nil, fmt.Errorf("decoding facebook profile: %w", err)
	}
	if profile.Error.Message != "" {
		return nil, fmt.Errorf("facebook profile: %s", profile.Error.Message)
	}
	if profile.ID == "" {
		return nil, fmt.Errorf("facebook profile: missing id")
	}

	return &UserProfile{
		ID:          profile.ID,
		Username:    profile.Name,
		DisplayName: profile.Name,
	}, nil
}

func (f *FacebookAdapter) ListAccountSelections(ctx context.Context, token *TokenResult) ([]AccountSelectionOption, error) {
	pages, err := f.listPages(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	options := make([]AccountSelectionOption, 0, len(pages))
	for _, page := range pages {
		options = append(options, AccountSelectionOption{
			ID:          page.ID,
			Username:    firstNonEmptyString(page.Username, page.Name),
			DisplayName: page.Name,
			AvatarURL:   page.Picture.Data.URL,
			Kind:        "page",
			Extra: map[string]string{
				"page_id": page.ID,
			},
		})
	}
	return options, nil
}

func (f *FacebookAdapter) SelectAccount(ctx context.Context, token *TokenResult, selectionID string) (*SelectedAccount, error) {
	pages, err := f.listPages(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		if page.ID != selectionID {
			continue
		}
		if page.AccessToken == "" {
			return nil, fmt.Errorf("facebook page %s did not include a page access token", page.ID)
		}
		pageToken := *token
		pageToken.AccessToken = page.AccessToken
		pageToken.RefreshToken = ""
		pageToken.ExpiresIn = 0
		pageToken.Extra = map[string]string{}
		for key, value := range token.Extra {
			pageToken.Extra[key] = value
		}
		pageToken.Extra["page_id"] = page.ID
		pageToken.Extra["page_name"] = page.Name

		return &SelectedAccount{
			AccountID:        page.ID,
			AccountUsername:  firstNonEmptyString(page.Username, page.Name),
			AccountAvatarURL: page.Picture.Data.URL,
			Token:            &pageToken,
		}, nil
	}
	return nil, fmt.Errorf("facebook page selection %s was not found", selectionID)
}

func (f *FacebookAdapter) listPages(ctx context.Context, accessToken string) ([]facebookPage, error) {
	fields := "id,name,username,access_token,picture.type(square)"
	endpoint := f.graphURL("me/accounts") + "?fields=" + url.QueryEscape(fields) + "&access_token=" + url.QueryEscape(accessToken)
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("facebook pages: %w", err)
	}

	var pagesResp struct {
		Data  []facebookPage `json:"data"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &pagesResp); err != nil {
		return nil, fmt.Errorf("decoding facebook pages: %w", err)
	}
	if pagesResp.Error.Message != "" {
		return nil, fmt.Errorf("facebook pages: %s", pagesResp.Error.Message)
	}
	if len(pagesResp.Data) == 0 {
		return nil, fmt.Errorf("facebook account has no manageable pages")
	}
	return pagesResp.Data, nil
}

func (f *FacebookAdapter) UploadMedia(_ context.Context, _ string, _ string, _ string, _ io.Reader) (string, error) {
	return "", fmt.Errorf("facebook uses publicly accessible HTTPS media URLs for the initial adapter")
}

func (f *FacebookAdapter) Publish(ctx context.Context, accessToken, pageID string, req *PublishRequest) (string, error) {
	if req.ReplyToID != "" {
		return "", fmt.Errorf("facebook thread replies are not supported yet")
	}
	switch len(req.PlatformMediaIDs) {
	case 0:
		return f.publishFeedPost(ctx, accessToken, pageID, req.Content)
	case 1:
		if len(req.Media) != 1 {
			return "", fmt.Errorf("facebook media publishing requires media metadata")
		}
		mediaURL := req.PlatformMediaIDs[0]
		if !strings.HasPrefix(mediaURL, "https://") {
			return "", fmt.Errorf("facebook requires a publicly-accessible HTTPS media URL. Set OPENPOST_MEDIA_URL to your public media base URL")
		}
		if isVideoMime(req.Media[0].MimeType) {
			return f.publishVideo(ctx, accessToken, pageID, req.Content, mediaURL)
		}
		return f.publishPhoto(ctx, accessToken, pageID, req.Content, mediaURL)
	default:
		return "", fmt.Errorf("facebook initial adapter supports at most one media attachment")
	}
}

func (f *FacebookAdapter) publishFeedPost(ctx context.Context, accessToken, pageID, message string) (string, error) {
	values := map[string]string{
		"message":             strings.TrimSpace(message),
		oauthParamAccessToken: accessToken,
	}
	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, f.graphURL(pageID+"/feed"), values, nil)
	if err != nil {
		return "", fmt.Errorf("facebook feed publish: %w", err)
	}
	return facebookPublishedID("facebook feed publish", respBody)
}

func (f *FacebookAdapter) publishPhoto(ctx context.Context, accessToken, pageID, caption, mediaURL string) (string, error) {
	values := map[string]string{
		"url":                 mediaURL,
		"caption":             strings.TrimSpace(caption),
		"published":           "true",
		oauthParamAccessToken: accessToken,
	}
	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, f.graphURL(pageID+"/photos"), values, nil)
	if err != nil {
		return "", fmt.Errorf("facebook photo publish: %w", err)
	}
	return facebookPublishedID("facebook photo publish", respBody)
}

func (f *FacebookAdapter) publishVideo(ctx context.Context, accessToken, pageID, description, mediaURL string) (string, error) {
	values := map[string]string{
		"file_url":            mediaURL,
		"description":         strings.TrimSpace(description),
		oauthParamAccessToken: accessToken,
	}
	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, f.graphURL(pageID+"/videos"), values, nil)
	if err != nil {
		return "", fmt.Errorf("facebook video publish: %w", err)
	}
	return facebookPublishedID("facebook video publish", respBody)
}

func facebookPublishedID(label string, respBody []byte) (string, error) {
	var publishResp struct {
		ID     string `json:"id"`
		PostID string `json:"post_id"`
		Error  struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &publishResp); err != nil {
		return "", fmt.Errorf("decoding %s: %w", label, err)
	}
	if publishResp.Error.Message != "" {
		return "", fmt.Errorf("%s: %s", label, publishResp.Error.Message)
	}
	id := firstNonEmptyString(publishResp.PostID, publishResp.ID)
	if id == "" {
		return "", fmt.Errorf("%s: missing published id", label)
	}
	return id, nil
}

func validateFacebookMedia(media []MediaItem) []MediaValidationIssue {
	if len(media) <= 1 {
		return nil
	}
	return []MediaValidationIssue{{
		Provider: providerFacebook,
		Severity: severityError,
		Message:  "Facebook publishing currently supports at most one media attachment.",
	}}
}

func facebookScopes() []string {
	return []string{
		"pages_show_list",
		"pages_read_engagement",
		"pages_manage_posts",
	}
}

type facebookPage struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	AccessToken string `json:"access_token"`
	Picture     struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}
