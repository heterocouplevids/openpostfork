package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openpost/cli/internal/api"
	"github.com/openpost/cli/internal/auth"
	"github.com/openpost/cli/internal/config"
	"github.com/openpost/cli/internal/output"
)

func runtimeFrom(cmd *cobra.Command) (*config.Runtime, error) {
	cfg := config.FromCommand(cmd)
	if cfg == nil {
		return nil, fmt.Errorf("runtime config was not loaded")
	}
	return cfg, nil
}

func printerFrom(cfg *config.Runtime) *output.Printer {
	return output.New(cfg.AsJSON, cfg.Quiet)
}

func clientFrom(cfg *config.Runtime) (*api.Client, error) {
	if cfg.Instance == "" {
		return nil, fmt.Errorf("instance is required: run `openpost instance add <name> <url>` or pass --instance")
	}
	token := cfg.Token
	if token == "" {
		stored, err := auth.NewStore(cfg).Get(cfg.ProfileName)
		if err != nil {
			return nil, api.ErrAuthRequired
		}
		token = stored
	}
	c := api.New(cfg.Instance, token)
	return c, c.CheckToken()
}

func updateProfile(profileName string, mutate func(*config.Profile)) error {
	file, err := config.LoadConfig()
	if err != nil {
		return err
	}
	if file.Profiles == nil {
		file.Profiles = map[string]config.Profile{}
	}
	prof := file.Profiles[profileName]
	mutate(&prof)
	file.Profiles[profileName] = prof
	if file.CurrentProfile == "" {
		file.CurrentProfile = profileName
	}
	return config.Save(file)
}

func yesNo(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}

func normalizeInstanceURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}
