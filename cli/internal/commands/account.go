package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openpost/cli/internal/api"
)

func newAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage connected social accounts",
	}
	cmd.AddCommand(newAccountListCmd())
	cmd.AddCommand(newAccountDisconnectCmd())
	cmd.AddCommand(newAccountConnectCmd())
	return cmd
}

func newAccountListCmd() *cobra.Command {
	var platform string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connected social accounts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			client, err := clientFrom(cfg)
			if err != nil {
				return err
			}
			workspaceID, err := activeWorkspaceID(cmd, client)
			if err != nil {
				return err
			}
			accounts, err := client.ListAccounts(cmd.Context(), workspaceID)
			if err != nil {
				return err
			}
			if platform != "" {
				accounts = filterAccountsByPlatform(accounts, platform)
			}

			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(accounts)
			}
			if len(accounts) == 0 {
				if platform != "" {
					p.Printf("No %s accounts are connected for this workspace.", platform)
				} else {
					p.Printf("No accounts are connected for this workspace.")
				}
				return nil
			}
			rows := make([][]string, 0, len(accounts))
			for _, acc := range accounts {
				rows = append(rows, []string{
					acc.ID,
					acc.Platform,
					emptyDash(acc.AccountUsername),
					emptyDash(acc.InstanceURL),
					yesNo(acc.IsActive),
				})
			}
			p.Table([]string{"ID", "PLATFORM", "USERNAME", "INSTANCE", "ACTIVE"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&platform, "platform", "", "filter by platform")
	return cmd
}

func newAccountDisconnectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disconnect <account-id>",
		Short: "Disconnect a social account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			client, err := clientFrom(cfg)
			if err != nil {
				return err
			}
			accountID := args[0]
			if !cfg.Yes && !cfg.AsJSON {
				ok, err := confirm(fmt.Sprintf("Disconnect account %s?", accountID))
				if err != nil {
					return err
				}
				if !ok {
					printerFrom(cfg).Printf("Canceled.")
					return nil
				}
			}
			if err := client.DisconnectAccount(cmd.Context(), accountID); err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(map[string]string{"id": accountID, "status": "disconnected"})
			}
			p.Printf("Disconnected account %s.", accountID)
			return nil
		},
	}
}

func newAccountConnectCmd() *cobra.Command {
	var server string

	cmd := &cobra.Command{
		Use:   "connect <platform>",
		Short: "Connect a social account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := runtimeFrom(cmd)
			if err != nil {
				return err
			}
			platform := args[0]
			if cfg.AsJSON {
				return printerFrom(cfg).PrintJSON(map[string]string{
					"platform": platform,
					"server":   server,
					"status":   "not_implemented",
					"message":  "Account connection is not yet implemented in the CLI. Use the web UI to connect " + platform + ".",
				})
			}
			printerFrom(cfg).Printf("Account connection is not yet implemented in the CLI. Use the web UI to connect %s.", platform)
			return nil
		},
	}
	cmd.Flags().StringVar(&server, "server", "", "server URL for platforms that need one")
	return cmd
}

func filterAccountsByPlatform(accounts []api.SocialAccount, platform string) []api.SocialAccount {
	out := accounts[:0]
	for _, acc := range accounts {
		if acc.Platform == platform {
			out = append(out, acc)
		}
	}
	return out
}

func confirm(prompt string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
