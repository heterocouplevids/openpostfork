package providerapps

import (
	"context"
	"fmt"

	"github.com/openpost/backend/internal/models"
	"github.com/openpost/backend/internal/platform"
	"github.com/openpost/backend/internal/services/crypto"
	"github.com/uptrace/bun"
)

type Service struct {
	db        *bun.DB
	encryptor *crypto.TokenEncryptor
}

func NewService(db *bun.DB, encryptor *crypto.TokenEncryptor) *Service {
	return &Service{db: db, encryptor: encryptor}
}

func (s *Service) ListActiveAppConfigs(ctx context.Context) ([]platform.AppConfig, error) {
	var apps []models.ProviderApp
	if err := s.db.NewSelect().
		Model(&apps).
		Where("is_active = ?", true).
		Order("provider ASC", "name ASC", "instance_url ASC").
		Scan(ctx); err != nil {
		return nil, fmt.Errorf("failed to list provider apps: %w", err)
	}

	configs := make([]platform.AppConfig, 0, len(apps))
	for _, app := range apps {
		config, err := s.toAppConfig(app)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func (s *Service) toAppConfig(app models.ProviderApp) (platform.AppConfig, error) {
	secret, err := s.encryptor.Decrypt(app.ClientSecretEnc)
	if err != nil {
		return platform.AppConfig{}, fmt.Errorf("failed to decrypt provider app %s (%s): %w", app.ID, app.Provider, err)
	}

	return platform.NormalizeAppConfig(platform.AppConfig{
		Provider:     app.Provider,
		Name:         app.Name,
		ClientID:     app.ClientID,
		ClientSecret: secret,
		RedirectURI:  app.RedirectURI,
		InstanceURL:  app.InstanceURL,
	}), nil
}
