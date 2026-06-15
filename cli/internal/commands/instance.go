package commands

import (
	"fmt"

	"github.com/spf13/cobra"

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
