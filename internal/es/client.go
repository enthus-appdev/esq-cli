package es

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Client is an Elasticsearch HTTP client.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Elasticsearch client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// request executes an HTTP request and returns the raw response body.
func (c *Client) request(method, path string, body io.Reader) ([]byte, error) {
	reqURL := c.baseURL + path

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ES error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// requestRaw executes a request and returns the raw string response (for _cat API).
func (c *Client) requestRaw(method, path string) (string, error) {
	data, err := c.request(method, path, nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Health returns cluster health information.
func (c *Client) Health() (json.RawMessage, error) {
	data, err := c.request("GET", "/_cluster/health", nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// Nodes returns node stats as a formatted table string.
func (c *Client) Nodes() (string, error) {
	return c.requestRaw("GET", "/_cat/nodes?v&h=name,ip,heap.percent,ram.percent,cpu,load_1m,disk.used_percent,node.role")
}

// CatIndices returns the index listing, optionally filtered.
func (c *Client) CatIndices(filter string) (string, error) {
	raw, err := c.requestRaw("GET", "/_cat/indices?v&h=index,health,docs.count,store.size&s=index")
	if err != nil {
		return "", err
	}

	if filter == "" {
		return filterInternalIndices(raw), nil
	}

	lines := strings.Split(raw, "\n")
	var result []string
	for i, line := range lines {
		if i == 0 {
			result = append(result, line) // header
			continue
		}
		if strings.Contains(strings.ToLower(line), strings.ToLower(filter)) && !strings.HasPrefix(strings.TrimSpace(line), ".") {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n"), nil
}

// filterInternalIndices removes .* indices from _cat output.
func filterInternalIndices(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string
	for i, line := range lines {
		if i == 0 || !strings.HasPrefix(strings.TrimSpace(line), ".") {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// ResolveIndex resolves a partial index name to its full versioned name.
// If the input contains "_v", it's treated as a full name.
// Otherwise, it searches for matching indices and returns the latest version.
func (c *Client) ResolveIndex(input string) (string, error) {
	if strings.Contains(input, "_v") {
		return input, nil
	}

	raw, err := c.requestRaw("GET", "/_cat/indices?h=index")
	if err != nil {
		return "", fmt.Errorf("listing indices: %w", err)
	}

	var matches []string
	for _, line := range strings.Split(raw, "\n") {
		name := strings.TrimSpace(line)
		if name == "" || strings.HasPrefix(name, ".") {
			continue
		}
		if strings.Contains(strings.ToLower(name), strings.ToLower(input)) {
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no index matching %q (use 'esq indices' to list available indices)", input)
	}

	sort.Strings(matches)
	resolved := matches[len(matches)-1]

	return resolved, nil
}

// ResolveIndexVerbose resolves an index and returns info about ambiguity.
func (c *Client) ResolveIndexVerbose(input string) (resolved string, wasPartial bool, alternatives []string, err error) {
	if strings.Contains(input, "_v") {
		return input, false, nil, nil
	}

	raw, err := c.requestRaw("GET", "/_cat/indices?h=index")
	if err != nil {
		return "", false, nil, fmt.Errorf("listing indices: %w", err)
	}

	var matches []string
	for _, line := range strings.Split(raw, "\n") {
		name := strings.TrimSpace(line)
		if name == "" || strings.HasPrefix(name, ".") {
			continue
		}
		if strings.Contains(strings.ToLower(name), strings.ToLower(input)) {
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		return "", false, nil, fmt.Errorf("no index matching %q (use 'esq indices' to list available indices)", input)
	}

	sort.Strings(matches)
	resolved = matches[len(matches)-1]

	return resolved, len(matches) > 1, matches[:len(matches)-1], nil
}

// Search performs a Lucene query string search.
func (c *Client) Search(index, query string, size int) (json.RawMessage, error) {
	path := fmt.Sprintf("/%s/_search?q=%s&size=%d", url.PathEscape(index), url.QueryEscape(query), size)
	data, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// SearchDSL performs a full Query DSL search.
func (c *Client) SearchDSL(index string, body io.Reader) (json.RawMessage, error) {
	path := fmt.Sprintf("/%s/_search", url.PathEscape(index))
	data, err := c.request("POST", path, body)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// Get retrieves a single document by ID.
func (c *Client) Get(index, docID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/%s/_doc/%s", url.PathEscape(index), url.PathEscape(docID))
	data, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// Count returns the document count, optionally filtered by query.
func (c *Client) Count(index, query string) (json.RawMessage, error) {
	path := fmt.Sprintf("/%s/_count", url.PathEscape(index))
	if query != "" {
		path += "?q=" + url.QueryEscape(query)
	}
	data, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}

// Mapping returns the index mapping.
func (c *Client) Mapping(index string) (json.RawMessage, error) {
	path := fmt.Sprintf("/%s/_mapping", url.PathEscape(index))
	data, err := c.request("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(data), nil
}
