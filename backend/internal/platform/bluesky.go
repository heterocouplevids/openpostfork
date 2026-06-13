package platform

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type BlueskyAdapter struct {
	pdsURL string
}

func NewBlueskyAdapter(pdsURL string) *BlueskyAdapter {
	if pdsURL == "" {
		pdsURL = "https://bsky.social"
	}
	return &BlueskyAdapter{pdsURL: pdsURL}
}

func (b *BlueskyAdapter) GenerateAuthURL(_ string) (string, map[string]string) {
	return "", nil
}

func (b *BlueskyAdapter) CreateSession(ctx context.Context, handle, appPassword string) (did string, accessToken string, refreshToken string, expiresIn int, err error) {
	payload := map[string]string{
		"identifier": handle,
		"password":   appPassword,
	}

	body, err := jsonMarshal(payload)
	if err != nil {
		return "", "", "", 0, err
	}

	respBody, err := DoRequest(ctx, "POST", b.pdsURL+"/xrpc/com.atproto.server.createSession", bytes.NewReader(body), map[string]string{
		headerContentType: contentTypeJSON,
	})
	if err != nil {
		return "", "", "", 0, fmt.Errorf("bluesky create session: %w", err)
	}

	var session struct {
		Did        string `json:"did"`
		Handle     string `json:"handle"`
		AccessJwt  string `json:"accessJwt"`
		RefreshJwt string `json:"refreshJwt"`
	}
	if err := json.Unmarshal(respBody, &session); err != nil {
		return "", "", "", 0, fmt.Errorf("decoding bluesky session: %w", err)
	}

	expiresIn, err = blueskyJWTExpiresIn(session.AccessJwt)
	if err != nil {
		return "", "", "", 0, err
	}

	return session.Did, session.AccessJwt, session.RefreshJwt, expiresIn, nil
}

func (b *BlueskyAdapter) ExchangeCode(_ context.Context, _ string, _ map[string]string) (*TokenResult, error) {
	return nil, fmt.Errorf("bluesky uses app passwords, not OAuth")
}

func (b *BlueskyAdapter) RefreshCapability() RefreshCapability {
	return RefreshCapability{
		Supported:        true,
		CredentialSource: RefreshCredentialRefreshToken,
	}
}

func (b *BlueskyAdapter) RefreshToken(ctx context.Context, input RefreshTokenInput) (*TokenResult, error) {
	if input.RefreshToken == "" {
		return nil, fmt.Errorf("bluesky refresh requires a refresh token")
	}

	respBody, err := DoRequest(ctx, "POST", b.pdsURL+"/xrpc/com.atproto.server.refreshSession", nil, map[string]string{
		headerAuthorization: bearerPrefix + input.RefreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("bluesky refresh: %w", err)
	}

	var session struct {
		AccessJwt  string `json:"accessJwt"`
		RefreshJwt string `json:"refreshJwt"`
	}
	if err := json.Unmarshal(respBody, &session); err != nil {
		return nil, fmt.Errorf("decoding bluesky refresh: %w", err)
	}

	expiresIn, err := blueskyJWTExpiresIn(session.AccessJwt)
	if err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  session.AccessJwt,
		RefreshToken: session.RefreshJwt,
		ExpiresIn:    expiresIn,
	}, nil
}

func blueskyJWTExpiresIn(token string) (int, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid bluesky jwt format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("decode bluesky jwt payload: %w", err)
	}

	var payload struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return 0, fmt.Errorf("decode bluesky jwt claims: %w", err)
	}
	if payload.Exp == 0 {
		return 0, fmt.Errorf("bluesky jwt missing exp claim")
	}

	expiresIn := int(time.Until(time.Unix(payload.Exp, 0).UTC()).Seconds())
	if expiresIn <= 0 {
		return 0, fmt.Errorf("bluesky jwt already expired")
	}

	return expiresIn, nil
}

func (b *BlueskyAdapter) GetProfile(ctx context.Context, accessToken string) (*UserProfile, error) {
	type blueskySession struct {
		Did    string `json:"did"`
		Handle string `json:"handle"`
	}

	session, err := DoBearerJSON[blueskySession](ctx, "GET", b.pdsURL+"/xrpc/com.atproto.server.getSession", accessToken, nil, "bluesky session")
	if err != nil {
		return nil, err
	}

	return &UserProfile{
		ID:       session.Did,
		Username: session.Handle,
	}, nil
}

