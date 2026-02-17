package configcmd

import (
	"fmt"
	"sort"

	"github.com/enthus-appdev/esq-cli/internal/config"
	"github.com/enthus-appdev/esq-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewConfigCmd creates the config command group.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Elasticsearch environments",
		Example: `  esq config add prod --url http://es-prod:9200
  esq config use prod
  esq config list`,
	}

	cmd.AddCommand(
		newAddCmd(),
		newUseCmd(),
		newListCmd(),
		newRemoveCmd(),
	)

	return cmd
}

func newAddCmd() *cobra.Command {
	var url string

	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add an Elasticsearch environment",
		Example: `  esq config add prod --url http://es-prod:9200
  esq config add stage --url http://es-stage:9200
  esq config add local --url http://localhost:9200`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if url == "" {
				return fmt.Errorf("--url is required\n\nExample:\n  esq config add %s --url http://hostname:9200", name)
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			_, exists := cfg.Environments[name]
			cfg.Environments[name] = &config.Environment{URL: url}

			// Auto-set as current if it's the first environment
			if cfg.CurrentEnv == "" {
				cfg.CurrentEnv = name
			}

			if err := config.Save(cfg); err != nil {
				return err
			}

			if exists {
				output.Success("Updated environment %q (url: %s)", name, url)
			} else {
				output.Success("Added environment %q (url: %s)", name, url)
			}

			if cfg.CurrentEnv == name {
				output.Info("Active environment: %s", name)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "Elasticsearch base URL (e.g. http://hostname:9200)")

	return cmd
}

func newUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Set the active environment",
		Example: `  esq config use prod
  esq config use stage`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if _, ok := cfg.Environments[name]; !ok {
				return fmt.Errorf("environment %q not found (available: %s)", name, formatNames(cfg.EnvNames()))
			}

			cfg.CurrentEnv = name
			if err := config.Save(cfg); err != nil {
				return err
			}

			output.Success("Switched to %q (%s)", name, cfg.Environments[name].URL)
			return nil
		},
	}
	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all configured environments",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if len(cfg.Environments) == 0 {
				output.Warn("No environments configured. Run 'esq config add <name> --url <url>' to add one.")
				return nil
			}

			names := cfg.EnvNames()
			sort.Strings(names)

			for _, name := range names {
				env := cfg.Environments[name]
				marker := "  "
				if name == cfg.CurrentEnv {
					marker = "* "
				}
				fmt.Printf("%s%-10s %s\n", marker, name, env.URL)
			}

			return nil
		},
	}
	return cmd
}

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove an environment",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.Load()
			if err != nil {
				return err
			}

			if _, ok := cfg.Environments[name]; !ok {
				return fmt.Errorf("environment %q not found", name)
			}

			delete(cfg.Environments, name)
			if cfg.CurrentEnv == name {
				cfg.CurrentEnv = ""
				output.Warn("Removed active environment. Use 'esq config use <name>' to set a new one.")
			}

			if err := config.Save(cfg); err != nil {
				return err
			}

			output.Success("Removed environment %q", name)
			return nil
		},
	}
	return cmd
}

func formatNames(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}
