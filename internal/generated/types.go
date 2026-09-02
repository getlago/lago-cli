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
	// Mutation marks a create or update whose default table output is the terse
	// identifier block. Full detail stays available through --output json|yaml.
	Mutation bool `json:"mutation"`
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
	// Required mirrors requestBody.required from the spec. When false the command may be
	// run with no body flags at all: `invoices void` voids without a body, for instance.
	Required bool `json:"required"`
}

type Field struct {
	Path     []string `json:"path"`
	Flag     string   `json:"flag"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	// Nullable records a `type: [T, 'null']` union. A nullable field is never Required:
	// the spec listing it under `required` means the key must be present, and null
	// satisfies that, so the CLI must not demand a value.
	Nullable    bool     `json:"nullable,omitempty"`
	Complex     bool     `json:"complex"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}