func (b *BlueskyAdapter) UploadMedia(ctx context.Context, accessToken, accountID, mimeType string, reader io.Reader) (string, error) {
	if isVideoMime(mimeType) {
		return b.uploadVideo(ctx, accessToken, accountID, mimeType, reader)
	}

	respBody, err := DoRequest(ctx, "POST", b.pdsURL+"/xrpc/com.atproto.repo.uploadBlob", reader, map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
		headerContentType:   mimeType,
	})
	if err != nil {
		return "", fmt.Errorf("bluesky upload blob: %w", err)
	}

	var result struct {
		Blob struct {
			Type string `json:"$type"`
			Ref  struct {
				Link string `json:"$link"`
			} `json:"ref"`
			MimeType string `json:"mimeType"`
			Size     int    `json:"size"`
		} `json:"blob"`
	}
	if unmarshalErr := json.Unmarshal(respBody, &result); unmarshalErr != nil {
		return "", fmt.Errorf("decoding bluesky blob: %w", unmarshalErr)
	}

	blobJSON, err := json.Marshal(result.Blob)
	if err != nil {
		return "", fmt.Errorf("encoding bluesky blob: %w", err)
	}

	return string(blobJSON), nil
}

func (b *BlueskyAdapter) uploadVideo(ctx context.Context, accessToken, did, mimeType string, reader io.Reader) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("reading video data: %w", err)
	}

	serviceToken, err := b.videoServiceAuthToken(ctx, accessToken)
	if err != nil {
		return "", err
	}

	// Derive filename from MIME type
	filename := "video.mp4"
	switch {
	case strings.Contains(mimeType, "quicktime"):
		filename = "video.mov"
	case strings.Contains(mimeType, "webm"):
		filename = "video.webm"
	}

	// Upload video to Bluesky video service
	uploadURL := "https://video.bsky.app/xrpc/app.bsky.video.uploadVideo?did=" + url.QueryEscape(did) + "&name=" + url.QueryEscape(filename)
	jobResp, err := DoRequest(ctx, "POST", uploadURL, bytes.NewReader(data), map[string]string{
		headerAuthorization: bearerPrefix + serviceToken,
		headerContentType:   mimeType,
		"Content-Length":    strconv.Itoa(len(data)),
	})
	if err != nil {
		return "", fmt.Errorf("bluesky video upload: %w", err)
	}

	jobStatus, err := decodeBlueskyVideoJobStatus(jobResp)
	if err != nil {
		return "", fmt.Errorf("decoding bluesky video job: %w", err)
	}
	if jobStatus.State == "JOB_STATE_FAILED" {
		return "", fmt.Errorf("bluesky video processing failed: %s", jobStatus.failureMessage())
	}

	// Poll if video is still processing
	if jobStatus.Blob == nil {
		if jobStatus.JobID == "" {
			return "", fmt.Errorf("bluesky video upload returned no job ID")
		}
		blob, pollErr := b.pollVideoJob(ctx, serviceToken, jobStatus.JobID)
		if pollErr != nil {
			return "", pollErr
		}
		jobStatus.Blob = blob
	}

	blobJSON, err := json.Marshal(jobStatus.Blob)
	if err != nil {
		return "", fmt.Errorf("encoding bluesky video blob: %w", err)
	}

	return string(blobJSON), nil
}

func (b *BlueskyAdapter) videoServiceAuthToken(ctx context.Context, accessToken string) (string, error) {
	audience, err := blueskyServiceAuthAudience(accessToken, b.pdsURL)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("aud", audience)
	params.Set("lxm", "com.atproto.repo.uploadBlob")
	params.Set("exp", strconv.FormatInt(time.Now().UTC().Add(30*time.Minute).Unix(), 10))

	authURL := b.pdsURL + "/xrpc/com.atproto.server.getServiceAuth?" + params.Encode()
	authResp, err := DoRequest(ctx, "GET", authURL, nil, map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
	})
	if err != nil {
		return "", fmt.Errorf("bluesky video service auth: %w", err)
	}

	var authResult struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(authResp, &authResult); err != nil {
		return "", fmt.Errorf("decoding bluesky service auth: %w", err)
	}
	if authResult.Token == "" {
		return "", fmt.Errorf("bluesky service auth returned no token")
	}

	return authResult.Token, nil
}

