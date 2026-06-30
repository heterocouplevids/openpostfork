package account_saver

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/platform"
	"github.com/openpost/backend/internal/services/crypto"
	"github.com/openpost/backend/internal/services/entitlements"
	"github.com/openpost/backend/internal/services/tokenmanager"
	"github.com/uptrace/bun"
)

var slugUnsafeChars = regexp.MustCompile(`[^a-z0-9]+`)

// AccountSaver handles saving social account information to the database.
// This service extracts the duplicated account-saving logic from the OAuth handler.
type AccountSaver struct {
	db          *bun.DB
	crypto      *crypto.TokenEncryptor
	entitlement entitlements.Service
}

// NewAccountSaver creates a new AccountSaver instance.
func NewAccountSaver(db *bun.DB, crypto *crypto.TokenEncryptor, entitlement ...entitlements.Service) *AccountSaver {
	entitlementService := entitlements.Service(entitlements.NewSelfHostedService())
	if len(entitlement) > 0 && entitlement[0] != nil {
		entitlementService = entitlement[0]
	}
	return &AccountSaver{
		db:          db,
		crypto:      crypto,
		entitlement: entitlementService,
	}
}

func (s *AccountSaver) SetEntitlement(entitlement entitlements.Service) {
	if entitlement != nil {
		s.entitlement = entitlement
	}
}

// SaveAccount saves a social account with encrypted tokens.
// It handles the common logic of extracting account info, encrypting tokens,
// and inserting into the social_accounts table.
//
//nolint:gocyclo
func (s *AccountSaver) SaveAccount(ctx context.Context, userID, platformName, workspaceID, accountID, accountUsername, instanceURL string, tokenResp *platform.TokenResult) (*models.SocialAccount, error) {
	if s == nil || s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace id is required")
	}

	memberCount, err := s.db.NewSelect().
		Model((*models.WorkspaceMember)(nil)).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("validating workspace membership: %w", err)
	}
	if memberCount == 0 {
		return nil, fmt.Errorf("workspace not accessible")
	}

	// For Threads, the account ID comes from the token response extra
	if tokenResp.Extra != nil {
		if uid, ok := tokenResp.Extra["user_id"]; ok && uid != "" {
			accountID = uid
		}
	}
	if err := s.checkSocialAccountQuota(ctx, workspaceID); err != nil {
		return nil, err
	}

	encAccess, err := s.crypto.Encrypt(tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	var encRefresh []byte
	if tokenResp.RefreshToken != "" {
		encRefresh, err = s.crypto.Encrypt(tokenResp.RefreshToken)
		if err != nil {
			return nil, err
		}
	}

	var expiresAt time.Time
	if tokenResp.ExpiresIn > 0 {
		expiresAt = time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	account := &models.SocialAccount{
		ID:              uuid.New().String(),
		WorkspaceID:     workspaceID,
		Slug:            "",
		Platform:        platformName,
		AccountID:       accountID,
		AccountUsername: accountUsername,
		InstanceURL:     instanceURL,
		AccessTokenEnc:  encAccess,
		RefreshTokenEnc: encRefresh,
		TokenExpiresAt:  expiresAt,
		IsActive:        true,
		CreatedAt:       time.Now().UTC(),
	}
	account.Slug = s.uniqueSlug(ctx, workspaceID, defaultSlug(platformName, accountUsername, accountID, instanceURL))

	if _, err := s.db.NewInsert().Model(account).Exec(ctx); err != nil {
		return nil, err
	}

	if err := tokenmanager.ScheduleRefreshJob(ctx, s.db, account.ID, expiresAt); err != nil {
		log.Printf("[AccountSaver] Failed to schedule refresh job for account %s: %v", account.ID, err)
	}

	return account, nil
}

func (s *AccountSaver) checkSocialAccountQuota(ctx context.Context, workspaceID string) error {
	current, err := s.db.NewSelect().
		Model((*models.SocialAccount)(nil)).
		Where("workspace_id = ?", workspaceID).
		Where("is_active = ?", true).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("loading social account usage: %w", err)
	}

	decision, err := s.entitlement.Check(ctx, entitlements.Request{
		WorkspaceID: workspaceID,
		Limit:       entitlements.LimitSocialAccounts,
		Current:     int64(current),
		Amount:      1,
	})
	if err != nil {
		return fmt.Errorf("checking social account limit: %w", err)
	}
	if !decision.Allowed {
		if decision.Reason != "" {
			return fmt.Errorf("%s", decision.Reason)
		}
		return fmt.Errorf("social account limit exceeded")
	}
	return nil
}

func defaultSlug(platformName, accountUsername, accountID, instanceURL string) string {
	label := strings.TrimSpace(accountUsername)
	if label == "" {
		label = strings.TrimSpace(accountID)
	}
	if label == "" {
		label = strings.TrimSpace(instanceURL)
	}
	base := strings.Trim(strings.ToLower(platformName+"-"+label), "-")
	base = slugUnsafeChars.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if len(base) > 63 {
		base = strings.Trim(base[:63], "-")
	}
	if base == "" {
		return strings.ToLower(platformName)
	}
	return base
}

func (s *AccountSaver) uniqueSlug(ctx context.Context, workspaceID, base string) string {
	if base == "" {
		base = "account"
	}
	for i := 0; ; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}

		var existing models.SocialAccount
		err := s.db.NewSelect().
			Model(&existing).
			Where("workspace_id = ?", workspaceID).
			Where("slug = ?", candidate).
			Where("is_active = ?", true).
			Scan(ctx)
		if err != nil {
			if err == sql.ErrNoRows {
				return candidate
			}
			return base
		}
	}
}
