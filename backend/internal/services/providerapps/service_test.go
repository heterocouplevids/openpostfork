package providerapps

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/services/crypto"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

func createProviderAppsTestDB(t *testing.T) *bun.DB {
	t.Helper()

	sqldb, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString()))
	require.NoError(t, err)

	db := bun.NewDB(sqldb, sqlitedialect.New())
	_, err = db.NewCreateTable().
		Model((*models.ProviderApp)(nil)).
		IfNotExists().
		Exec(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	return db
}

func TestListActiveAppConfigsDecryptsAndNormalizesRows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := createProviderAppsTestDB(t)
	encryptor := crypto.NewTokenEncryptor("0123456789abcdef0123456789abcdef")
	service := NewService(db, encryptor)

	xSecret, err := encryptor.Encrypt("x-secret")
	require.NoError(t, err)
	mastodonSecret, err := encryptor.Encrypt("masto-secret")
	require.NoError(t, err)
	inactiveSecret, err := encryptor.Encrypt("youtube-secret")
	require.NoError(t, err)

	now := time.Now().UTC()
	rows := []models.ProviderApp{
		{
			ID:              "x-app",
			Provider:        " X ",
			ClientID:        " x-client ",
			ClientSecretEnc: xSecret,
			RedirectURI:     " https://app.test/api/v1/accounts/x/callback ",
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "mastodon-app",
			Provider:        "mastodon",
			Name:            " Personal ",
			ClientID:        " masto-client ",
			ClientSecretEnc: mastodonSecret,
			RedirectURI:     " urn:ietf:wg:oauth:2.0:oob ",
			InstanceURL:     "https://masto.pt/",
			IsActive:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "youtube-app",
			Provider:        "youtube",
			ClientID:        "youtube-client",
			ClientSecretEnc: inactiveSecret,
			IsActive:        false,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}
	for i := range rows {
		_, err = db.NewInsert().Model(&rows[i]).Exec(ctx)
		require.NoError(t, err)
	}
	_, err = db.NewUpdate().
		Model((*models.ProviderApp)(nil)).
		Set("is_active = ?", false).
		Where("id = ?", "youtube-app").
		Exec(ctx)
	require.NoError(t, err)

	configs, err := service.ListActiveAppConfigs(ctx)
	require.NoError(t, err)
	require.Len(t, configs, 2)

	byProvider := map[string]int{}
	for i, config := range configs {
		byProvider[config.Provider] = i
	}
	x := configs[byProvider["x"]]
	require.Equal(t, "x-client", x.ClientID)
	require.Equal(t, "x-secret", x.ClientSecret)
	require.Equal(t, "https://app.test/api/v1/accounts/x/callback", x.RedirectURI)

	mastodon := configs[byProvider["mastodon"]]
	require.Equal(t, "Personal", mastodon.Name)
	require.Equal(t, "masto-client", mastodon.ClientID)
	require.Equal(t, "masto-secret", mastodon.ClientSecret)
	require.Equal(t, "urn:ietf:wg:oauth:2.0:oob", mastodon.RedirectURI)
	require.Equal(t, "https://masto.pt", mastodon.InstanceURL)
}

func TestListActiveAppConfigsFailsWhenSecretCannotDecrypt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := createProviderAppsTestDB(t)
	encryptor := crypto.NewTokenEncryptor("0123456789abcdef0123456789abcdef")
	service := NewService(db, encryptor)

	now := time.Now().UTC()
	_, err := db.NewInsert().Model(&models.ProviderApp{
		ID:              "bad-app",
		Provider:        "x",
		ClientID:        "x-client",
		ClientSecretEnc: []byte("not-gcm-ciphertext"),
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Exec(ctx)
	require.NoError(t, err)

	_, err = service.ListActiveAppConfigs(ctx)
	require.ErrorContains(t, err, "failed to decrypt provider app bad-app (x)")
}
