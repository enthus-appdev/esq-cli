package searchcmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/enthus-appdev/esq-cli/internal/config"
	"github.com/enthus-appdev/esq-cli/internal/es"
	"github.com/enthus-appdev/esq-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewSearchCmd creates the search command.
func NewSearchCmd() *cobra.Command {
	var (
		size   int
		source []string
	)

	cmd := &cobra.Command{
		Use:   "search <index> <query>",
		Short: "Search with a Lucene query string",
		Long: `Search an Elasticsearch index using Lucene query syntax.

Index names support partial matching - "documents" resolves to the latest
version of erp.sales.documents. Use full names for exact matches.`,
		Example: `  esq search documents "DocumentNo:12813636"
  esq search customer "Name:Müller" --size 50
  esq search documents "DocumentStateId:2 AND CustomerName:enthus"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, envName, err := getClient(cmd)
			if err != nil {
				return err
			}

			index, err := resolveIndex(client, args[0])
			if err != nil {
				return err
			}

			query := args[1]
			output.Info("[%s] SEARCH %s for: %s (limit: %d)", envName, index, query, size)

			if len(source) > 0 {
				return searchWithSource(client, index, query, size, source)
			}

			result, err := client.Search(index, query, size)
			if err != nil {
				return err
			}
			return output.RawJSON(os.Stdout, result)
		},
	}

	cmd.Flags().IntVarP(&size, "size", "s", 10, "Maximum number of results")
	cmd.Flags().StringSliceVar(&source, "source", nil, "Limit returned fields (e.g. --source DocumentNo,DocumentStateId)")

	return cmd
}

// searchWithSource uses Query DSL to filter _source fields.
func searchWithSource(client *es.Client, index, query string, size int, source []string) error {
	body := fmt.Sprintf(`{"query":{"query_string":{"query":%q}},"size":%d,"_source":%s}`,
		query, size, formatSourceFields(source))

	result, err := client.SearchDSL(index, strings.NewReader(body))
	if err != nil {
		return err
	}
	return output.RawJSON(os.Stdout, result)
}

func formatSourceFields(fields []string) string {
	result := "["
	for i, f := range fields {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%q", f)
	}
	result += "]"
	return result
}

// getClient resolves the ES client from config, respecting --env override.
func getClient(cmd *cobra.Command) (*es.Client, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, "", err
	}

	envName, _ := cmd.Flags().GetString("env")
	if envName == "" {
		envName = cfg.CurrentEnv
	}
	if envName == "" {
		return nil, "", fmt.Errorf("no active environment. Run 'esq config add <name> --url <url>' and 'esq config use <name>'")
	}

	env, ok := cfg.Environments[envName]
	if !ok {
		return nil, "", fmt.Errorf("environment %q not found", envName)
	}

	return es.NewClient(env.URL), envName, nil
}

// resolveIndex resolves a partial index name and prints info about ambiguity.
func resolveIndex(client *es.Client, input string) (string, error) {
	resolved, wasPartial, alternatives, err := client.ResolveIndexVerbose(input)
	if err != nil {
		return "", err
	}
	if wasPartial {
		output.Warn("Multiple indices match %q, using latest: %s", input, resolved)
		output.Dim("  Also matched: %s", joinStrings(alternatives))
	}
	return resolved, nil
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
