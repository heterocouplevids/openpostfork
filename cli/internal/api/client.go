// Package api is the OpenPost HTTP client used by every CLI
// subcommand. It is a hand-rolled typed client over the same
// /api/v1 surface the web frontend consumes. The CLI is treated as
// an external client of a running OpenPost instance; it does not
// import backend/internal/... or touch SQLite.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Client is a thin wrapper around *http.Client with the bits every
// OpenPost call needs: base URL, bearer token, JSON helpers, multipart
// upload, and a typed error surface.
type Client struct {
	BaseURL   string
	Token     string
	HTTP      *http.Client
	UserAgent string
}

// New returns a client targeting baseURL with the given bearer token.
// Pass an empty token for unauthenticated calls (e.g. /health).
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		Token:     token,
		HTTP:      &http.Client{Timeout: 60 * time.Second},
		UserAgent: "openpost-cli/0.1.0",
	}
}

// Error is the wire-format error returned by Huma: { "error": "..." }.
// The CLI maps this to a friendly stderr line.
type Error struct {
	StatusCode int    `json:"-"`
	Title      string `json:"title"`
	Detail     string `json:"detail"`
	Message    string `json:"message"`
}

func (e *Error) Error() string {
	for _, s := range []string{e.Detail, e.Message, e.Title} {
		if s != "" {
			return fmt.Sprintf("HTTP %d: %s", e.StatusCode, s)
		}
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any, contentType string) error {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return fmt.Errorf("invalid URL %s: %w", c.BaseURL+path, err)
	}
	var rdr io.Reader
	if body != nil {
		if s, ok := body.(string); ok && contentType == "" {
			rdr = strings.NewReader(s)
		} else {
			data, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("marshal request: %w", err)
			}
			rdr = bytes.NewReader(data)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), rdr)
	if err != nil {
		return err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &Error{StatusCode: resp.StatusCode}
		_ = json.Unmarshal(respBody, apiErr)
		if apiErr.Message == "" && apiErr.Detail == "" {
			apiErr.Message = strings.TrimSpace(string(respBody))
		}
		return apiErr
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out, "")
}

func (c *Client) PostJSON(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out, "")
}

func (c *Client) PatchJSON(ctx context.Context, path string, body, out any) error {
	return c.do(ctx, http.MethodPatch, path, body, out, "")
}

func (c *Client) DeleteJSON(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodDelete, path, nil, out, "")
}

// PostForm posts a multipart/form-data request with the given fields
// and a file under fieldName. out is decoded from the JSON body.
func (c *Client) PostForm(ctx context.Context, path, fileField, filePath string, fields map[string]string, out any) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return err
		}
	}
	fw, err := mw.CreateFormFile(fileField, filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}
	return c.do(ctx, http.MethodPost, path, &buf, out, mw.FormDataContentType())
}

// ----- typed endpoints used by the CLI -----

// Health is the public /api/v1/health probe.
func (c *Client) Health(ctx context.Context) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.GetJSON(ctx, "/api/v1/health", &out); err != nil {
		return err
	}
	if out.Status != "ok" {
		return fmt.Errorf("health: unexpected status %q", out.Status)
	}
	return nil
}

// Me fetches the authenticated user profile.
func (c *Client) Me(ctx context.Context) (*Me, error) {
	var m Me
	if err := c.GetJSON(ctx, "/api/v1/auth/me", &m); err != nil {
		return nil, err
	}
	return &m, nil
}

type Me struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// ----- Workspaces -----

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Client) ListWorkspaces(ctx context.Context) ([]Workspace, error) {
	var out []Workspace
	if err := c.GetJSON(ctx, "/api/v1/workspaces", &out); err != nil {
		return nil, err
	}
	return out, nil
}

type CreateWorkspaceInput struct {
	Name string `json:"name"`
}

