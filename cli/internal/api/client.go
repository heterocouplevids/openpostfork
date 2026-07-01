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
		if r, ok := body.(io.Reader); ok {
			rdr = r
		} else if s, ok := body.(string); ok && contentType == "" {
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

type Readiness struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

// Ready is the public /api/v1/ready dependency probe.
func (c *Client) Ready(ctx context.Context) (*Readiness, error) {
	var out Readiness
	if err := c.GetJSON(ctx, "/api/v1/ready", &out); err != nil {
		return nil, err
	}
	if out.Status != "ready" {
		return nil, fmt.Errorf("readiness: unexpected status %q", out.Status)
	}
	if out.Database != "ok" {
		return nil, fmt.Errorf("readiness: unexpected database status %q", out.Database)
	}
	return &out, nil
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
	var out WorkspaceSettings
	if err := c.GetJSON(ctx, "/api/v1/workspaces/"+workspaceID+"/settings", &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ----- Billing -----

type BillingStatus struct {
	WorkspaceID       string           `json:"workspace_id"`
	Provider          string           `json:"provider,omitempty"`
	Status            string           `json:"status"`
	PlanID            string           `json:"plan_id,omitempty"`
	CurrentPeriodEnd  string           `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd bool             `json:"cancel_at_period_end"`
	Limits            map[string]int64 `json:"limits"`
	Usage             map[string]int64 `json:"usage"`
	PeriodStart       string           `json:"period_start"`
}

type BillingURL struct {
	URL string `json:"url"`
	ID  string `json:"id,omitempty"`
}

func (c *Client) BillingStatus(ctx context.Context, workspaceID string) (*BillingStatus, error) {
	v := url.Values{}
	v.Set("workspace_id", workspaceID)
	var out BillingStatus
	if err := c.GetJSON(ctx, "/api/v1/billing/status?"+v.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateBillingCheckout(ctx context.Context, workspaceID, planID string) (*BillingURL, error) {
	var out BillingURL
	if err := c.PostJSON(ctx, "/api/v1/billing/checkout", map[string]string{
		"workspace_id": workspaceID,
		"plan_id":      planID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateBillingPortal(ctx context.Context, workspaceID string) (*BillingURL, error) {
	var out BillingURL
	if err := c.PostJSON(ctx, "/api/v1/billing/portal", map[string]string{
		"workspace_id": workspaceID,
	}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ----- Posting schedules -----

type PostingSchedule struct {
	ID             string `json:"id"`
	WorkspaceID    string `json:"workspace_id"`
	SetID          string `json:"set_id,omitempty"`
	UTCHour        int    `json:"utc_hour"`
	UTCMinute      int    `json:"utc_minute"`
	DayOfWeek      int    `json:"day_of_week"`
	LocalHour      int    `json:"local_hour"`
	LocalMinute    int    `json:"local_minute"`
	LocalDayOfWeek int    `json:"local_day_of_week"`
	Label          string `json:"label,omitempty"`
	IsActive       bool   `json:"is_active"`
	CreatedAt      string `json:"created_at"`
}

type NextAvailableSlotInput struct {
	WorkspaceID string
	SetID       string
}

type NextAvailableSlotOutput struct {
	Slot     *PostingSchedule `json:"slot,omitempty"`
	SlotTime string           `json:"slot_time"`
	Message  string           `json:"message"`
}

func (c *Client) NextAvailableSlot(ctx context.Context, in NextAvailableSlotInput) (*NextAvailableSlotOutput, error) {
	v := url.Values{}
	v.Set("workspace_id", in.WorkspaceID)
	if in.SetID != "" {
		v.Set("set_id", in.SetID)
	}
	var out NextAvailableSlotOutput
	if err := c.GetJSON(ctx, "/api/v1/posting-schedules/next-slot?"+v.Encode(), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ----- Accounts -----

type SocialAccount struct {
	ID                     string `json:"id"`
	Slug                   string `json:"slug"`
	Platform               string `json:"platform"`
	AccountID              string `json:"account_id"`
	AccountUsername        string `json:"account_username"`
	InstanceURL            string `json:"instance_url"`
	IsActive               bool   `json:"is_active"`
	ThreadRepliesSupported bool   `json:"thread_replies_supported"`
}

type ProviderInfo struct {
	Platform     string   `json:"platform"`
	DisplayName  string   `json:"display_name"`
	AuthMode     string   `json:"auth_mode"`
	Name         string   `json:"name,omitempty"`
	InstanceURL  string   `json:"instance_url,omitempty"`
	Configured   bool     `json:"configured"`
	Status       string   `json:"status,omitempty"`
	Description  string   `json:"description,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

type UpdateAccountInput struct {
	Slug string `json:"slug"`
}

func (c *Client) ListAccountProviders(ctx context.Context) ([]ProviderInfo, error) {
	var out []ProviderInfo
	if err := c.GetJSON(ctx, "/api/v1/accounts/providers", &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListAccounts(ctx context.Context, workspaceID string) ([]SocialAccount, error) {
	var out []SocialAccount
	if err := c.GetJSON(ctx, "/api/v1/accounts?workspace_id="+url.QueryEscape(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateAccount(ctx context.Context, accountID string, in UpdateAccountInput) (*SocialAccount, error) {
	var out SocialAccount
	if err := c.PatchJSON(ctx, "/api/v1/accounts/"+url.PathEscape(accountID), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DisconnectAccount deactivates a connected social account.
func (c *Client) DisconnectAccount(ctx context.Context, accountID string) error {
	return c.DeleteJSON(ctx, "/api/v1/accounts/"+url.PathEscape(accountID), nil)
}

// ----- Social media sets -----

type SetAccount struct {
	SocialAccountID string `json:"social_account_id"`
	Platform        string `json:"platform"`
	AccountUsername string `json:"account_username"`
	IsMain          bool   `json:"is_main"`
}

type SocialSet struct {
	ID          string       `json:"id"`
	WorkspaceID string       `json:"workspace_id"`
	Name        string       `json:"name"`
	IsDefault   bool         `json:"is_default"`
	CreatedAt   string       `json:"created_at"`
	Accounts    []SetAccount `json:"accounts,omitempty"`
}

type CreateSetInput struct {
	WorkspaceID string   `json:"workspace_id"`
	Name        string   `json:"name"`
	IsDefault   bool     `json:"is_default"`
	AccountIDs  []string `json:"account_ids"`
}

func (c *Client) ListSets(ctx context.Context, workspaceID string) ([]SocialSet, error) {
	var out []SocialSet
	if err := c.GetJSON(ctx, "/api/v1/sets?workspace_id="+url.QueryEscape(workspaceID), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateSet(ctx context.Context, in CreateSetInput) (*SocialSet, error) {
	var out SocialSet
	if err := c.PostJSON(ctx, "/api/v1/sets", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type UpdateSetInput struct {
	Name      *string `json:"name,omitempty"`
	IsDefault *bool   `json:"is_default,omitempty"`
}

func (c *Client) UpdateSet(ctx context.Context, setID string, in UpdateSetInput) (*SocialSet, error) {
	var out SocialSet
	if err := c.PatchJSON(ctx, "/api/v1/sets/"+url.PathEscape(setID), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSet(ctx context.Context, setID string) error {
	return c.DeleteJSON(ctx, "/api/v1/sets/"+url.PathEscape(setID), nil)
}

type AddSetAccountsInput struct {
	AccountIDs []string `json:"account_ids"`
	IsMain     *bool    `json:"is_main,omitempty"`
}

func (c *Client) AddSetAccounts(ctx context.Context, setID string, in AddSetAccountsInput) (*SocialSet, error) {
	var out SocialSet
	if err := c.PostJSON(ctx, "/api/v1/sets/"+url.PathEscape(setID)+"/accounts", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RemoveSetAccount(ctx context.Context, setID, accountID string) (*SocialSet, error) {
	var out SocialSet
	if err := c.DeleteJSON(ctx, "/api/v1/sets/"+url.PathEscape(setID)+"/accounts/"+url.PathEscape(accountID), &out); err != nil {
		return nil, err
	}
	return &out, nil
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

func (c *Client) ListMedia(ctx context.Context, workspaceID string, limit int) ([]MediaListItem, error) {
	v := url.Values{}
	v.Set("workspace_id", workspaceID)
	if limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	var out struct {
		Media []MediaListItem `json:"media"`
		Total int             `json:"total"`
	}
	if err := c.GetJSON(ctx, "/api/v1/media?"+v.Encode(), &out); err != nil {
		return nil, err
	}
	return out.Media, nil
}

// UploadMedia uploads a local file to the active workspace using the legacy multipart media endpoint.
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

// ----- Publications -----

type Publication struct {
	ID              string   `json:"id"`
	WorkspaceID     string   `json:"workspace_id"`
	CreatedBy       string   `json:"created_by"`
	Title           string   `json:"title"`
	SourceContent   string   `json:"source_content"`
	SourceURL       string   `json:"source_url,omitempty"`
	Goal            string   `json:"goal,omitempty"`
	Audience        string   `json:"audience,omitempty"`
	Status          string   `json:"status"`
	ReleasePlanJSON string   `json:"release_plan_json"`
	MediaIDs        []string `json:"media_ids,omitempty"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type CreatePublicationInput struct {
	WorkspaceID   string   `json:"workspace_id"`
	Title         string   `json:"title"`
	SourceContent string   `json:"source_content"`
	SourceURL     string   `json:"source_url,omitempty"`
	Goal          string   `json:"goal,omitempty"`
	Audience      string   `json:"audience,omitempty"`
	MediaIDs      []string `json:"media_ids,omitempty"`
}

func (c *Client) CreatePublication(ctx context.Context, in CreatePublicationInput) (*Publication, error) {
	var out Publication
	if err := c.PostJSON(ctx, "/api/v1/publications", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ListPublicationsInput struct {
	WorkspaceID string
	Status      string
	Limit       int
	Offset      int
}

func (c *Client) ListPublications(ctx context.Context, in ListPublicationsInput) ([]Publication, error) {
	v := url.Values{}
	if in.WorkspaceID != "" {
		v.Set("workspace_id", in.WorkspaceID)
	}
	if in.Status != "" {
		v.Set("status", in.Status)
	}
	if in.Limit > 0 {
		v.Set("limit", strconv.Itoa(in.Limit))
	}
	if in.Offset > 0 {
		v.Set("offset", strconv.Itoa(in.Offset))
	}
	path := "/api/v1/publications"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out []Publication
	if err := c.GetJSON(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetPublication(ctx context.Context, id string) (*Publication, error) {
	var out Publication
	if err := c.GetJSON(ctx, "/api/v1/publications/"+url.PathEscape(id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type UpdatePublicationInput struct {
	Title         *string   `json:"title,omitempty"`
	SourceContent *string   `json:"source_content,omitempty"`
	SourceURL     *string   `json:"source_url,omitempty"`
	Goal          *string   `json:"goal,omitempty"`
	Audience      *string   `json:"audience,omitempty"`
	Status        *string   `json:"status,omitempty"`
	MediaIDs      *[]string `json:"media_ids,omitempty"`
}

func (c *Client) UpdatePublication(ctx context.Context, id string, in UpdatePublicationInput) (*Publication, error) {
	var out Publication
	if err := c.PatchJSON(ctx, "/api/v1/publications/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
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
	PublicationID      string            `json:"publication_id,omitempty"`
	Content            string            `json:"content"`
	Status             string            `json:"status"`
	ScheduledAt        string            `json:"scheduled_at"`
	ActualRunAt        string            `json:"actual_run_at,omitempty"`
	CreatedAt          string            `json:"created_at"`
	RandomDelayMinutes int               `json:"random_delay_minutes"`
	Destinations       []PostDestination `json:"destinations,omitempty"`
	MediaIDs           []string          `json:"media_ids,omitempty"`
	Media              []PostMedia       `json:"media,omitempty"`
	ThreadDraft        *string           `json:"thread_draft,omitempty"`
}

type CreatePostInput struct {
	WorkspaceID        string     `json:"workspace_id"`
	Content            string     `json:"content"`
	ScheduledAt        *time.Time `json:"scheduled_at,omitempty"`
	SocialAccountIDs   []string   `json:"social_account_ids"`
	MediaIDs           []string   `json:"media_ids,omitempty"`
	PublicationID      string     `json:"publication_id,omitempty"`
	RandomDelayMinutes int        `json:"random_delay_minutes,omitempty"`
	ThreadDraft        *string    `json:"thread_draft,omitempty"`
}

func (c *Client) CreatePost(ctx context.Context, in CreatePostInput) (*Post, error) {
	var out Post
	if err := c.PostJSON(ctx, "/api/v1/posts", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type ListPostsInput struct {
	WorkspaceID string
	Status      string
	Date        string
	Limit       int
	Offset      int
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
	if in.Offset > 0 {
		v.Set("offset", strconv.Itoa(in.Offset))
	}
	path := "/api/v1/posts"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out []Post
	if err := c.GetJSON(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetPost(ctx context.Context, id string) (*Post, error) {
	var out Post
	if err := c.GetJSON(ctx, "/api/v1/posts/"+url.PathEscape(id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeletePost(ctx context.Context, id string) error {
	return c.DeleteJSON(ctx, "/api/v1/posts/"+url.PathEscape(id), nil)
}

type UpdatePostInput struct {
	Content            *string  `json:"content,omitempty"`
	ScheduledAt        *string  `json:"scheduled_at,omitempty"`
	SocialAccountIDs   []string `json:"social_account_ids,omitempty"`
	MediaIDs           []string `json:"media_ids,omitempty"`
	PublicationID      *string  `json:"publication_id,omitempty"`
	RandomDelayMinutes *int     `json:"random_delay_minutes,omitempty"`
	ThreadDraft        *string  `json:"thread_draft,omitempty"`
}

func (c *Client) UpdatePost(ctx context.Context, id string, in UpdatePostInput) (*Post, error) {
	var out Post
	if err := c.PatchJSON(ctx, "/api/v1/posts/"+url.PathEscape(id), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type PostMedia struct {
	MediaID      string `json:"media_id"`
	DisplayOrder int    `json:"display_order"`
	FilePath     string `json:"file_path"`
	MimeType     string `json:"mime_type"`
	AltText      string `json:"alt_text"`
}

type ThreadPostInput struct {
	Content  string   `json:"content"`
	MediaIDs []string `json:"media_ids,omitempty"`
}

type CreateThreadInput struct {
	WorkspaceID        string            `json:"workspace_id"`
	Posts              []ThreadPostInput `json:"posts"`
	ScheduledAt        *time.Time        `json:"scheduled_at,omitempty"`
	SocialAccountIDs   []string          `json:"social_account_ids"`
	PublicationID      string            `json:"publication_id,omitempty"`
	RandomDelayMinutes int               `json:"random_delay_minutes,omitempty"`
}

type CreateThreadOutput struct {
	PostIDs []string `json:"post_ids"`
}

func (c *Client) CreateThread(ctx context.Context, in CreateThreadInput) (*CreateThreadOutput, error) {
	var out CreateThreadOutput
	if err := c.PostJSON(ctx, "/api/v1/posts/thread", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
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
	RawToken string `json:"token"`
	Item     Token  `json:"item"`
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
	ID          string `json:"id"`
	Type        string `json:"type"`
	Payload     string `json:"payload"`
	Status      string `json:"status"`
	RunAt       string `json:"run_at"`
	Attempts    int    `json:"attempts"`
	MaxAttempts int    `json:"max_attempts"`
	LastError   string `json:"last_error,omitempty"`
}

type ListJobsInput struct {
	Status      string
	Limit       int
	Offset      int
	WorkspaceID string
}

func (c *Client) ListJobs(ctx context.Context, in ListJobsInput) ([]Job, error) {
	v := url.Values{}
	if in.Status != "" {
		v.Set("status", in.Status)
	}
	if in.Limit > 0 {
		v.Set("limit", strconv.Itoa(in.Limit))
	}
	if in.Offset > 0 {
		v.Set("offset", strconv.Itoa(in.Offset))
	}
	if in.WorkspaceID != "" {
		v.Set("workspace_id", in.WorkspaceID)
	}
	path := "/api/v1/jobs"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out []Job
	if err := c.GetJSON(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
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
