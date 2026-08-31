package generated

type Operation struct {
	Resource    string      `json:"resource"`
	Action      string      `json:"action"`
	OperationID string      `json:"operation_id"`
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Summary     string      `json:"summary"`
	Description string      `json:"description,omitempty"`
	DocsURL     string      `json:"docs_url,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"`
	Body        *Body       `json:"body,omitempty"`
	Idempotent  bool        `json:"idempotent"`
	Dangerous   bool        `json:"dangerous"`
	Paginated   bool        `json:"paginated"`
}

type Parameter struct {
	Name        string   `json:"name"`
	Flag        string   `json:"flag"`
	In          string   `json:"in"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type Body struct {
	Wrapper string  `json:"wrapper,omitempty"`
	Fields  []Field `json:"fields,omitempty"`
}

type Field struct {
	Path        []string `json:"path"`
	Flag        string   `json:"flag"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Complex     bool     `json:"complex"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}
