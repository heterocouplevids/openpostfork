package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

const (
	googleOAuthURL          = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL          = "https://oauth2.googleapis.com/token"
	googleUserInfoURL       = "https://www.googleapis.com/oauth2/v2/userinfo"
	youtubeAPIBaseURL       = "https://www.googleapis.com/youtube/v3"
	youtubeUploadBaseURL    = "https://www.googleapis.com/upload/youtube/v3"
	youtubeDefaultVideoName = "OpenPost video"
	youtubeTitleMaxRunes    = 100
)

type YouTubeAdapter struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewYouTubeAdapter(clientID, clientSecret, redirectURI string) *YouTubeAdapter {
	return &YouTubeAdapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
	}
}

func (y *YouTubeAdapter) GenerateAuthURL(state string) (string, map[string]string) {
	params := url.Values{}
	params.Set(oauthParamClientID, y.clientID)
	params.Set(oauthParamRedirectURI, y.redirectURI)
	params.Set("response_type", oauthResponseType)
	params.Set("scope", strings.Join(youtubeScopes(), " "))
	params.Set("state", state)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")
	params.Set("include_granted_scopes", "true")
	return googleOAuthURL + "?" + params.Encode(), nil
}

func (y *YouTubeAdapter) ExchangeCode(ctx context.Context, code string, _ map[string]string) (*TokenResult, error) {
	values := map[string]string{
		oauthParamClientID:     y.clientID,
		oauthParamClientSecret: y.clientSecret,
		oauthParamCode:         code,
		oauthParamRedirectURI:  y.redirectURI,
		grantType:              oauthGrantAuthCode,
	}
	return y.exchangeToken(ctx, values, "youtube token exchange")
}

func (y *YouTubeAdapter) RefreshCapability() RefreshCapability {
	return RefreshCapability{
		Supported:        true,
		CredentialSource: RefreshCredentialRefreshToken,
	}
}

func (y *YouTubeAdapter) RefreshToken(ctx context.Context, input RefreshTokenInput) (*TokenResult, error) {
	if input.RefreshToken == "" {
		return nil, fmt.Errorf("youtube refresh requires a refresh token")
	}
	values := map[string]string{
		oauthParamClientID:                    y.clientID,
		oauthParamClientSecret:                y.clientSecret,
		grantType:                             oauthGrantRefresh,
		string(RefreshCredentialRefreshToken): input.RefreshToken,
	}
	return y.exchangeToken(ctx, values, "youtube token refresh")
}

func (y *YouTubeAdapter) exchangeToken(ctx context.Context, values map[string]string, label string) (*TokenResult, error) {
	respBody, err := DoFormURLEncoded(ctx, http.MethodPost, googleTokenURL, values, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
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

func (y *YouTubeAdapter) GetProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	respBody, err := DoRequest(ctx, http.MethodGet, googleUserInfoURL, nil, bearerHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("youtube google profile: %w", err)
	}

	var profile struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Email   string `json:"email"`
		Error   struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &profile); err != nil {
		return nil, fmt.Errorf("decoding youtube google profile: %w", err)
	}
	if profile.Error.Message != "" {
		return nil, fmt.Errorf("youtube google profile: %s", profile.Error.Message)
	}
	return &UserProfile{
		ID:          profile.ID,
		Username:    firstNonEmptyString(profile.Email, profile.Name, profile.ID),
		DisplayName: firstNonEmptyString(profile.Name, profile.Email, profile.ID),
	}, nil
}