func (c *Client) CreateWorkspace(ctx context.Context, in CreateWorkspaceInput) (*Workspace, error) {
	var out Workspace
	if err := c.PostJSON(ctx, "/api/v1/workspaces", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type WorkspaceSettings struct {
	Timezone            string `json:"timezone"`
	WeekStart           int    `json:"week_start"`
	MediaCleanupDays    int    `json:"media_cleanup_days"`
	RandomDelayMinutes  int    `json:"random_delay_minutes"`
	DraftGapMinutes     int    `json:"draft_gap_minutes"`
	SlotStartHour       int    `json:"slot_start_hour"`
	SlotEndHour         int    `json:"slot_end_hour"`
	SlotIntervalMinutes int    `json:"slot_interval_minutes"`
}

func (c *Client) GetWorkspaceSettings(ctx context.Context, workspaceID string) (*WorkspaceSettings, error) {
	var out struct {
		Body WorkspaceSettings `json:"body"`
	}
	if err := c.GetJSON(ctx, "/api/v1/workspaces/"+workspaceID+"/settings", &out); err != nil {
		return nil, err
	}
	return &out.Body, nil
}

// ----- Accounts -----

type SocialAccount struct {
	ID                     string `json:"id"`
	Platform               string `json:"platform"`
	AccountID              string `json:"account_id"`
	AccountUsername        string `json:"account_username"`
	InstanceURL            string `json:"instance_url"`
	IsActive               bool   `json:"is_active"`
	ThreadRepliesSupported bool   `json:"thread_replies_supported"`
}

func (c *Client) ListAccounts(ctx context.Context, workspaceID string) ([]SocialAccount, error) {
	var out struct {
		Body []SocialAccount `json:"body"`
	}
	if err := c.GetJSON(ctx, "/api/v1/accounts?workspace_id="+url.QueryEscape(workspaceID), &out); err != nil {
		return nil, err
	}
	return out.Body, nil
}

// ----- Media -----

type Media struct {
	ID               string `json:"id"`
	MimeType         string `json:"mime_type"`
	URL              string `json:"url"`
	Size             int64  `json:"size"`
	Deduped          bool   `json:"deduped"`
	AltText          string `json:"alt_text"`
	OriginalFilename string `json:"original_filename"`
}

type MediaListItem struct {
	ID               string `json:"id"`
	WorkspaceID      string `json:"workspace_id"`
	MimeType         string `json:"mime_type"`
	Size             int64  `json:"size"`
	OriginalFilename string `json:"original_filename"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	AltText          string `json:"alt_text"`
	IsFavorite       bool   `json:"is_favorite"`
	CreatedAt        string `json:"created_at"`
	URL              string `json:"url"`
	ThumbnailURL     string `json:"thumbnail_url"`
	UsageCount       int    `json:"usage_count"`
	CanDelete        bool   `json:"can_delete"`
	ProcessingStatus string `json:"processing_status"`
}

func (c *Client) ListMedia(ctx context.Context, workspaceID string) ([]MediaListItem, error) {
	var out struct {
		Body struct {
			Media []MediaListItem `json:"media"`
			Total int             `json:"total"`
		} `json:"body"`
	}
	if err := c.GetJSON(ctx, "/api/v1/media?workspace_id="+url.QueryEscape(workspaceID), &out); err != nil {
		return nil, err
	}
	return out.Body.Media, nil
}

func (c *Client) UploadMedia(ctx context.Context, workspaceID, filePath, altText string) (*Media, error) {
	fields := map[string]string{"workspace_id": workspaceID}
	if altText != "" {
		fields["alt_text"] = altText
	}
	var m Media
	if err := c.PostForm(ctx, "/api/v1/media/upload", "file", filePath, fields, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *Client) DeleteMedia(ctx context.Context, mediaID string) error {
	return c.DeleteJSON(ctx, "/api/v1/media/"+mediaID, nil)
}

// ----- Posts -----

type PostDestination struct {
	SocialAccountID string `json:"social_account_id"`
	Platform        string `json:"platform"`
	Status          string `json:"status"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type Post struct {
	ID                 string            `json:"id"`
	WorkspaceID        string            `json:"workspace_id"`
	CreatedBy          string            `json:"created_by"`
	Content            string            `json:"content"`
	Status             string            `json:"status"`
	ScheduledAt        string            `json:"scheduled_at"`
	ActualRunAt        string            `json:"actual_run_at,omitempty"`
	CreatedAt          string            `json:"created_at"`
	RandomDelayMinutes int               `json:"random_delay_minutes"`
	Destinations       []PostDestination `json:"destinations,omitempty"`
	MediaIDs           []string          `json:"media_ids,omitempty"`
}

type CreatePostInput struct {
	WorkspaceID        string     `json:"workspace_id"`
	Content            string     `json:"content"`
	ScheduledAt        *time.Time `json:"scheduled_at,omitempty"`
	SocialAccountIDs   []string   `json:"social_account_ids"`
	MediaIDs           []string   `json:"media_ids,omitempty"`
	RandomDelayMinutes int        `json:"random_delay_minutes,omitempty"`
}

func (c *Client) CreatePost(ctx context.Context, in CreatePostInput) (*Post, error) {
	var out struct {
		Body Post `json:"body"`
	}
	if err := c.PostJSON(ctx, "/api/v1/posts", in, &out); err != nil {
		return nil, err
	}
	return &out.Body, nil
}

type ListPostsInput struct {
	WorkspaceID string
	Status      string
	Date        string
	Limit       int
}

func (c *Client) ListPosts(ctx context.Context, in ListPostsInput) ([]Post, error) {
	v := url.Values{}
	if in.WorkspaceID != "" {
		v.Set("workspace_id", in.WorkspaceID)
	}
	if in.Status != "" {
		v.Set("status", in.Status)
	}
	if in.Date != "" {
		v.Set("date", in.Date)
	}
	if in.Limit > 0 {
		v.Set("limit", strconv.Itoa(in.Limit))
	}
	path := "/api/v1/posts"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out struct {
		Body []Post `json:"body"`
	}
	if err := c.GetJSON(ctx, path, &out); err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (c *Client) GetPost(ctx context.Context, id string) (*Post, error) {
	var out struct {
		Body Post `json:"body"`
	}
	if err := c.GetJSON(ctx, "/api/v1/posts/"+id, &out); err != nil {
		return nil, err
	}
	return &out.Body, nil
}

func (c *Client) DeletePost(ctx context.Context, id string) error {
	return c.DeleteJSON(ctx, "/api/v1/posts/"+id, nil)
}

type CreateThreadInput struct {
	WorkspaceID        string     `json:"workspace_id"`
	Posts              []string   `json:"posts"`
	ScheduledAt        *time.Time `json:"scheduled_at,omitempty"`
	SocialAccountIDs   []string   `json:"social_account_ids"`
	MediaIDs           []string   `json:"media_ids,omitempty"`
	RandomDelayMinutes int        `json:"random_delay_minutes,omitempty"`
}

func (c *Client) CreateThread(ctx context.Context, in CreateThreadInput) (*Post, error) {
	var out struct {
		Body Post `json:"body"`
	}
	if err := c.PostJSON(ctx, "/api/v1/posts/thread", in, &out); err != nil {
		return nil, err
	}
	return &out.Body, nil
}

// ----- Auth: CLI device flow + API token management -----

type CLIAuthStartInput struct {
	ClientName      string `json:"client_name"`
	ClientVersion   string `json:"client_version"`
	ClientOS        string `json:"client_os"`
	RequestedScopes string `json:"requested_scopes"`
}

type CLIAuthStartOutput struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// StartCLIAuth begins a CLI device-flow session.
func (c *Client) StartCLIAuth(ctx context.Context, in CLIAuthStartInput) (*CLIAuthStartOutput, error) {
	var out CLIAuthStartOutput
	if err := c.PostJSON(ctx, "/api/v1/cli/auth/start", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type CLIAuthPollInput struct {
	DeviceCode string `json:"device_code"`
}

type CLIAuthPollOutput struct {
	Status    string    `json:"status"` // authorization_pending, approved, access_denied, expired_token
	Token     string    `json:"token,omitempty"`
	TokenID   string    `json:"token_id,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// PollCLIAuth reports the current status of a pending CLI device-flow.
func (c *Client) PollCLIAuth(ctx context.Context, deviceCode string) (*CLIAuthPollOutput, error) {
	var out CLIAuthPollOutput
	if err := c.PostJSON(ctx, "/api/v1/cli/auth/poll", CLIAuthPollInput{DeviceCode: deviceCode}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type Token struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Scope       string     `json:"scope"`
	WorkspaceID string     `json:"workspace_id"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type CreateAPITokenInput struct {
	Name        string     `json:"name"`
	Scope       string     `json:"scope"`
	WorkspaceID string     `json:"workspace_id,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type CreateAPITokenOutput struct {
	Token
	RawToken string `json:"token"`
}

func (c *Client) ListAPITokens(ctx context.Context) ([]Token, error) {
	var out []Token
	if err := c.GetJSON(ctx, "/api/v1/api-tokens", &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateAPIToken(ctx context.Context, in CreateAPITokenInput) (*CreateAPITokenOutput, error) {
	var out CreateAPITokenOutput
	if err := c.PostJSON(ctx, "/api/v1/api-tokens", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RevokeAPIToken(ctx context.Context, id string) error {
	return c.DeleteJSON(ctx, "/api/v1/api-tokens/"+id, nil)
}

// ----- Jobs -----

type Job struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Payload     string    `json:"payload"`
	Status      string    `json:"status"`
	RunAt       time.Time `json:"run_at"`
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"max_attempts"`
	LastError   string    `json:"last_error,omitempty"`
}

func (c *Client) ListJobs(ctx context.Context) ([]Job, error) {
	var out struct {
		Body []Job `json:"body"`
	}
	if err := c.GetJSON(ctx, "/api/v1/jobs", &out); err != nil {
		return nil, err
	}
	return out.Body, nil
}

// ----- helpers -----

// ErrAuthRequired is returned when a call needs a token and none is set.
var ErrAuthRequired = errors.New("not logged in: run `openpost auth login <instance>` or set OPENPOST_TOKEN")

// CheckToken returns ErrAuthRequired if c.Token is empty. Subcommands
// that always need auth (post create, account list, ...) call this
// before doing any work.
func (c *Client) CheckToken() error {
	if c.Token == "" {
		return ErrAuthRequired
	}
	return nil
}