func blueskyServiceAuthAudience(accessToken, pdsURL string) (string, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return "", fmt.Errorf("decode bluesky jwt payload: %w", err)
		}

		var claims struct {
			Aud json.RawMessage `json:"aud"`
		}
		if err := json.Unmarshal(payloadBytes, &claims); err != nil {
			return "", fmt.Errorf("decode bluesky jwt claims: %w", err)
		}

		var audience string
		if err := json.Unmarshal(claims.Aud, &audience); err == nil && audience != "" {
			return audience, nil
		}

		var audiences []string
		if err := json.Unmarshal(claims.Aud, &audiences); err == nil && len(audiences) > 0 && audiences[0] != "" {
			return audiences[0], nil
		}

		return "", fmt.Errorf("bluesky jwt missing aud claim")
	}

	pdsHost, err := serviceAuthPDSHost(pdsURL)
	if err != nil {
		return "", err
	}
	return "did:web:" + pdsHost, nil
}

func serviceAuthPDSHost(pdsURL string) (string, error) {
	parsed, err := url.Parse(pdsURL)
	if err != nil {
		return "", fmt.Errorf("parsing bluesky PDS URL: %w", err)
	}
	if parsed.Host != "" {
		return parsed.Host, nil
	}

	host := strings.TrimPrefix(strings.TrimPrefix(pdsURL, "https://"), "http://")
	host = strings.TrimRight(host, "/")
	if host == "" {
		return "", fmt.Errorf("bluesky PDS URL has no host")
	}
	return host, nil
}

type blueskyVideoJobStatus struct {
	JobID   string      `json:"jobId"`
	State   string      `json:"state"`
	Blob    interface{} `json:"blob"`
	Error   string      `json:"error"`
	Message string      `json:"message"`
}

func (s blueskyVideoJobStatus) failureMessage() string {
	if s.Message != "" {
		return s.Message
	}
	if s.Error != "" {
		return s.Error
	}
	return s.State
}