func (y *YouTubeAdapter) ListAccountSelections(ctx context.Context, token *TokenResult) ([]AccountSelectionOption, error) {
	channels, err := y.listChannels(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	options := make([]AccountSelectionOption, 0, len(channels))
	for _, channel := range channels {
		options = append(options, AccountSelectionOption{
			ID:          channel.ID,
			Username:    firstNonEmptyString(channel.Snippet.CustomURL, channel.Snippet.Title, channel.ID),
			DisplayName: channel.Snippet.Title,
			AvatarURL:   channel.Snippet.Thumbnails.Default.URL,
			Description: youtubeSubscriberDescription(channel.Statistics.SubscriberCount),
			Kind:        "channel",
		})
	}
	return options, nil
}

func (y *YouTubeAdapter) SelectAccount(ctx context.Context, token *TokenResult, selectionID string) (*SelectedAccount, error) {
	channels, err := y.listChannels(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		if channel.ID != selectionID {
			continue
		}
		selectedToken := *token
		selectedToken.Extra = map[string]string{}
		for key, value := range token.Extra {
			selectedToken.Extra[key] = value
		}
		selectedToken.Extra["channel_id"] = channel.ID

		return &SelectedAccount{
			AccountID:        channel.ID,
			AccountUsername:  firstNonEmptyString(channel.Snippet.CustomURL, channel.Snippet.Title, channel.ID),
			AccountAvatarURL: channel.Snippet.Thumbnails.Default.URL,
			Token:            &selectedToken,
		}, nil
	}
	return nil, fmt.Errorf("youtube channel selection %s was not found", selectionID)
}

func (y *YouTubeAdapter) listChannels(ctx context.Context, accessToken string) ([]youtubeChannel, error) {
	params := url.Values{}
	params.Set("part", "snippet,statistics")
	params.Set("mine", "true")
	params.Set("maxResults", "50")
	endpoint := youtubeAPIBaseURL + "/channels?" + params.Encode()
	respBody, err := DoRequest(ctx, http.MethodGet, endpoint, nil, bearerHeaders(accessToken))
	if err != nil {
		return nil, fmt.Errorf("youtube channels: %w", err)
	}

	var channelsResp struct {
		Items []youtubeChannel `json:"items"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &channelsResp); err != nil {
		return nil, fmt.Errorf("decoding youtube channels: %w", err)
	}
	if channelsResp.Error.Message != "" {
		return nil, fmt.Errorf("youtube channels: %s", channelsResp.Error.Message)
	}
	if len(channelsResp.Items) == 0 {
		return nil, fmt.Errorf("google account has no YouTube channels")
	}
	return channelsResp.Items, nil
}

func (y *YouTubeAdapter) UploadMedia(_ context.Context, _ string, _ string, _ string, _ io.Reader) (string, error) {
	return "", fmt.Errorf("youtube video upload requires post metadata")
}

func (y *YouTubeAdapter) UploadMediaWithMetadata(ctx context.Context, accessToken, _ string, req UploadMediaRequest) (string, error) {
	if req.Reader == nil {
		return "", fmt.Errorf("youtube upload requires a video reader")
	}
	if !isVideoMime(req.MimeType) {
		return "", fmt.Errorf("youtube upload requires a video attachment")
	}

	metadata := youtubeVideoInsertRequest{
		Snippet: youtubeVideoSnippet{
			Title:       youtubeTitle(req),
			Description: strings.TrimSpace(req.Description),
		},
		Status: youtubeVideoStatus{PrivacyStatus: "private"},
	}

	body, contentType := youtubeMultipartBody(metadata, req.MimeType, req.Reader)

	params := url.Values{}
	params.Set("part", "snippet,status")
	params.Set("uploadType", "multipart")
	params.Set("notifySubscribers", "false")
	endpoint := youtubeUploadBaseURL + "/videos?" + params.Encode()
	respBody, err := DoRequest(ctx, http.MethodPost, endpoint, body, map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
		headerContentType:   contentType,
	})
	if err != nil {
		return "", fmt.Errorf("youtube video upload: %w", err)
	}

	return youtubeVideoIDFromResponse("youtube video upload", respBody)
}

func (y *YouTubeAdapter) Publish(_ context.Context, _ string, _ string, req *PublishRequest) (string, error) {
	if req.ReplyToID != "" {
		return "", fmt.Errorf("youtube thread replies are not supported")
	}
	if len(req.PlatformMediaIDs) != 1 || len(req.Media) != 1 {
		return "", fmt.Errorf("youtube publishing requires exactly one video attachment")
	}
	if !isVideoMime(req.Media[0].MimeType) {
		return "", fmt.Errorf("youtube publishing requires a video attachment")
	}
	return req.PlatformMediaIDs[0], nil
}

func youtubeMultipartBody(metadata youtubeVideoInsertRequest, mimeType string, reader io.Reader) (io.Reader, string) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	contentType := "multipart/related; boundary=" + writer.Boundary()

	go func() {
		err := writeYouTubeMultipart(writer, metadata, mimeType, reader)
		if closeErr := writer.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_ = pw.Close()
	}()

	return pr, contentType
}

func writeYouTubeMultipart(writer *multipart.Writer, metadata youtubeVideoInsertRequest, mimeType string, reader io.Reader) error {
	jsonPart, err := writer.CreatePart(textproto.MIMEHeader{
		headerContentType: []string{contentTypeJSON + "; charset=UTF-8"},
	})
	if err != nil {
		return fmt.Errorf("creating youtube metadata part: %w", err)
	}
	metaBytes, err := jsonMarshal(metadata)
	if err != nil {
		return fmt.Errorf("marshaling youtube metadata: %w", err)
	}
	if _, err := jsonPart.Write(metaBytes); err != nil {
		return fmt.Errorf("writing youtube metadata: %w", err)
	}

	mediaPart, err := writer.CreatePart(textproto.MIMEHeader{
		headerContentType: []string{firstNonEmptyString(mimeType, videoTypeMP4)},
	})
	if err != nil {
		return fmt.Errorf("creating youtube media part: %w", err)
	}
	if _, err := io.Copy(mediaPart, reader); err != nil {
		return fmt.Errorf("copying youtube media: %w", err)
	}
	return nil
}

func youtubeVideoIDFromResponse(label string, respBody []byte) (string, error) {
	var resp struct {
		ID    string `json:"id"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("decoding %s: %w", label, err)
	}
	if resp.Error.Message != "" {
		return "", fmt.Errorf("%s: %s", label, resp.Error.Message)
	}
	if resp.ID == "" {
		return "", fmt.Errorf("%s: missing video id", label)
	}
	return resp.ID, nil
}

func validateYouTubeMedia(media []MediaItem) []MediaValidationIssue {
	if len(media) != 1 {
		return []MediaValidationIssue{{
			Provider: providerYouTube,
			Severity: severityError,
			Message:  "YouTube publishing currently requires exactly one video attachment.",
		}}
	}
	if !isVideoMime(media[0].MimeType) {
		return []MediaValidationIssue{{
			Provider: providerYouTube,
			MediaID:  media[0].ID,
			Severity: severityError,
			Message:  "YouTube publishing supports video attachments only.",
		}}
	}
	return nil
}

func youtubeTitle(req UploadMediaRequest) string {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		for _, line := range strings.Split(req.Description, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				title = trimmed
				break
			}
		}
	}
	if title == "" {
		title = youtubeDefaultVideoName
	}
	return truncateRunes(title, youtubeTitleMaxRunes)
}

func youtubeSubscriberDescription(count string) string {
	if strings.TrimSpace(count) == "" {
		return ""
	}
	return count + " subscribers"
}

func youtubeScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/youtube.readonly",
		"https://www.googleapis.com/auth/youtube.upload",
	}
}

func bearerHeaders(accessToken string) map[string]string {
	return map[string]string{headerAuthorization: bearerPrefix + accessToken}
}

type youtubeVideoInsertRequest struct {
	Snippet youtubeVideoSnippet `json:"snippet"`
	Status  youtubeVideoStatus  `json:"status"`
}

type youtubeVideoSnippet struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type youtubeVideoStatus struct {
	PrivacyStatus string `json:"privacyStatus"`
}

type youtubeChannel struct {
	ID      string `json:"id"`
	Snippet struct {
		Title      string `json:"title"`
		CustomURL  string `json:"customUrl"`
		Thumbnails struct {
			Default struct {
				URL string `json:"url"`
			} `json:"default"`
		} `json:"thumbnails"`
	} `json:"snippet"`
	Statistics struct {
		SubscriberCount string `json:"subscriberCount"`
	} `json:"statistics"`
}
