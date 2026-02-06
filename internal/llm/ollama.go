package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// OllamaClient implements Client using a local Ollama server.
type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// CheckOllamaConnection verifies that an Ollama server is reachable.
func CheckOllamaConnection(baseURL string) error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return fmt.Errorf("cannot reach Ollama at %s: %w", baseURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status %d at %s", resp.StatusCode, baseURL)
	}
	return nil
}

// NewOllamaClient creates a new Ollama API client.
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	System string `json:"system,omitempty"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func (c *OllamaClient) GenerateIssue(ctx context.Context, req IssueRequest) (*GeneratedIssue, error) {
	var titleInstruction string
	if req.TitleHint != "" {
		titleInstruction = fmt.Sprintf("Use this as the title (keep it concise, under 80 chars): %s", req.TitleHint)
	} else if req.GenerateTitle {
		titleInstruction = "Generate a clear, concise title (under 80 characters)"
	}

	var bodyInstruction string
	if req.DescriptionHint != "" {
		bodyInstruction = fmt.Sprintf(`Expand and structure the following into a well-formatted GitHub issue body:
"%s"

Include relevant sections such as:
- Description (what needs to be done)
- Acceptance Criteria (if applicable)
- Technical Notes (if applicable)

Use markdown formatting.`, req.DescriptionHint)
	} else if req.UserPrompt != "" {
		bodyInstruction = `Generate a well-structured GitHub issue body with:
- Description
- Acceptance Criteria (if applicable)
- Technical Notes (if applicable)

Use markdown formatting.`
	} else {
		bodyInstruction = "Generate a brief issue body based on the title."
	}

	labelInstruction := "Set labels to an empty array []."
	if req.SuggestLabels && len(req.AllowedLabels) > 0 {
		labelInstruction = fmt.Sprintf("Suggest labels ONLY from: %v. If none fit, use empty array.", req.AllowedLabels)
	}

	systemPrompt := fmt.Sprintf(`You are a GitHub issue generator that creates well-structured issues.

%s

%s

%s

Respond with ONLY valid JSON:
{
  "title": "Issue title",
  "body": "Markdown formatted body",
  "labels": []
}`, titleInstruction, bodyInstruction, labelInstruction)

	userMessage := req.UserPrompt
	if userMessage == "" && req.DescriptionHint != "" {
		userMessage = req.DescriptionHint
	}
	if userMessage == "" && req.TitleHint != "" {
		userMessage = req.TitleHint
	}
	if userMessage == "" {
		userMessage = "Generate a GitHub issue."
	}

	resp, err := c.call(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, err
	}

	issue, err := parseGeneratedIssue(resp)
	if err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w", err)
	}

	if req.IssuePrefix != "" && issue.Title != "" {
		issue.Title = req.IssuePrefix + issue.Title
	}

	if !req.SuggestLabels {
		issue.Labels = nil
	} else if len(req.AllowedLabels) > 0 {
		issue.Labels = filterLabels(issue.Labels, req.AllowedLabels)
	}

	return issue, nil
}

func (c *OllamaClient) GenerateComment(ctx context.Context, issueContext, userPrompt string) (string, error) {
	systemPrompt := `Write a GitHub issue comment based on the user's intent. Respond with ONLY the comment text.`
	userMessage := fmt.Sprintf("Issue context:\n%s\n\nWrite a comment that: %s", issueContext, userPrompt)
	return c.call(ctx, systemPrompt, userMessage)
}

func (c *OllamaClient) call(ctx context.Context, system, prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		System: system,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}

func parseGeneratedIssue(s string) (*GeneratedIssue, error) {
	title := extractField(s, "title")
	body := extractField(s, "body")
	labels := extractLabels(s)

	if title == "" {
		return nil, fmt.Errorf("could not extract title from response: %s", s)
	}

	return &GeneratedIssue{
		Title:  title,
		Body:   body,
		Labels: labels,
	}, nil
}

func extractField(s, field string) string {
	pattern := fmt.Sprintf(`"%s"\s*:\s*"`, field)
	re := regexp.MustCompile(pattern)
	loc := re.FindStringIndex(s)
	if loc == nil {
		pattern = fmt.Sprintf(`"%s"\s*:\s*"""`, field)
		re = regexp.MustCompile(pattern)
		loc = re.FindStringIndex(s)
		if loc == nil {
			return ""
		}
	}

	start := loc[1]
	end := findStringEnd(s, start)
	if end == -1 {
		return ""
	}

	value := s[start:end]
	value = strings.ReplaceAll(value, `\n`, "\n")
	value = strings.ReplaceAll(value, `\r`, "\r")
	value = strings.ReplaceAll(value, `\t`, "\t")
	value = strings.ReplaceAll(value, `\"`, `"`)

	return strings.TrimSpace(value)
}

func findStringEnd(s string, start int) int {
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			if i+1 < len(s) && s[i+1] == '"' {
				i++
				continue
			}
			return i
		}
	}
	return -1
}

func extractLabels(s string) []string {
	pattern := `"labels"\s*:\s*\[([^\]]*)\]`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(s)
	if match == nil || len(match) < 2 {
		return nil
	}

	labelsStr := match[1]
	labelPattern := `"([^"]*)"`
	labelRe := regexp.MustCompile(labelPattern)
	matches := labelRe.FindAllStringSubmatch(labelsStr, -1)

	var labels []string
	for _, m := range matches {
		if len(m) >= 2 && m[1] != "" {
			labels = append(labels, m[1])
		}
	}

	return labels
}
