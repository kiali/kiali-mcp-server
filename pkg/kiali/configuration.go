package kiali

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// resolveKialiRequiredConfigurations resolves the required kiali configurations from Kubernetes
func resolveKialiRequiredConfigurations(kiali *Manager) error {
	// Always set clientCmdConfig
	pathOptions := clientcmd.NewDefaultPathOptions()
	if kiali.staticConfig.KubeConfig != "" {
		pathOptions.LoadingRules.ExplicitPath = kiali.staticConfig.KubeConfig
	}
	var err error
	kiali.clientCmdConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		pathOptions.LoadingRules,
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})
	// Out of cluster
	kiali.cfg, err = kiali.clientCmdConfig.ClientConfig()
	if kiali.cfg != nil && kiali.cfg.UserAgent == "" {
		kiali.cfg.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	return err
}
