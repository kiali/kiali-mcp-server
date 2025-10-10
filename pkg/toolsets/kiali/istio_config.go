package kiali

import (
	"fmt"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"k8s.io/utils/ptr"

	"github.com/kiali/kiali-mcp-server/pkg/api"
	internalkiali "github.com/kiali/kiali-mcp-server/pkg/kiali"
	internalk8s "github.com/kiali/kiali-mcp-server/pkg/kubernetes"
)

func initIstioConfig() []api.ServerTool {
	ret := make([]api.ServerTool, 0)
	ret = append(ret, api.ServerTool{
		Tool: api.Tool{
			Name:        "istio_config",
			Description: "Get all Istio configuration objects in the mesh including their full YAML resources and details",
			InputSchema: &jsonschema.Schema{
				Type:       "object",
				Properties: map[string]*jsonschema.Schema{},
				Required:   []string{},
			},
			Annotations: api.ToolAnnotations{
				Title:           "Istio Config: List All",
				ReadOnlyHint:    ptr.To(true),
				DestructiveHint: ptr.To(false),
				IdempotentHint:  ptr.To(true),
				OpenWorldHint:   ptr.To(true),
			},
		}, Handler: istioConfigHandler,
	})
	return ret
}

func istioConfigHandler(params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	// Extract the Authorization header from context
	authHeader, _ := params.Context.Value(internalk8s.OAuthAuthorizationHeader).(string)
	if strings.TrimSpace(authHeader) == "" {
		// Fall back to using the same token that the Kubernetes client is using
		if params.Kubernetes != nil {
			authHeader = params.Kubernetes.CurrentAuthorizationHeader()
		}
	}
	// Build a Kiali client from static config
	kialiClient := internalkiali.NewFromConfig(params.Kubernetes.StaticConfig())

	content, err := kialiClient.IstioConfig(params.Context, authHeader)
	if err != nil {
		return api.NewToolCallResult("", fmt.Errorf("failed to retrieve Istio configuration: %v", err)), nil
	}
	return api.NewToolCallResult(content, nil), nil
}

func initIstioObjectDetails() []api.ServerTool {
	ret := make([]api.ServerTool, 0)
	ret = append(ret, api.ServerTool{
		Tool: api.Tool{
			Name:        "istio_object_details",
			Description: "Get detailed information about a specific Istio object including validation and help information",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"namespace": {
						Type:        "string",
						Description: "Namespace containing the Istio object",
					},
					"group": {
						Type:        "string",
						Description: "API group of the Istio object (e.g., 'networking.istio.io', 'gateway.networking.k8s.io')",
					},
					"version": {
						Type:        "string",
						Description: "API version of the Istio object (e.g., 'v1', 'v1beta1')",
					},
					"kind": {
						Type:        "string",
						Description: "Kind of the Istio object (e.g., 'DestinationRule', 'VirtualService', 'HTTPRoute', 'Gateway')",
					},
					"name": {
						Type:        "string",
						Description: "Name of the Istio object",
					},
				},
				Required: []string{"namespace", "group", "version", "kind", "name"},
			},
			Annotations: api.ToolAnnotations{
				Title:           "Istio Object: Details",
				ReadOnlyHint:    ptr.To(true),
				DestructiveHint: ptr.To(false),
				IdempotentHint:  ptr.To(true),
				OpenWorldHint:   ptr.To(true),
			},
		}, Handler: istioObjectDetailsHandler,
	})
	return ret
}

func istioObjectDetailsHandler(params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	// Extract the Authorization header from context
	authHeader, _ := params.Context.Value(internalk8s.OAuthAuthorizationHeader).(string)
	if strings.TrimSpace(authHeader) == "" {
		// Fall back to using the same token that the Kubernetes client is using
		if params.Kubernetes != nil {
			authHeader = params.Kubernetes.CurrentAuthorizationHeader()
		}
	}
	// Build a Kiali client from static config
	kialiClient := internalkiali.NewFromConfig(params.Kubernetes.StaticConfig())

	// Extract required parameters
	namespace, _ := params.GetArguments()["namespace"].(string)
	group, _ := params.GetArguments()["group"].(string)
	version, _ := params.GetArguments()["version"].(string)
	kind, _ := params.GetArguments()["kind"].(string)
	name, _ := params.GetArguments()["name"].(string)

	content, err := kialiClient.IstioObjectDetails(params.Context, authHeader, namespace, group, version, kind, name)
	if err != nil {
		return api.NewToolCallResult("", fmt.Errorf("failed to retrieve Istio object details: %v", err)), nil
	}
	return api.NewToolCallResult(content, nil), nil
}
