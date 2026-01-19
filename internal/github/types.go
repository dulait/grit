package github

import "time"

// Issue represents a GitHub issue.
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateIssueRequest contains the parameters for creating a new issue.
type CreateIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

// CreateCommentRequest contains the parameters for adding a comment.
type CreateCommentRequest struct {
	Body string `json:"body"`
}

// IssueComment represents a comment on a GitHub issue.
type IssueComment struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorResponse represents an error returned by the GitHub API.
type ErrorResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors,omitempty"`
}
