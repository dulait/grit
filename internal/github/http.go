package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// HTTPClient implements Client using the GitHub REST API.
type HTTPClient struct {
	baseURL    string
	token      string
	owner      string
	repo       string
	httpClient *http.Client
}

// NewHTTPClient creates a new GitHub API client for the specified repository.
func NewHTTPClient(owner, repo, token string) *HTTPClient {
	return &HTTPClient{
		baseURL: "https://api.github.com",
		token:   token,
		owner:   owner,
		repo:    repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPClient) do(ctx context.Context, method, path string, body, result any) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Message != "" {
			return fmt.Errorf("github api error (%d): %s", resp.StatusCode, errResp.Message)
		}
		return fmt.Errorf("github api error (%d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}

	return nil
}

func (c *HTTPClient) repoPath(format string, args ...any) string {
	prefix := fmt.Sprintf("/repos/%s/%s", c.owner, c.repo)
	return prefix + fmt.Sprintf(format, args...)
}

func (c *HTTPClient) ListIssues(ctx context.Context, req ListIssuesRequest) ([]Issue, error) {
	params := url.Values{}

	if req.State != "" {
		params.Set("state", req.State)
	}
	if req.Assignee != "" {
		params.Set("assignee", req.Assignee)
	}
	if req.Labels != "" {
		params.Set("labels", req.Labels)
	}
	if req.PerPage > 0 {
		params.Set("per_page", strconv.Itoa(req.PerPage))
	}
	if req.Page > 0 {
		params.Set("page", strconv.Itoa(req.Page))
	}

	path := c.repoPath("/issues")
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var issues []Issue
	if err := c.do(ctx, http.MethodGet, path, nil, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *HTTPClient) CreateIssue(ctx context.Context, req CreateIssueRequest) (*Issue, error) {
	var issue Issue
	path := c.repoPath("/issues")
	if err := c.do(ctx, http.MethodPost, path, req, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *HTTPClient) GetIssue(ctx context.Context, number int) (*Issue, error) {
	var issue Issue
	path := c.repoPath("/issues/%d", number)
	if err := c.do(ctx, http.MethodGet, path, nil, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *HTTPClient) CloseIssue(ctx context.Context, number int, comment string) (*Issue, error) {
	if comment != "" {
		if _, err := c.AddComment(ctx, number, comment); err != nil {
			return nil, fmt.Errorf("adding closing comment: %w", err)
		}
	}

	var issue Issue
	path := c.repoPath("/issues/%d", number)
	body := map[string]string{"state": "closed"}
	if err := c.do(ctx, http.MethodPatch, path, body, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *HTTPClient) AddComment(ctx context.Context, number int, body string) (*IssueComment, error) {
	var comment IssueComment
	path := c.repoPath("/issues/%d/comments", number)
	req := CreateCommentRequest{Body: body}
	if err := c.do(ctx, http.MethodPost, path, req, &comment); err != nil {
		return nil, err
	}
	return &comment, nil
}

func (c *HTTPClient) AssignIssue(ctx context.Context, number int, assignees []string) (*Issue, error) {
	var issue Issue
	path := c.repoPath("/issues/%d", number)
	body := map[string][]string{"assignees": assignees}
	if err := c.do(ctx, http.MethodPatch, path, body, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *HTTPClient) UpdateIssue(ctx context.Context, number int, req UpdateIssueRequest) (*Issue, error) {
	body := map[string]any{}
	if req.Title != nil {
		body["title"] = *req.Title
	}
	if req.Body != nil {
		body["body"] = *req.Body
	}
	if req.State != nil {
		body["state"] = *req.State
	}
	if req.Labels != nil {
		body["labels"] = req.Labels
	}
	if req.Assignees != nil {
		body["assignees"] = req.Assignees
	}

	var issue Issue
	path := c.repoPath("/issues/%d", number)
	if err := c.do(ctx, http.MethodPatch, path, body, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

func (c *HTTPClient) SearchIssues(ctx context.Context, req SearchIssuesRequest) (*SearchIssuesResponse, error) {
	qualifiers := []string{fmt.Sprintf("repo:%s/%s", c.owner, c.repo), "is:issue"}

	if req.State != "" && req.State != "all" {
		qualifiers = append(qualifiers, "state:"+req.State)
	}
	if req.Labels != "" {
		for _, l := range strings.Split(req.Labels, ",") {
			l = strings.TrimSpace(l)
			if l != "" {
				qualifiers = append(qualifiers, "label:"+l)
			}
		}
	}
	if req.Query != "" {
		qualifiers = append(qualifiers, req.Query)
	}

	params := url.Values{}
	params.Set("q", strings.Join(qualifiers, " "))
	if req.PerPage > 0 {
		params.Set("per_page", strconv.Itoa(req.PerPage))
	}
	if req.Page > 0 {
		params.Set("page", strconv.Itoa(req.Page))
	}

	path := "/search/issues?" + params.Encode()

	var resp SearchIssuesResponse
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
