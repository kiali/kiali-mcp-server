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

func initWorkloads() []api.ServerTool {
	ret := make([]api.ServerTool, 0)

	// Workloads list tool
	ret = append(ret, api.ServerTool{
		Tool: api.Tool{
			Name:        "workloads_list",
			Description: "Get all workloads in the mesh across specified namespaces with health and Istio resource information",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"namespaces": {
						Type:        "string",
						Description: "Comma-separated list of namespaces to get workloads from (e.g. 'bookinfo' or 'bookinfo,default'). If not provided, will list workloads from all accessible namespaces",
					},
				},
			},
			Annotations: api.ToolAnnotations{
				Title:           "Workloads: List",
				ReadOnlyHint:    ptr.To(true),
				DestructiveHint: ptr.To(false),
				IdempotentHint:  ptr.To(true),
				OpenWorldHint:   ptr.To(true),
			},
		}, Handler: workloadsListHandler,
	})

	// Workload details tool
	ret = append(ret, api.ServerTool{
		Tool: api.Tool{
			Name:        "workload_details",
			Description: "Get detailed information for a specific workload in a namespace, including validation, health status, and configuration",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"namespace": {
						Type:        "string",
						Description: "Namespace containing the workload",
					},
					"workload": {
						Type:        "string",
						Description: "Name of the workload to get details for",
					},
				},
				Required: []string{"namespace", "workload"},
			},
			Annotations: api.ToolAnnotations{
				Title:           "Workload: Details",
				ReadOnlyHint:    ptr.To(true),
				DestructiveHint: ptr.To(false),
				IdempotentHint:  ptr.To(true),
				OpenWorldHint:   ptr.To(true),
			},
		}, Handler: workloadDetailsHandler,
	})

	return ret
}

func workloadsListHandler(params api.ToolHandlerParams) (*api.ToolCallResult, error) {
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

	content, err := kialiClient.WorkloadsList(params.Context, authHeader, namespaces)
	if err != nil {
		return api.NewToolCallResult("", fmt.Errorf("failed to list workloads: %v", err)), nil
	}
	return api.NewToolCallResult(content, nil), nil
}

func workloadDetailsHandler(params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	// Extract parameters
	namespace, _ := params.GetArguments()["namespace"].(string)
	workload, _ := params.GetArguments()["workload"].(string)

	if namespace == "" {
		return api.NewToolCallResult("", fmt.Errorf("namespace parameter is required")), nil
	}
	if workload == "" {
		return api.NewToolCallResult("", fmt.Errorf("workload parameter is required")), nil
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

	content, err := kialiClient.WorkloadDetails(params.Context, authHeader, namespace, workload)
	if err != nil {
		return api.NewToolCallResult("", fmt.Errorf("failed to get workload details: %v", err)), nil
	}
	return api.NewToolCallResult(content, nil), nil
}
