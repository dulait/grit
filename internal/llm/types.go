package llm

import "strings"

// IssueRequest contains the parameters for generating an issue.
type IssueRequest struct {
	UserPrompt      string
	TitleHint       string
	DescriptionHint string
	RepoContext     string
	IssuePrefix     string
	AllowedLabels   []string
	GenerateTitle   bool
	GenerateBody    bool
	SuggestLabels   bool
}

// GeneratedIssue contains the LLM-generated issue content.
type GeneratedIssue struct {
	Title     string
	Body      string
	Labels    []string
	Reasoning string
}

// filterLabels returns only labels that exist in the allowed list.
func filterLabels(suggested, allowed []string) []string {
	allowedSet := make(map[string]bool, len(allowed))
	for _, l := range allowed {
		allowedSet[strings.ToLower(l)] = true
	}

	var filtered []string
	for _, l := range suggested {
		if allowedSet[strings.ToLower(l)] {
			filtered = append(filtered, l)
		}
	}
	return filtered
}
