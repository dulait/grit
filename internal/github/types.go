package github

import "time"

type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

type CreateCommentRequest struct {
	Body string `json:"body"`
}

type IssueComment struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors,omitempty"`
}
