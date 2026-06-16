package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openpost/cli/internal/accountpicker"
	"github.com/openpost/cli/internal/api"
	"github.com/openpost/cli/internal/config"
)

type setFlags struct {
	accounts  string
	isDefault bool
	unset     bool
	isMain    bool
}

func newSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{"sets"},
		Short:   "Manage workspace social sets",
		Long: "Manage workspace social sets: reusable groups of social accounts.\n\n" +
			"Posts and threads use the workspace default set when neither --accounts nor --set is passed.",
	}
	cmd.AddCommand(newSetListCmd())
	cmd.AddCommand(newSetCreateCmd())
	cmd.AddCommand(newSetRenameCmd())
	cmd.AddCommand(newSetDefaultCmd())
	cmd.AddCommand(newSetAddCmd())
	cmd.AddCommand(newSetRemoveCmd())
	cmd.AddCommand(newSetDeleteCmd())
	return cmd
}

func newSetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List social sets",
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
			sets, err := client.ListSets(cmd.Context(), workspaceID)
			if err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(sets)
			}
			rows := make([][]string, 0, len(sets))
			for _, set := range sets {
				rows = append(rows, []string{
					set.ID,
					set.Name,
					yesNo(set.IsDefault),
					strconv.Itoa(len(set.Accounts)),
					formatSetAccounts(set.Accounts),
				})
			}
			p.Table([]string{"ID", "NAME", "DEFAULT", "ACCOUNTS", "MEMBERS"}, rows)
			return nil
		},
	}
}

func newSetCreateCmd() *cobra.Command {
	var flags setFlags
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a social set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, workspaceID, err := setRuntime(cmd)
			if err != nil {
				return err
			}
			accountIDs, err := resolveAccounts(cmd, client, workspaceID, flags.accounts)
			if err != nil {
				return err
			}
			set, err := client.CreateSet(cmd.Context(), api.CreateSetInput{
				WorkspaceID: workspaceID,
				Name:        args[0],
				IsDefault:   flags.isDefault,
				AccountIDs:  accountIDs,
			})
			if err != nil {
				return err
			}
			return printSetSummary(cfg, set)
		},
	}
	cmd.Flags().StringVar(&flags.accounts, "accounts", "", "comma-separated account selectors to include")
	cmd.Flags().BoolVar(&flags.isDefault, "default", false, "make this the workspace default set")
	return cmd
}

func newSetRenameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <set> <name>",
		Short: "Rename a social set",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, workspaceID, err := setRuntime(cmd)
			if err != nil {
				return err
			}
			set, err := resolveSet(cmd, client, workspaceID, args[0])
			if err != nil {
				return err
			}
			name := args[1]
			updated, err := client.UpdateSet(cmd.Context(), set.ID, api.UpdateSetInput{Name: &name})
			if err != nil {
				return err
			}
			return printSetSummary(cfg, updated)
		},
	}
}

func newSetDefaultCmd() *cobra.Command {
	var flags setFlags
	cmd := &cobra.Command{
		Use:   "default <set>",
		Short: "Set or clear the workspace default social set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, workspaceID, err := setRuntime(cmd)
			if err != nil {
				return err
			}
			set, err := resolveSet(cmd, client, workspaceID, args[0])
			if err != nil {
				return err
			}
			isDefault := !flags.unset
			updated, err := client.UpdateSet(cmd.Context(), set.ID, api.UpdateSetInput{IsDefault: &isDefault})
			if err != nil {
				return err
			}
			return printSetSummary(cfg, updated)
		},
	}
	cmd.Flags().BoolVar(&flags.unset, "unset", false, "clear default status instead of making the set default")
	return cmd
}