func decodeBlueskyVideoJobStatus(data []byte) (blueskyVideoJobStatus, error) {
	var result struct {
		blueskyVideoJobStatus
		JobStatus blueskyVideoJobStatus `json:"jobStatus"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return blueskyVideoJobStatus{}, err
	}

	if result.JobStatus.JobID != "" || result.JobStatus.State != "" || result.JobStatus.Blob != nil {
		return result.JobStatus, nil
	}
	return result.blueskyVideoJobStatus, nil
}

func (b *BlueskyAdapter) pollVideoJob(ctx context.Context, serviceToken, jobID string) (interface{}, error) {
	statusURL := "https://video.bsky.app/xrpc/app.bsky.video.getJobStatus?jobId=" + url.QueryEscape(jobID)

	for i := 0; i < 60; i++ {
		respBody, err := DoRequest(ctx, "GET", statusURL, nil, map[string]string{
			headerAuthorization: bearerPrefix + serviceToken,
		})
		if err != nil {
			return nil, fmt.Errorf("bluesky video job status: %w", err)
		}

		jobStatus, err := decodeBlueskyVideoJobStatus(respBody)
		if err != nil {
			return nil, fmt.Errorf("decoding bluesky video job status: %w", err)
		}

		switch jobStatus.State {
		case "JOB_STATE_COMPLETED":
			if jobStatus.Blob != nil {
				return jobStatus.Blob, nil
			}
		case "JOB_STATE_FAILED":
			return nil, fmt.Errorf("bluesky video processing failed: %s", jobStatus.failureMessage())
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	return nil, fmt.Errorf("bluesky video processing timed out")
}

func (b *BlueskyAdapter) Publish(ctx context.Context, accessToken, accountID string, req *PublishRequest) (string, error) {
	record := map[string]interface{}{
		bskyRecordTypeField: "app.bsky.feed.post",
		jsonFieldText:       req.Content,
		"createdAt":         time.Now().UTC().Format(time.RFC3339Nano),
	}

	if err := b.attachMediaToRecord(record, req); err != nil {
		return "", err
	}

	attachReplyToRecord(record, req.ReplyToID)

	payload := map[string]interface{}{
		"repo":       accountID,
		"collection": "app.bsky.feed.post",
		"record":     record,
	}

	respBody, err := DoJSON(ctx, "POST", b.pdsURL+"/xrpc/com.atproto.repo.createRecord", payload, map[string]string{
		headerAuthorization: bearerPrefix + accessToken,
	})
	if err != nil {
		return "", fmt.Errorf("posting to bluesky: %w", err)
	}

	var result struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decoding bluesky post: %w", err)
	}

	externalID, _ := json.Marshal(map[string]interface{}{
		"uri":   result.URI,
		"cid":   result.CID,
		"_root": getParentRoot(req.ReplyToID),
	})
	return string(externalID), nil
}

func (b *BlueskyAdapter) attachMediaToRecord(record map[string]interface{}, req *PublishRequest) error {
	if len(req.PlatformMediaIDs) == 0 {
		return nil
	}

	isVideo := len(req.Media) > 0 && isVideoMime(req.Media[0].MimeType)
	if isVideo {
		return b.attachVideoToRecord(record, req)
	}

	return b.attachImagesToRecord(record, req)
}

func (b *BlueskyAdapter) attachVideoToRecord(record map[string]interface{}, req *PublishRequest) error {
	var blob map[string]interface{}
	if err := json.Unmarshal([]byte(req.PlatformMediaIDs[0]), &blob); err != nil {
		return fmt.Errorf("decoding bluesky video blob: %w", err)
	}
	altText := ""
	if len(req.MediaAltTexts) > 0 {
		altText = req.MediaAltTexts[0]
	}
	record["embed"] = map[string]interface{}{
		bskyRecordTypeField: "app.bsky.embed.video",
		jsonFieldVideo:      blob,
		"alt":               altText,
	}
	return nil
}

func (b *BlueskyAdapter) attachImagesToRecord(record map[string]interface{}, req *PublishRequest) error {
	images := make([]map[string]interface{}, 0, len(req.PlatformMediaIDs))
	for i, blobJSON := range req.PlatformMediaIDs {
		var blob map[string]interface{}
		if err := json.Unmarshal([]byte(blobJSON), &blob); err != nil {
			return fmt.Errorf("decoding bluesky blob: %w", err)
		}
		altText := ""
		if i < len(req.MediaAltTexts) {
			altText = req.MediaAltTexts[i]
		}
		images = append(images, map[string]interface{}{
			"alt":   altText,
			"image": blob,
		})
	}
	if len(images) > 0 {
		record["embed"] = map[string]interface{}{
			"$type":  "app.bsky.embed.images",
			"images": images,
		}
	}
	return nil
}

func attachReplyToRecord(record map[string]interface{}, replyToID string) {
	if replyToID == "" {
		return
	}

	var parentRef map[string]interface{}
	if err := json.Unmarshal([]byte(replyToID), &parentRef); err != nil {
		return
	}

	rootRef := parentRef
	if rootRef["_root"] != nil {
		if rootMap, ok := rootRef["_root"].(map[string]interface{}); ok {
			rootRef = rootMap
		}
	}

	delete(parentRef, "_root")

	record["reply"] = map[string]interface{}{
		"root":   rootRef,
		"parent": parentRef,
	}
}

func validateBlueskyMedia(media []MediaItem) []MediaValidationIssue {
	if len(media) == 0 {
		return nil
	}

	hasVideo := false
	for _, item := range media {
		if isVideoMime(item.MimeType) {
			if hasVideo {
				return []MediaValidationIssue{{
					Provider: providerBluesky,
					MediaID:  item.ID,
					Severity: severityError,
					Message:  "Bluesky supports only 1 video per post.",
				}}
			}
			hasVideo = true
			if item.MimeType != videoTypeMP4 {
				return []MediaValidationIssue{{
					Provider: providerBluesky,
					MediaID:  item.ID,
					Severity: severityError,
					Message:  "Bluesky supports MP4 video only.",
				}}
			}
			if item.Size > 100*1024*1024 {
				return []MediaValidationIssue{{
					Provider: providerBluesky,
					MediaID:  item.ID,
					Severity: severityError,
					Message:  "Bluesky video must be under 100MB.",
				}}
			}
		}
	}

	if hasVideo {
		if len(media) > 1 {
			return []MediaValidationIssue{{
				Provider: providerBluesky,
				Severity: severityError,
				Message:  "Bluesky does not support mixing video and images in one post.",
			}}
		}
		return nil
	}

	if len(media) > 4 {
		return []MediaValidationIssue{{
			Provider: providerBluesky,
			Severity: severityError,
			Message:  "Bluesky supports up to 4 images per post.",
		}}
	}

	return nil
}

func getParentRoot(replyToID string) interface{} {
	if replyToID == "" {
		return nil
	}
	var parent map[string]interface{}
	if err := json.Unmarshal([]byte(replyToID), &parent); err != nil {
		return nil
	}
	if parent["_root"] != nil {
		return parent["_root"]
	}
	return parent
}
