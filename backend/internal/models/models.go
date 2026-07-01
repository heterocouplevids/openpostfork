package models

import (
	"time"

	"github.com/uptrace/bun"
)

// Post status values stored in the `posts.status` column.
const (
	PostStatusDraft      = "draft"
	PostStatusScheduled  = "scheduled"
	PostStatusPublishing = "publishing"
	PostStatusPublished  = "published"
	PostStatusFailed     = "failed"
)

// Publication status values stored in the `publications.status` column.
const (
	PublicationStatusDraft     = "draft"
	PublicationStatusReady     = "ready"
	PublicationStatusScheduled = "scheduled"
	PublicationStatusPublished = "published"
	PublicationStatusFailed    = "failed"
)

// Workspace role values stored in the `workspace_members.role` column.
const (
	WorkspaceRoleAdmin  = "admin"
	WorkspaceRoleEditor = "editor"
	WorkspaceRoleViewer = "viewer"
)

type Workspace struct {
	bun.BaseModel `bun:"table:workspaces"`

	ID                  string    `bun:",pk" json:"id"`
	Name                string    `bun:",notnull" json:"name"`
	Timezone            string    `bun:",default:'UTC'" json:"timezone"`
	WeekStart           int       `bun:",default:1" json:"week_start"`             // 0=Sunday, 1=Monday
	MediaCleanupDays    int       `bun:",default:0" json:"media_cleanup_days"`     // 0 = disabled
	RandomDelayMinutes  int       `bun:",default:0" json:"random_delay_minutes"`   // ±N minutes natural posting
	DraftGapMinutes     int       `bun:",default:60" json:"draft_gap_minutes"`     // Minimum gap when spilling past configured schedule slots
	SlotStartHour       int       `bun:",default:5" json:"slot_start_hour"`        // 0-23
	SlotEndHour         int       `bun:",default:23" json:"slot_end_hour"`         // 0-23
	SlotIntervalMinutes int       `bun:",default:15" json:"slot_interval_minutes"` // 1-180
	CreatedAt           time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID               string    `bun:",pk" json:"id"`
	Email            string    `bun:",unique,notnull" json:"email"`
	PasswordHash     string    `bun:",notnull" json:"-"`
	IsAdmin          bool      `bun:",notnull,default:false" json:"is_admin"`
	TOTPSecretEnc    []byte    `bun:"totp_secret_encrypted" json:"-"`
	TOTPEnabledAt    time.Time `bun:",nullzero" json:"totp_enabled_at"`
	PasskeyEnabledAt time.Time `bun:",nullzero" json:"passkey_enabled_at"`
	CreatedAt        time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type UserPasskey struct {
	bun.BaseModel `bun:"table:user_passkeys"`

	ID             string    `bun:",pk" json:"id"`
	UserID         string    `bun:",notnull" json:"user_id"`
	Name           string    `bun:",notnull" json:"name"`
	CredentialID   []byte    `bun:",notnull,unique" json:"-"`
	CredentialJSON string    `bun:"credential_json,notnull" json:"-"`
	CreatedAt      time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	LastUsedAt     time.Time `bun:",nullzero" json:"last_used_at"`
}

type AuthChallenge struct {
	bun.BaseModel `bun:"table:auth_challenges"`

	ID        string    `bun:",pk" json:"id"`
	UserID    string    `bun:",notnull" json:"user_id"`
	Type      string    `bun:",notnull" json:"type"`
	Payload   string    `bun:",notnull" json:"payload"`
	ExpiresAt time.Time `bun:",notnull" json:"expires_at"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type APIToken struct {
	bun.BaseModel `bun:"table:api_tokens"`

	ID          string    `bun:",pk" json:"id"`
	UserID      string    `bun:",notnull" json:"user_id"`
	Name        string    `bun:",notnull" json:"name"`
	TokenHash   string    `bun:",unique,notnull" json:"-"`
	TokenPrefix string    `bun:",notnull" json:"token_prefix"`
	Scope       string    `bun:",notnull,default:'cli:full'" json:"scope"`
	Audience    string    `json:"audience"`
	ExpiresAt   time.Time `bun:",nullzero" json:"expires_at"`
	LastUsedAt  time.Time `bun:",nullzero" json:"last_used_at"`
	RevokedAt   time.Time `bun:",nullzero" json:"revoked_at"`
	CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type MCPOAuthCode struct {
	bun.BaseModel `bun:"table:mcp_oauth_codes"`

	ID                  string    `bun:",pk" json:"id"`
	CodeHash            string    `bun:",unique,notnull" json:"-"`
	UserID              string    `bun:",notnull" json:"user_id"`
	ClientID            string    `bun:",notnull" json:"client_id"`
	ClientName          string    `json:"client_name"`
	RedirectURI         string    `bun:",notnull" json:"redirect_uri"`
	Scope               string    `bun:",notnull,default:'mcp:full'" json:"scope"`
	Resource            string    `json:"resource"`
	CodeChallenge       string    `bun:",notnull" json:"code_challenge"`
	CodeChallengeMethod string    `bun:",notnull" json:"code_challenge_method"`
	ExpiresAt           time.Time `bun:",notnull" json:"expires_at"`
	ConsumedAt          time.Time `bun:",nullzero" json:"consumed_at"`
	CreatedAt           time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type CLIAuthSession struct {
	bun.BaseModel `bun:"table:cli_auth_sessions"`

	ID              string    `bun:",pk" json:"id"`
	UserID          string    `json:"user_id"`
	DeviceCodeHash  string    `bun:",unique,notnull" json:"-"`
	UserCodeHash    string    `bun:",unique,notnull" json:"-"`
	ClientName      string    `bun:",notnull" json:"client_name"`
	ClientVersion   string    `json:"client_version"`
	ClientOS        string    `json:"client_os"`
	RequestedScopes string    `bun:",notnull,default:'cli:full'" json:"requested_scopes"`
	Status          string    `bun:",notnull,default:'pending'" json:"status"`
	IntervalSeconds int       `bun:",notnull,default:5" json:"interval_seconds"`
	ExpiresAt       time.Time `bun:",notnull" json:"expires_at"`
	LastPolledAt    time.Time `bun:",nullzero" json:"last_polled_at"`
	ApprovedAt      time.Time `bun:",nullzero" json:"approved_at"`
	DeniedAt        time.Time `bun:",nullzero" json:"denied_at"`
	CreatedAt       time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type WorkspaceMember struct {
	bun.BaseModel `bun:"table:workspace_members"`

	WorkspaceID string `bun:",pk" json:"workspace_id"`
	UserID      string `bun:",pk" json:"user_id"`
	Role        string `bun:",notnull" json:"role"` // 'admin', 'editor', 'viewer'
}

type UsageCounter struct {
	bun.BaseModel `bun:"table:usage_counters"`

	WorkspaceID string    `bun:",pk" json:"workspace_id"`
	Metric      string    `bun:",pk" json:"metric"`
	PeriodStart time.Time `bun:",pk" json:"period_start"`
	Value       int64     `bun:",notnull,default:0" json:"value"`
	CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

type BillingSubscription struct {
	bun.BaseModel `bun:"table:billing_subscriptions"`

	WorkspaceID            string    `bun:",pk" json:"workspace_id"`
	Provider               string    `bun:",notnull,default:'polar'" json:"provider"`
	ProviderCustomerID     string    `bun:",notnull" json:"provider_customer_id"`
	ProviderSubscriptionID string    `bun:",notnull,unique" json:"provider_subscription_id"`
	ProviderProductID      string    `json:"provider_product_id"`
	ProviderPriceID        string    `json:"provider_price_id"`
	Status                 string    `bun:",notnull" json:"status"`
	PlanID                 string    `bun:",notnull,default:''" json:"plan_id"`
	EntitlementSnapshot    string    `bun:",notnull,default:'{}'" json:"entitlement_snapshot"`
	CurrentPeriodEnd       time.Time `bun:",nullzero" json:"current_period_end"`
	CancelAtPeriodEnd      bool      `bun:",notnull,default:false" json:"cancel_at_period_end"`
	RawPayload             string    `bun:",notnull,default:'{}'" json:"raw_payload"`
	CreatedAt              time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt              time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

type BillingWebhookEvent struct {
	bun.BaseModel `bun:"table:billing_webhook_events"`

	EventID     string    `bun:",pk" json:"event_id"`
	Provider    string    `bun:",notnull,default:'polar'" json:"provider"`
	EventType   string    `bun:",notnull" json:"event_type"`
	ProcessedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"processed_at"`
}

type MCPToolCall struct {
	bun.BaseModel `bun:"table:mcp_tool_calls"`

	ID                string    `bun:",pk" json:"id"`
	UserID            string    `bun:",notnull" json:"user_id"`
	WorkspaceID       string    `bun:",nullzero" json:"workspace_id"`
	ClientID          string    `bun:",nullzero" json:"client_id"`
	ClientName        string    `json:"client_name"`
	ClientScope       string    `json:"client_scope"`
	ClientTokenPrefix string    `json:"client_token_prefix"`
	ToolName          string    `bun:",notnull" json:"tool_name"`
	Status            string    `bun:",notnull" json:"status"`
	ErrorMessage      string    `json:"error_message"`
	DurationMs        int64     `bun:",notnull,default:0" json:"duration_ms"`
	CreatedAt         time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type MastodonInstance struct {
	bun.BaseModel `bun:"table:mastodon_instances"`

	ID                 string    `bun:",pk" json:"id"`
	InstanceURL        string    `bun:",unique,notnull" json:"instance_url"`
	Host               string    `bun:",notnull" json:"host"`
	ClientID           string    `bun:",notnull" json:"client_id"`
	ClientSecretEnc    []byte    `bun:"client_secret_encrypted,notnull" json:"-"`
	RedirectURI        string    `bun:",notnull" json:"redirect_uri"`
	Scopes             string    `bun:",notnull,default:'read write'" json:"scopes"`
	RegistrationStatus string    `bun:",notnull,default:'registered'" json:"registration_status"`
	LastVerifiedAt     time.Time `bun:",nullzero" json:"last_verified_at"`
	BlockedAt          time.Time `bun:",nullzero" json:"blocked_at"`
	BlockReason        string    `json:"block_reason"`
	CreatedAt          time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt          time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

type SocialAccount struct {
	bun.BaseModel `bun:"table:social_accounts"`

	ID               string `bun:",pk" json:"id"`
	WorkspaceID      string `bun:",notnull" json:"workspace_id"`
	Slug             string `bun:",notnull" json:"slug"`
	Platform         string `bun:",notnull" json:"platform"` // 'x', 'threads', 'linkedin', 'mastodon', 'bluesky', 'instagram', 'facebook', 'tiktok'
	AccountID        string `bun:",notnull" json:"account_id"`
	AccountUsername  string `json:"account_username"`
	AccountAvatarURL string `json:"account_avatar_url"`
	InstanceURL      string `json:"instance_url"` // Used for Mastodon domains and Bluesky PDS

	AccessTokenEnc  []byte    `bun:"access_token_encrypted,notnull" json:"-"`
	RefreshTokenEnc []byte    `bun:"refresh_token_encrypted" json:"-"`
	TokenExpiresAt  time.Time `json:"token_expires_at"`

	IsActive     bool      `bun:",default:true" json:"is_active"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type XOAuthRequestToken struct {
	bun.BaseModel `bun:"table:x_oauth_request_tokens"`

	RequestToken  string    `bun:",pk" json:"request_token"`
	RequestSecret string    `bun:",notnull" json:"-"`
	WorkspaceID   string    `bun:",notnull" json:"workspace_id"`
	UserID        string    `bun:",notnull" json:"user_id"`
	CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type OAuthAccountSelection struct {
	bun.BaseModel `bun:"table:oauth_account_selections"`

	ID              string    `bun:",pk" json:"id"`
	UserID          string    `bun:",notnull" json:"user_id"`
	WorkspaceID     string    `bun:",notnull" json:"workspace_id"`
	Platform        string    `bun:",notnull" json:"platform"`
	InstanceURL     string    `json:"instance_url"`
	AccessTokenEnc  []byte    `bun:"access_token_encrypted,notnull" json:"-"`
	RefreshTokenEnc []byte    `bun:"refresh_token_encrypted" json:"-"`
	TokenType       string    `json:"token_type"`
	TokenExpiresAt  time.Time `json:"token_expires_at"`
	TokenExtraJSON  string    `bun:"token_extra_json,notnull,default:'{}'" json:"-"`
	OptionsJSON     string    `bun:"options_json,notnull,default:'[]'" json:"-"`
	ExpiresAt       time.Time `bun:",notnull" json:"expires_at"`
	ConsumedAt      time.Time `bun:",nullzero" json:"consumed_at"`
	CreatedAt       time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// Publication is the user's canonical unit of intent: one idea, launch,
// update, or announcement that can produce platform-specific posts.
type Publication struct {
	bun.BaseModel `bun:"table:publications"`

	ID              string    `bun:",pk" json:"id"`
	WorkspaceID     string    `bun:",notnull" json:"workspace_id"`
	CreatedByID     string    `bun:"created_by,notnull" json:"created_by"`
	Title           string    `bun:",notnull" json:"title"`
	SourceContent   string    `bun:"source_content,notnull" json:"source_content"`
	SourceURL       string    `bun:"source_url" json:"source_url"`
	Goal            string    `json:"goal"`
	Audience        string    `json:"audience"`
	Status          string    `bun:",notnull,default:'draft'" json:"status"`
	ReleasePlanJSON string    `bun:"release_plan_json,notnull,default:'{}'" json:"release_plan_json"`
	CreatedAt       time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

type Post struct {
	bun.BaseModel `bun:"table:posts"`

	ID            string `bun:",pk" json:"id"`
	WorkspaceID   string `bun:",notnull" json:"workspace_id"`
	CreatedByID   string `bun:"created_by,notnull" json:"created_by"`
	PublicationID string `bun:"publication_id" json:"publication_id"`
	Content       string `bun:",notnull" json:"content"`

	ParentPostID   string `json:"parent_post_id"`
	ThreadSequence int    `bun:",default:0" json:"thread_sequence"`

	Status             string    `bun:",notnull" json:"status"` // 'draft', 'scheduled', 'publishing', 'published', 'failed'
	ScheduledAt        time.Time `json:"scheduled_at"`
	PublishedAt        time.Time `json:"published_at"`
	RandomDelayMinutes int       `bun:",default:0" json:"random_delay_minutes"`
	ActualRunAt        time.Time `bun:",nullzero" json:"actual_run_at"` // Set by worker, differs from ScheduledAt if randomized
	CreatedAt          time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type PostDestination struct {
	bun.BaseModel `bun:"table:post_destinations"`

	ID              string `bun:",pk" json:"id"`
	PostID          string `bun:",notnull" json:"post_id"`
	SocialAccountID string `bun:",notnull" json:"social_account_id"`
	ExternalID      string `json:"external_id"`
	Status          string `bun:",notnull" json:"status"` // 'pending', 'success', 'failed'
	ErrorMessage    string `json:"error_message"`
}

type MediaAttachment struct {
	bun.BaseModel `bun:"table:media_attachments"`

	ID               string    `bun:",pk" json:"id"`
	WorkspaceID      string    `bun:",notnull" json:"workspace_id"`
	FilePath         string    `bun:",notnull" json:"file_path"`
	StorageType      string    `bun:",default:'local'" json:"storage_type"` // 'local', 's3'
	MimeType         string    `json:"mime_type"`
	ProcessingStatus string    `bun:",default:'ready'" json:"processing_status"` // 'processing', 'ready', 'failed'
	Size             int64     `json:"size"`
	OriginalFilename string    `json:"original_filename"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	ThumbnailsJSON   string    `bun:"thumbnails" json:"thumbnails"` // JSON: {"sm": "sm_xxx.jpg", "md": "md_xxx.jpg"}
	FileHash         string    `bun:",unique" json:"-"`             // SHA-256 for deduplication
	AltText          string    `json:"alt_text"`
	IsFavorite       bool      `bun:",default:false" json:"is_favorite"`
	CreatedAt        time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type PostMedia struct {
	bun.BaseModel `bun:"table:post_media"`

	PostID       string `bun:",pk" json:"post_id"`
	MediaID      string `bun:",pk" json:"media_id"`
	DisplayOrder int    `json:"display_order"`
}

type PublicationAsset struct {
	bun.BaseModel `bun:"table:publication_assets"`

	PublicationID string    `bun:",pk" json:"publication_id"`
	MediaID       string    `bun:",pk" json:"media_id"`
	DisplayOrder  int       `bun:",notnull,default:0" json:"display_order"`
	CreatedAt     time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type Job struct {
	bun.BaseModel `bun:"table:jobs"`

	ID          string    `bun:",pk" json:"id"`
	Type        string    `bun:",notnull" json:"type"` // 'publish_post', 'refresh_token'
	Payload     string    `bun:",notnull" json:"payload"`
	Status      string    `bun:",default:'pending'" json:"status"` // 'pending', 'processing', 'completed', 'failed'
	RunAt       time.Time `bun:",notnull" json:"run_at"`
	Attempts    int       `bun:",default:0" json:"attempts"`
	MaxAttempts int       `bun:",default:3" json:"max_attempts"`
	LastError   string    `json:"last_error"`
	LockedAt    time.Time `json:"locked_at"`
	LockedBy    string    `json:"locked_by"`
}

type SocialMediaSet struct {
	bun.BaseModel `bun:"table:social_media_sets"`

	ID          string    `bun:",pk" json:"id"`
	WorkspaceID string    `bun:",notnull" json:"workspace_id"`
	Name        string    `bun:",notnull" json:"name"`
	IsDefault   bool      `bun:",default:false" json:"is_default"`
	CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

type SocialMediaSetAccount struct {
	bun.BaseModel `bun:"table:social_media_set_accounts"`

	SetID           string `bun:",pk" json:"set_id"`
	SocialAccountID string `bun:",pk" json:"social_account_id"`
	IsMain          bool   `bun:",default:false" json:"is_main"`
}

type PostVariant struct {
	bun.BaseModel `bun:"table:post_variants"`

	ID              string    `bun:",pk" json:"id"`
	PostID          string    `bun:",notnull" json:"post_id"`
	SocialAccountID string    `bun:",notnull" json:"social_account_id"`
	Content         string    `bun:",notnull" json:"content"`
	MediaIDs        string    `bun:"media_ids,notnull" json:"media_ids"` // JSON array of media IDs override
	IsUnsynced      bool      `bun:",default:false" json:"is_unsynced"`
	CreatedAt       time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// PostingSchedule defines preferred time slots for posting per workspace.
type PostingSchedule struct {
	bun.BaseModel `bun:"table:posting_schedules"`

	ID          string `bun:",pk" json:"id"`
	WorkspaceID string `bun:",notnull" json:"workspace_id"`
	SetID       string `json:"set_id"` // Optional: per-set schedules

	// Store times in UTC for consistency, convert on read using workspace timezone
	UTCHour   int `bun:",notnull" json:"utc_hour"`    // 0-23 UTC
	UTCMinute int `bun:",notnull" json:"utc_minute"`  // 0-59 UTC
	DayOfWeek int `bun:",notnull" json:"day_of_week"` // 0=Sunday, 6=Saturday (in UTC)

	// Display/helpers
	Label    string `json:"label"` // e.g., "Morning", "Lunch", "Evening"
	IsActive bool   `bun:",default:true" json:"is_active"`

	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}

// ThreadDraft is the per-post, not-yet-published thread state.
//
// While a user is composing a multi-post thread, the parent Post row holds
// only the parent post's text in `content`, and the full unsaved thread state
// (parent + every child post + per-account content variants) is encoded as
// JSON in this table. The JSON shape is shared with the frontend
// (`frontend/src/lib/components/compose/draft-utils.ts`):
//
//	{ "p": [ { "k": "key", "c": "content", "m": ["media_id", ...] } ],
//	  "v": { "<social_account_id>": { "<post_key>": { "content": "...",
//	                                                 "mediaIds": [...] } } } }
//
// On publish, the thread becomes real `posts` rows linked by `ParentPostID`,
// and this row is no longer authoritative. It is left in place (cheap) and
// will be re-upserted on the next edit of the parent post.
//
// Cascade-delete with the parent post: deleting a draft thread removes this
// row, and publishing a thread leaves it behind as a benign cached draft.
type ThreadDraft struct {
	bun.BaseModel `bun:"table:thread_drafts"`

	PostID    string    `bun:",pk" json:"post_id"`
	DraftJSON string    `bun:"draft_json,notnull" json:"-"`
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// Prompt represents a writing prompt for content inspiration.
type Prompt struct {
	bun.BaseModel `bun:"table:prompts"`

	ID          string    `bun:",pk" json:"id"`
	WorkspaceID string    `json:"workspace_id"` // null = global prompt
	UserID      string    `json:"user_id"`      // null = workspace/global prompt
	Text        string    `bun:",notnull" json:"text"`
	Category    string    `bun:",notnull" json:"category"`
	IsBuiltIn   bool      `bun:",default:false" json:"is_built_in"`
	CreatedAt   time.Time `bun:",nullzero,notnull,default:current_timestamp" json:"created_at"`
}
