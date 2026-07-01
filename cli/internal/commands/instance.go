package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/openpost/cli/internal/api"
	"github.com/openpost/cli/internal/auth"
	"github.com/openpost/cli/internal/config"
)

func newInstanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage OpenPost instance profiles",
	}
	cmd.AddCommand(newInstanceAddCmd())
	cmd.AddCommand(newInstanceListCmd())
	cmd.AddCommand(newInstanceUseCmd())
	cmd.AddCommand(newInstanceRemoveCmd())
	cmd.AddCommand(newInstanceHealthCmd())
	cmd.AddCommand(newInstanceDiagnosticsCmd())
	return cmd
}

func newInstanceAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add or update an instance profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			file, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if file.Profiles == nil {
				file.Profiles = map[string]config.Profile{}
			}
			name := args[0]
			prof := file.Profiles[name]
			prof.Instance = normalizeInstanceURL(args[1])
			file.Profiles[name] = prof
			if file.CurrentProfile == "" {
				file.CurrentProfile = name
			}
			if err := config.Save(file); err != nil {
				return err
			}
			printerFrom(cfg).Printf("Saved instance profile %q.", name)
			return nil
		},
	}
}

func newInstanceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured instances",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			file, err := config.LoadConfig()
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(file)
			}
			rows := make([][]string, 0, len(file.Profiles))
			for name, prof := range file.Profiles {
				current := ""
				if name == file.CurrentProfile {
					current = "*"
				}
				rows = append(rows, []string{current, name, emptyDash(prof.Instance)})
			}
			p.Table([]string{"CURRENT", "NAME", "URL"}, rows)
			return nil
		},
	}
}

func newInstanceUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Set the active instance profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			file, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if _, ok := file.Profiles[args[0]]; !ok {
				return fmt.Errorf("profile %q not found", args[0])
			}
			file.CurrentProfile = args[0]
			if err := config.Save(file); err != nil {
				return err
			}
			printerFrom(cfg).Printf("Using profile %q.", args[0])
			return nil
		},
	}
}

func newInstanceRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an instance profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			file, err := config.LoadConfig()
			if err != nil {
				return err
			}
			if _, ok := file.Profiles[args[0]]; !ok {
				return fmt.Errorf("profile %q not found", args[0])
			}
			delete(file.Profiles, args[0])
			if file.CurrentProfile == args[0] {
				file.CurrentProfile = ""
			}
			if err := config.Save(file); err != nil {
				return err
			}
			printerFrom(cfg).Printf("Removed profile %q.", args[0])
			return nil
		},
	}
}

func newInstanceHealthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check the active instance liveness and readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			if cfg.Instance == "" {
				return fmt.Errorf("instance is required: run `openpost instance add <name> <url>` or pass --instance")
			}
			client := api.New(cfg.Instance, "")
			if err := client.Health(cmd.Context()); err != nil {
				return err
			}
			ready, err := client.Ready(cmd.Context())
			if err != nil {
				return err
			}

			p := printerFrom(cfg)
			out := map[string]string{
				"instance": cfg.Instance,
				"health":   "ok",
				"ready":    ready.Status,
				"database": ready.Database,
			}
			if cfg.AsJSON {
				return p.PrintJSON(out)
			}
			p.Table([]string{"INSTANCE", "HEALTH", "READY", "DATABASE"}, [][]string{{
				out["instance"],
				out["health"],
				out["ready"],
				out["database"],
			}})
			return nil
		},
	}
}

type instanceDiagnostics struct {
	CLIVersion     string   `json:"cli_version"`
	Platform       string   `json:"platform"`
	Profile        string   `json:"profile"`
	Instance       string   `json:"instance"`
	ConfigPath     string   `json:"config_path"`
	CredentialPath string   `json:"credential_path"`
	Token          bool     `json:"token"`
	TokenSource    string   `json:"token_source,omitempty"`
	Health         string   `json:"health"`
	Ready          string   `json:"ready"`
	Database       string   `json:"database"`
	Authenticated  bool     `json:"authenticated"`
	UserEmail      string   `json:"user_email,omitempty"`
	Workspace      string   `json:"workspace,omitempty"`
	WorkspaceCount int      `json:"workspace_count"`
	Errors         []string `json:"errors,omitempty"`
}

func newInstanceDiagnosticsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diagnostics",
		Short: "Collect a safe support snapshot for an OpenPost instance",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			if cfg.Instance == "" {
				return fmt.Errorf("instance is required: run `openpost instance add <name> <url>` or pass --instance")
			}

			out := instanceDiagnostics{
				CLIVersion:     cmd.Root().Version,
				Platform:       config.Platform(),
				Profile:        cfg.ProfileName,
				Instance:       cfg.Instance,
				ConfigPath:     cfg.ConfigPath,
				CredentialPath: cfg.CredentialPath,
				Workspace:      cfg.Workspace,
				Health:         "unknown",
				Ready:          "unknown",
				Database:       "unknown",
			}

			publicClient := api.New(cfg.Instance, "")
			if err := publicClient.Health(cmd.Context()); err != nil {
				out.Health = "error"
				out.Errors = append(out.Errors, "health: "+err.Error())
			} else {
				out.Health = "ok"
			}
			ready, err := publicClient.Ready(cmd.Context())
			if err != nil {
				out.Ready = "error"
				out.Database = "unknown"
				out.Errors = append(out.Errors, "readiness: "+err.Error())
			} else {
				out.Ready = ready.Status
				out.Database = ready.Database
			}

			token, source, hasToken := diagnosticsToken(cfg)
			out.Token = hasToken
			out.TokenSource = source
			if token != "" {
				authClient := api.New(cfg.Instance, token)
				if me, err := authClient.Me(cmd.Context()); err != nil {
					out.Errors = append(out.Errors, "auth: "+err.Error())
				} else {
					out.Authenticated = true
					out.UserEmail = me.Email
				}
				if workspaces, err := authClient.ListWorkspaces(cmd.Context()); err != nil {
					out.Errors = append(out.Errors, "workspaces: "+err.Error())
				} else {
					out.WorkspaceCount = len(workspaces)
				}
			}

			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(out)
			}
			p.Table([]string{"KEY", "VALUE"}, [][]string{
				{"CLI version", emptyDash(out.CLIVersion)},
				{"Platform", out.Platform},
				{"Profile", out.Profile},
				{"Instance", out.Instance},
				{"Config path", out.ConfigPath},
				{"Credential path", out.CredentialPath},
				{"Token", yesNo(out.Token)},
				{"Token source", emptyDash(out.TokenSource)},
				{"Health", out.Health},
				{"Ready", out.Ready},
				{"Database", out.Database},
				{"Authenticated", yesNo(out.Authenticated)},
				{"User email", emptyDash(out.UserEmail)},
				{"Workspace", emptyDash(out.Workspace)},
				{"Workspace count", fmt.Sprintf("%d", out.WorkspaceCount)},
			})
			if len(out.Errors) > 0 {
				p.Printf("")
				for _, diagnosticErr := range out.Errors {
					p.Printf("error: %s", diagnosticErr)
				}
			}
			return nil
		},
	}
}

func diagnosticsToken(cfg *config.Runtime) (token string, source string, ok bool) {
	if cfg.Token != "" {
		return cfg.Token, "flag/env", true
	}
	source, ok = auth.HasToken(cfg)
	if !ok {
		return "", "", false
	}
	token, err := auth.NewStore(cfg).Get(cfg.ProfileName)
	if err != nil {
		return "", source, true
	}
	return token, source, true
}
