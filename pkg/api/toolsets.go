package api

import (
	"context"
	"encoding/json"

	internalKiali "github.com/containers/kubernetes-mcp-server/pkg/kiali"
	internalk8s "github.com/containers/kubernetes-mcp-server/pkg/kubernetes"
	"github.com/containers/kubernetes-mcp-server/pkg/output"
	"github.com/google/jsonschema-go/jsonschema"
)

type ServerTool struct {
	Tool    Tool
	Handler ToolHandlerFunc
}

type Toolset interface {
	// GetName returns the name of the toolset.
	// Used to identify the toolset in configuration, logs, and command-line arguments.
	// Examples: "core", "metrics", "helm"
	GetName() string
	GetDescription() string
	GetTools(o internalk8s.Openshift) []ServerTool
}

type ToolCallRequest interface {
	GetArguments() map[string]any
}

type ToolCallResult struct {
	// Raw content returned by the tool.
	Content string
	// Error (non-protocol) to send back to the LLM.
	Error error
}

func NewToolCallResult(content string, err error) *ToolCallResult {
	return &ToolCallResult{
		Content: content,
		Error:   err,
	}
}

type ToolHandlerParams struct {
	context.Context
	*internalk8s.Kubernetes
	*internalKiali.Kiali
	ToolCallRequest
	ListOutput output.Output
}

type ToolHandlerFunc func(params ToolHandlerParams) (*ToolCallResult, error)

type Tool struct {
	// The name of the tool.
	// Intended for programmatic or logical use, but used as a display name in past
	// specs or fallback (if title isn't present).
	Name string `json:"name"`
	// A human-readable description of the tool.
	//
	// This can be used by clients to improve the LLM's understanding of available
	// tools. It can be thought of like a "hint" to the model.
	Description string `json:"description,omitempty"`
	// Additional tool information.
	Annotations ToolAnnotations `json:"annotations"`
	// A JSON Schema object defining the expected parameters for the tool.
	InputSchema *jsonschema.Schema
}

type ToolAnnotations struct {
	// Human-readable title for the tool
	Title string `json:"title,omitempty"`
	// If true, the tool does not modify its environment.
	ReadOnlyHint *bool `json:"readOnlyHint,omitempty"`
	// If true, the tool may perform destructive updates to its environment. If
	// false, the tool performs only additive updates.
	//
	// (This property is meaningful only when ReadOnlyHint == false.)
	DestructiveHint *bool `json:"destructiveHint,omitempty"`
	// If true, calling the tool repeatedly with the same arguments will have no
	// additional effect on its environment.
	//
	// (This property is meaningful only when ReadOnlyHint == false.)
	IdempotentHint *bool `json:"idempotentHint,omitempty"`
	// If true, this tool may interact with an "open world" of external entities. If
	// false, the tool's domain of interaction is closed. For example, the world of
	// a web search tool is open, whereas that of a memory tool is not.
	OpenWorldHint *bool `json:"openWorldHint,omitempty"`
}

func ToRawMessage(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
