package kiali

import (
	"context"
	"net/url"
	"strings"
)

// MeshStatus calls the Kiali mesh graph API to get the status of mesh components.
// This returns information about mesh components like Istio, Kiali, Grafana, Prometheus
// and their interactions, versions, and health status.
func (k *Kiali) MeshStatus(ctx context.Context) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/mesh/graph"

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("includeGateways", "false")
	q.Set("includeWaypoints", "false")
	u.RawQuery = q.Encode()
	endpoint = u.String()

	return k.executeRequest(ctx, endpoint)
}
