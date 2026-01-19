package llm

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

type GeneratedIssue struct {
	Title     string
	Body      string
	Labels    []string
	Reasoning string
}
