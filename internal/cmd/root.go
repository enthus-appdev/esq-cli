package cmd

import (
	"fmt"
	"os"
	"strings"

	configcmd "github.com/enthus-appdev/esq-cli/internal/cmd/config"
	searchcmd "github.com/enthus-appdev/esq-cli/internal/cmd/search"
	"github.com/enthus-appdev/esq-cli/internal/config"
	"github.com/enthus-appdev/esq-cli/internal/es"
	"github.com/enthus-appdev/esq-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	envOverride string
	version     string
)

// Execute runs the root command and returns the exit code.
func Execute(ver, commit string) int {
	version = ver

	rootCmd := &cobra.Command{
		Use:   "esq",
		Short: "Elasticsearch Query CLI for NegSoft",
		Long:  "Query and inspect Elasticsearch clusters across environments (prod, stage, local).",
		Example: `  # Set up environments
  esq config add prod --url http://10.11.20.41:9200
  esq config add stage --url http://10.11.20.44:9200
  esq config use prod

  # Search for a document
  esq search documents "DocumentNo:12813636"

  # Check cluster health
  esq health`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (commit: %s)", ver, commit),
	}

	rootCmd.PersistentFlags().StringVarP(&envOverride, "env", "e", "", "Override active environment")

	rootCmd.AddCommand(
		configcmd.NewConfigCmd(),
		searchcmd.NewSearchCmd(),
		newGetCmd(),
		newCountCmd(),
		newQueryCmd(),
		newIndicesCmd(),
		newMappingCmd(),
		newHealthCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		output.Err("%s", err)
		return 1
	}
	return 0
}

// getClient loads config, resolves the environment, and returns an ES client.
// It also returns the environment name for display purposes.
func getClient() (*es.Client, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, "", err
	}

	envName := envOverride
	if envName == "" {
		envName = cfg.CurrentEnv
	}

	if envName == "" {
		return nil, "", fmt.Errorf("no active environment. Run 'esq config add <name> --url <url>' and 'esq config use <name>'")
	}

	env, ok := cfg.Environments[envName]
	if !ok {
		return nil, "", fmt.Errorf("environment %q not found (available: %s)", envName, formatEnvList(cfg))
	}

	return es.NewClient(env.URL), envName, nil
}

func formatEnvList(cfg *config.Config) string {
	names := cfg.EnvNames()
	if len(names) == 0 {
		return "none"
	}
	result := ""
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += name
	}
	return result
}

// resolveIndex resolves a partial index name via the ES client and prints info about it.
func resolveIndex(client *es.Client, input string) (string, error) {
	resolved, wasPartial, alternatives, err := client.ResolveIndexVerbose(input)
	if err != nil {
		return "", err
	}
	if wasPartial {
		output.Warn("Multiple indices match %q, using latest: %s", input, resolved)
		output.Dim("  Also matched: %s", formatAlternatives(alternatives))
	}
	return resolved, nil
}

func formatAlternatives(alts []string) string {
	if len(alts) > 5 {
		return fmt.Sprintf("%s (and %d more)", joinStrings(alts[:5]), len(alts)-5)
	}
	return joinStrings(alts)
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

// --- Simple commands (kept in root.go since they're small) ---

func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <index> <doc-id>",
		Short: "Get a document by its _id",
		Example: `  esq get documents abc123
  esq get erp.sales.documents_v30.1.0 offer-122116`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, envName, err := getClient()
			if err != nil {
				return err
			}
			index, err := resolveIndex(client, args[0])
			if err != nil {
				return err
			}

			output.Info("[%s] GET %s/_doc/%s", envName, index, args[1])
			result, err := client.Get(index, args[1])
			if err != nil {
				return err
			}
			return output.RawJSON(os.Stdout, result)
		},
	}
	return cmd
}

func newCountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count <index> [query]",
		Short: "Count documents in an index",
		Example: `  esq count customer
  esq count documents "DocumentStateId:2"`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, envName, err := getClient()
			if err != nil {
				return err
			}
			index, err := resolveIndex(client, args[0])
			if err != nil {
				return err
			}

			query := ""
			if len(args) > 1 {
				query = args[1]
			}

			if query != "" {
				output.Info("[%s] COUNT %s where: %s", envName, index, query)
			} else {
				output.Info("[%s] COUNT %s", envName, index)
			}

			result, err := client.Count(index, query)
			if err != nil {
				return err
			}
			return output.RawJSON(os.Stdout, result)
		},
	}
	return cmd
}

func newQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query <index> <json-body>",
		Short: "Search with Elasticsearch Query DSL",
		Long:  "Execute a full Query DSL search. Pass the JSON body as a string argument.",
		Example: `  esq query documents '{"query":{"term":{"DocumentNo":12813636}}}'
  esq query documents '{"query":{"match_all":{}},"size":1,"_source":["DocumentNo","DocumentStateId"]}'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, envName, err := getClient()
			if err != nil {
				return err
			}
			index, err := resolveIndex(client, args[0])
			if err != nil {
				return err
			}

			output.Info("[%s] QUERY %s (DSL)", envName, index)
			result, err := client.SearchDSL(index, strings.NewReader(args[1]))
			if err != nil {
				return err
			}
			return output.RawJSON(os.Stdout, result)
		},
	}
	return cmd
}

func newIndicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indices [filter]",
		Short: "List indices in the cluster",
		Example: `  esq indices
  esq indices sales
  esq indices --env stage`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, envName, err := getClient()
			if err != nil {
				return err
			}

			filter := ""
			if len(args) > 0 {
				filter = args[0]
			}

			if filter != "" {
				output.Info("[%s] Indices matching %q:", envName, filter)
			} else {
				output.Info("[%s] All indices:", envName)
			}

			result, err := client.CatIndices(filter)
			if err != nil {
				return err
			}
			fmt.Print(result)
			return nil
		},
	}
	return cmd
}

func newMappingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mapping <index>",
		Short: "Show index field mapping",
		Example: `  esq mapping documents
  esq mapping erp.sales.documents_v30.1.0`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, envName, err := getClient()
			if err != nil {
				return err
			}
			index, err := resolveIndex(client, args[0])
			if err != nil {
				return err
			}

			output.Info("[%s] Mapping for %s:", envName, index)
			result, err := client.Mapping(index)
			if err != nil {
				return err
			}
			return output.RawJSON(os.Stdout, result)
		},
	}
	return cmd
}

func newHealthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "health",
		Short:   "Show cluster health and node stats",
		Example: "  esq health\n  esq health --env stage",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, envName, err := getClient()
			if err != nil {
				return err
			}

			output.Info("[%s] Cluster health:", envName)
			health, err := client.Health()
			if err != nil {
				return err
			}
			if err := output.RawJSON(os.Stdout, health); err != nil {
				return err
			}

			fmt.Println()
			output.Info("[%s] Nodes:", envName)
			nodes, err := client.Nodes()
			if err != nil {
				return err
			}
			fmt.Print(nodes)
			return nil
		},
	}
	return cmd
}
