package generated

import "encoding/json"

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
	// Mutation marks a write whose default table output is the terse identifier block
	// (identifiers plus status). Full detail stays available through --output json|yaml.
	Mutation bool `json:"mutation"`
	// CursorPaginated marks a list that pages with opaque `after`/`before` cursors and
	// reports `meta.next_cursor`, the API v2 scheme, instead of page/per_page.
	CursorPaginated bool `json:"cursor_paginated,omitempty"`
	// APIVersion is the API the path is relative to. It is empty for /api/v1 and "v2"
	// for operations generated from the v2 document, which run under `lago v2`.
	APIVersion string `json:"api_version,omitempty"`
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

// mustDecodeOperations decodes an embedded operation table. The JSON is written by
// internal/gen from the pinned spec, so a decode failure is a build defect and panics.
func mustDecodeOperations(data string) []Operation {
	var operations []Operation
	if err := json.Unmarshal([]byte(data), &operations); err != nil {
		panic("decode generated Lago operations: " + err.Error())
	}
	return operations
}
