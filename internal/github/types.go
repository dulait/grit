package github

import "time"

type Label struct {
	Name string `json:"name"`
}

type User struct {
	Login string `json:"login"`
}

type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	Labels    []Label   `json:"labels"`
	Assignees []User    `json:"assignees"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateIssueRequest struct {
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

type ListIssuesRequest struct {
	State    string
	Assignee string
	Labels   string
	PerPage  int
	Page     int
}

type SearchIssuesRequest struct {
	Query   string
	State   string
	Labels  string
	PerPage int
	Page    int
}

type SearchIssuesResponse struct {
	TotalCount int     `json:"total_count"`
	Items      []Issue `json:"items"`
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

type UpdateIssueRequest struct {
	Title     *string  `json:"title,omitempty"`
	Body      *string  `json:"body,omitempty"`
	State     *string  `json:"state,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Resource string `json:"resource"`
		Field    string `json:"field"`
		Code     string `json:"code"`
	} `json:"errors,omitempty"`
}
