package kiali

import (
	"context"
	"strings"
)

// ListNamespaces calls the Kiali namespaces API using the provided Authorization header value.
// Returns all namespaces in the mesh that the user has access to.
func (k *Kiali) ListNamespaces(ctx context.Context, authHeader string) (string, error) {
	baseURL, err := k.validateAndGetBaseURL()
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/api/namespaces"

	return k.executeRequest(ctx, authHeader, endpoint)
}
