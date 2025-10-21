package kiali

import (
	"slices"

	"github.com/kiali/kiali-mcp-server/pkg/api"
	internalk8s "github.com/kiali/kiali-mcp-server/pkg/kubernetes"
	"github.com/kiali/kiali-mcp-server/pkg/toolsets"
)

type Toolset struct{}

var _ api.Toolset = (*Toolset)(nil)

func (t *Toolset) GetName() string {
	return "kiali"
}

func (t *Toolset) GetDescription() string {
	return "Most common tools for managing Kiali"
}

func (t *Toolset) GetTools(_ internalk8s.Openshift) []api.ServerTool {
	return slices.Concat(
		initGraph(),
		initMeshStatus(),
		initIstioConfig(),
		initIstioObjectDetails(),
		initIstioObjectPatch(),
		initIstioObjectCreate(),
		initIstioObjectDelete(),
		initValidations(),
		initNamespaces(),
		initServices(),
		initWorkloads(),
		initHealth(),
		initLogs(),
	)
}

func init() {
	toolsets.Register(&Toolset{})
}
