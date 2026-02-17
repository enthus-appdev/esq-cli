# esq - Elasticsearch Query CLI

Query and inspect Elasticsearch clusters across environments from the command line.

## Installation

Requires Go 1.24+.

```bash
go install github.com/enthus-appdev/esq-cli/cmd/esq@latest
```

Ensure `~/go/bin` is in your `PATH`.

## Setup

Add your Elasticsearch environments:

```bash
esq config add prod  --url http://es-prod:9200
esq config add stage --url http://es-stage:9200
esq config add local --url http://localhost:9200
esq config use prod
```

Config is stored at `~/.config/esq/config.json`.

## Usage

### Searching

```bash
# Lucene query string syntax
esq search my-index "title:hello"
esq search users "name:John" --size 50
esq search logs "level:error AND service:api"

# Filter returned fields
esq search my-index "title:hello" --source title,status

# Full Query DSL
esq query my-index '{"query":{"term":{"title":"hello"}},"_source":["title","status"]}'
```

### Getting & Counting

```bash
esq get documents doc-42            # Get by _id
esq count users                     # Count all docs
esq count logs "level:error"        # Count matching
```

### Cluster Info

```bash
esq health              # Cluster health + node stats
esq indices             # List all indices
esq indices logs        # Filter by name
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
| `logs` | `logs_v3.0.0` |
| `users` | `users_v2.1.0` |
| `metrics` | `metrics_v1.5.0` |

If multiple indices match, the latest version is used and alternatives are shown.

### Piping

Info messages go to stderr, data to stdout — safe for piping:

```bash
esq search my-index "title:hello" | jq '.hits.hits[]._source'
esq count users 2>/dev/null | jq .count
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
go build ./cmd/esq              # Build the binary
go test -race ./...             # Run tests
golangci-lint run               # Lint
```

## License

MIT — see [LICENSE](LICENSE) for details.