func newSetAddCmd() *cobra.Command {
	var flags setFlags
	cmd := &cobra.Command{
		Use:   "add <set> --accounts <selectors>",
		Short: "Add accounts to a social set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, workspaceID, err := setRuntime(cmd)
			if err != nil {
				return err
			}
			if strings.TrimSpace(flags.accounts) == "" {
				return fmt.Errorf("--accounts is required")
			}
			set, err := resolveSet(cmd, client, workspaceID, args[0])
			if err != nil {
				return err
			}
			accountIDs, err := resolveAccounts(cmd, client, workspaceID, flags.accounts)
			if err != nil {
				return err
			}
			updated, err := client.AddSetAccounts(cmd.Context(), set.ID, api.AddSetAccountsInput{
				AccountIDs: accountIDs,
				IsMain:     boolPtr(flags.isMain, cmd.Flags().Changed("main")),
			})
			if err != nil {
				return err
			}
			return printSetSummary(cfg, updated)
		},
	}
	cmd.Flags().StringVar(&flags.accounts, "accounts", "", "comma-separated account selectors to add")
	cmd.Flags().BoolVar(&flags.isMain, "main", false, "mark added accounts as main accounts")
	_ = cmd.MarkFlagRequired("accounts")
	return cmd
}

func newSetRemoveCmd() *cobra.Command {
	var flags setFlags
	cmd := &cobra.Command{
		Use:   "remove <set> --accounts <selectors>",
		Short: "Remove accounts from a social set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, workspaceID, err := setRuntime(cmd)
			if err != nil {
				return err
			}
			if strings.TrimSpace(flags.accounts) == "" {
				return fmt.Errorf("--accounts is required")
			}
			set, err := resolveSet(cmd, client, workspaceID, args[0])
			if err != nil {
				return err
			}
			accounts, err := client.ListAccounts(cmd.Context(), workspaceID)
			if err != nil {
				return err
			}
			accountIDs, err := accountpicker.Resolve(workspaceID, splitCSV(flags.accounts), accounts)
			if err != nil {
				return err
			}
			updated := set
			for _, accountID := range accountIDs {
				updated, err = client.RemoveSetAccount(cmd.Context(), set.ID, accountID)
				if err != nil {
					return err
				}
			}
			return printSetSummary(cfg, updated)
		},
	}
	cmd.Flags().StringVar(&flags.accounts, "accounts", "", "comma-separated account selectors to remove")
	_ = cmd.MarkFlagRequired("accounts")
	return cmd
}

func newSetDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <set>",
		Short: "Delete a social set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, client, workspaceID, err := setRuntime(cmd)
			if err != nil {
				return err
			}
			set, err := resolveSet(cmd, client, workspaceID, args[0])
			if err != nil {
				return err
			}
			if !cfg.Yes && !cfg.AsJSON {
				ok, err := confirm(fmt.Sprintf("Delete set %s?", set.Name))
				if err != nil {
					return err
				}
				if !ok {
					printerFrom(cfg).Printf("Canceled.")
					return nil
				}
			}
			if err := client.DeleteSet(cmd.Context(), set.ID); err != nil {
				return err
			}
			p := printerFrom(cfg)
			if cfg.AsJSON {
				return p.PrintJSON(map[string]string{"id": set.ID, "status": "deleted"})
			}
			p.Printf("Deleted set %s.", set.Name)
			return nil
		},
	}
}

func setRuntime(cmd *cobra.Command) (*config.Runtime, *api.Client, string, error) {
	cfg, err := runtimeFrom(cmd)
	if err != nil {
		return nil, nil, "", err
	}
	client, err := clientFrom(cfg)
	if err != nil {
		return nil, nil, "", err
	}
	workspaceID, err := activeWorkspaceID(cmd, client)
	if err != nil {
		return nil, nil, "", err
	}
	return cfg, client, workspaceID, nil
}

func boolPtr(value bool, include bool) *bool {
	if !include {
		return nil
	}
	return &value
}

func printSetSummary(cfg *config.Runtime, set *api.SocialSet) error {
	p := printerFrom(cfg)
	if cfg.AsJSON {
		return p.PrintJSON(set)
	}
	p.Table([]string{"ID", "NAME", "DEFAULT", "ACCOUNTS", "MEMBERS"}, [][]string{{
		set.ID,
		set.Name,
		yesNo(set.IsDefault),
		strconv.Itoa(len(set.Accounts)),
		formatSetAccounts(set.Accounts),
	}})
	return nil
}

func formatSetAccounts(accounts []api.SetAccount) string {
	if len(accounts) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		name := acc.AccountUsername
		if name == "" {
			name = acc.SocialAccountID
		}
		parts = append(parts, fmt.Sprintf("%s:%s", acc.Platform, name))
	}
	return strings.Join(parts, ",")
}
