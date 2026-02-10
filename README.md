# esq - Elasticsearch Query CLI

Query and inspect Elasticsearch clusters across environments from the command line.

## Installation

Requires Go 1.24+.

```bash
# Via go install (requires SSH access to the repo)
GOPRIVATE=github.com/enthus-appdev/* go install github.com/enthus-appdev/esq-cli/cmd/esq@latest

# Or clone and build
git clone git@github.com:enthus-appdev/esq-cli.git
cd esq-cli
make install   # builds and copies to ~/bin/
```

Ensure `~/go/bin` or `~/bin` is in your `PATH`.

## Setup

Add your Elasticsearch environments:

```bash
esq config add prod  --url http://10.11.20.41:9200
esq config add stage --url http://10.11.20.44:9200
esq config add local --url http://localhost:29200
esq config use prod
```

Config is stored at `~/.config/esq/config.json`.

## Usage

### Searching

```bash
# Lucene query string syntax
esq search documents "DocumentNo:12345"
esq search customer "CustomerName:Müller" --size 50
esq search documents "DocumentStateId:2 AND CustomerName:enthus"

# Filter returned fields
esq search documents "DocumentNo:12345" --source DocumentNo,DocumentStateId

# Full Query DSL
esq query documents '{"query":{"term":{"DocumentNo":12345}},"_source":["DocumentNo","DocumentStateId"]}'
```

### Getting & Counting

```bash
esq get documents offer-122116      # Get by _id
esq count customer                  # Count all docs
esq count documents "Status:active" # Count matching
```

### Cluster Info

```bash
esq health              # Cluster health + node stats
esq indices             # List all indices
esq indices sales       # Filter by name
esq mapping documents   # Show field mapping
```

### Environment Override

```bash
esq health --env stage             # One-off override without switching
esq search documents "test" -e local
```

### Index Name Resolution

You don't need to type full versioned index names. Partial names auto-resolve to the latest version:

| You type | Resolves to |
|----------|-------------|
| `documents` | `erp.sales.documents_v30.1.0` |
| `customer` | `crm.customer_v15.0.0` |
| `items` | `erp.items_v15.0.0` |
| `servicetickets` | `jira.servicetickets_v4.0.0` |

If multiple indices match, the latest version is used and alternatives are shown.

### Piping

Info messages go to stderr, data to stdout — safe for piping:

```bash
esq search documents "DocumentNo:12345" | jq '.hits.hits[]._source'
esq count customer 2>/dev/null | jq .count
```

## Shell Completion

```bash
# Zsh
esq completion zsh > "${fpath[1]}/_esq"

# Bash
esq completion bash > /etc/bash_completion.d/esq

# Fish
esq completion fish > ~/.config/fish/completions/esq.fish
```

## Development

```bash
make build     # Build to bin/esq
make install   # Build + copy to ~/bin/
make lint      # goimports + golangci-lint
make test      # Run tests
```
