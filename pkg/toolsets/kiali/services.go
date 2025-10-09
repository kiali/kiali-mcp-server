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

func initServices() []api.ServerTool {
	ret := make([]api.ServerTool, 0)

	// Services list tool
	ret = append(ret, api.ServerTool{
		Tool: api.Tool{
			Name:        "services_list",
			Description: "Get all services in the mesh across specified namespaces with health and Istio resource information",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"namespaces": {
						Type:        "string",
						Description: "Comma-separated list of namespaces to get services from (e.g. 'bookinfo' or 'bookinfo,default'). If not provided, will list services from all accessible namespaces",
					},
				},
			},
			Annotations: api.ToolAnnotations{
				Title:           "Services: List",
				ReadOnlyHint:    ptr.To(true),
				DestructiveHint: ptr.To(false),
				IdempotentHint:  ptr.To(true),
				OpenWorldHint:   ptr.To(true),
			},
		}, Handler: servicesListHandler,
	})

	// Service details tool
	ret = append(ret, api.ServerTool{
		Tool: api.Tool{
			Name:        "service_details",
			Description: "Get detailed information for a specific service in a namespace, including validation, health status, and configuration",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"namespace": {
						Type:        "string",
						Description: "Namespace containing the service",
					},
					"service": {
						Type:        "string",
						Description: "Name of the service to get details for",
					},
				},
				Required: []string{"namespace", "service"},
			},
			Annotations: api.ToolAnnotations{
				Title:           "Service: Details",
				ReadOnlyHint:    ptr.To(true),
				DestructiveHint: ptr.To(false),
				IdempotentHint:  ptr.To(true),
				OpenWorldHint:   ptr.To(true),
			},
		}, Handler: serviceDetailsHandler,
	})

	return ret
}

func servicesListHandler(params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	// Extract parameters
	namespaces, _ := params.GetArguments()["namespaces"].(string)

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

	content, err := kialiClient.ServicesList(params.Context, authHeader, namespaces)
	if err != nil {
		return api.NewToolCallResult("", fmt.Errorf("failed to list services: %v", err)), nil
	}
	return api.NewToolCallResult(content, nil), nil
}

func serviceDetailsHandler(params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	// Extract parameters
	namespace, _ := params.GetArguments()["namespace"].(string)
	service, _ := params.GetArguments()["service"].(string)

	if namespace == "" {
		return api.NewToolCallResult("", fmt.Errorf("namespace parameter is required")), nil
	}
	if service == "" {
		return api.NewToolCallResult("", fmt.Errorf("service parameter is required")), nil
	}

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

	content, err := kialiClient.ServiceDetails(params.Context, authHeader, namespace, service)
	if err != nil {
		return api.NewToolCallResult("", fmt.Errorf("failed to get service details: %v", err)), nil
	}
	return api.NewToolCallResult(content, nil), nil
}
