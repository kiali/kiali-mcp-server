# Deploying Kiali MCP Server with OpenShift Lightspeed

This guide explains how to deploy the Kiali MCP Server as an MCP server within OpenShift Lightspeed, enabling AI-powered service mesh management directly in your OpenShift cluster.

## Prerequisites

Before deploying, ensure you have:

- **OpenShift Cluster**: Running OpenShift 4.6+ with cluster admin access
- **Istio Service Mesh**: Installed and configured with Kiali
- **OpenShift Lightspeed**: Version 1.0.5+ installed from OperatorHub
- **LLM Provider Access**: OpenAI API key or compatible provider credentials
- **Cluster Monitoring**: Enabled for user-defined projects

### Enable Cluster Monitoring

If using CodeReady Containers (CRC), enable monitoring before starting your cluster:

```bash
crc config set enable-cluster-monitoring true
```

For other OpenShift installations, follow the [official documentation](https://docs.redhat.com/en/documentation/openshift_container_platform/4.6/html/monitoring/enabling-monitoring-for-user-defined-projects) to enable monitoring.

## Deployment Steps

### 1. Create the OpenShift Lightspeed Namespace

```bash
oc create ns openshift-lightspeed
```

### 2. Store Your LLM Provider API Key

Create a secret with your API credentials:

```bash
oc create secret generic -n openshift-lightspeed credentials --from-literal=apitoken=<YOUR_API_KEY>
```

Replace `<YOUR_API_KEY>` with your actual API key from your chosen LLM provider.

### 3. Deploy the Kiali MCP Server

Apply the Kubernetes manifests to deploy the MCP server:

```bash
oc apply -f lightspeed/service_account.yaml
oc apply -f lightspeed/deployment.yaml
oc apply -f lightspeed/mcp_service.yaml
```

This creates:
- **Service Account**: With necessary RBAC permissions for cluster access
- **Deployment**: Single pod running the Kiali MCP server
- **Service**: Exposes the MCP server on port 8080

### 4. Install OpenShift Lightspeed Operator

Install OpenShift Lightspeed 1.0.5+ from OperatorHub:

```bash
oc apply -f lightspeed/operator.yaml
```

Wait for the operator to be installed and running before proceeding.

### 5. Configure OpenShift Lightspeed

Configure Lightspeed with your LLM provider and register the Kiali MCP server:

```bash
oc apply -f - <<EOF
apiVersion: ols.openshift.io/v1alpha1
kind: OLSConfig
metadata:
  name: cluster
spec:
  featureGates:
  - MCPServer
  llm:
    providers:
    - name: myGeminiOpenai
      type: openai
      credentialsSecretRef:
        name: credentials
      url: https://generativelanguage.googleapis.com/v1beta/openai/
      models:
      - name: gemini-2.5-flash
  mcpServers:
  - name: kiali-mcp
    streamableHTTP:
      enableSSE: false
      sseReadTimeout: 10
      timeout: 5
      url: 'http://kiali-mcp-istio-system.apps-crc.testing/mcp'
  ols:
    defaultModel: gemini-2.5-flash
    defaultProvider: myGeminiOpenai
EOF
```

## Supported LLM Providers

The Kiali MCP Server works with any OpenAI-compatible API. Here are tested configurations:

### OpenAI (GPT Models)

```yaml
llm:
  providers:
  - name: myOpenai
    type: openai
    credentialsSecretRef:
      name: credentials
    url: https://api.openai.com/v1
    models:
    - name: gpt-4o-mini
```

### Google Gemini

```yaml
llm:
  providers:
  - name: myGeminiOpenai
    type: openai
    credentialsSecretRef:
      name: credentials
    url: https://generativelanguage.googleapis.com/v1beta/openai/
    models:
    - name: gemini-2.0-flash-exp
```

### Tested Models

- ✅ `gpt-4o-mini`
- ✅ `gemini-2.0-flash-exp`
- ✅ `gemini-2.5-flash`

## Configuration Details

### MCP Server Configuration

The Kiali MCP server is configured with:

- **Toolsets**: Only Kiali tools enabled (`--toolsets kiali`)
- **Port**: HTTP server on port 8080
- **Kiali URL**: Auto-detected or explicitly configured
- **Security**: Non-root user with minimal privileges

### Service Account Permissions

The service account includes:

- **Cluster Access**: Read access to cluster resources
- **Auth Delegation**: Required for OpenShift integration
- **Console Access**: Read access to OpenShift console configuration

## Verification

### Check Deployment Status

```bash
# Verify the MCP server pod is running
oc get pods -n istio-system -l app.kubernetes.io/name=kiali-mcp-server

# Check service is available
oc get svc -n istio-system kiali-mcp-server

# Verify OpenShift Lightspeed configuration
oc get olsconfig cluster -n openshift-lightspeed
```

### Test MCP Server Connectivity

```bash
# Port-forward to test locally
oc port-forward -n istio-system svc/kiali-mcp-server 8080:8080

# Test the MCP endpoint
curl http://localhost:8080/mcp
```

## Troubleshooting

### Common Issues

**Configuration Not Applied**
- Wait a few minutes for the configuration to propagate
- Check OpenShift Lightspeed operator logs: `oc logs -n openshift-lightspeed deployment/lightspeed-operator`

**MCP Server Not Accessible**
- Verify the service URL in the OLSConfig matches your cluster's routing
- Check pod logs: `oc logs -n istio-system deployment/kiali-mcp-server`
- Ensure Istio sidecar injection is working correctly

**LLM Provider Errors**
- Verify API key is correctly stored in the secret
- Check provider URL and model names are correct
- Test API connectivity outside of OpenShift Lightspeed

**Permission Issues**
- Ensure service account has proper RBAC permissions
- Verify cluster monitoring is enabled
- Check if Istio and Kiali are properly installed

### Logs and Debugging

```bash
# Check MCP server logs
oc logs -n istio-system deployment/kiali-mcp-server

# Check OpenShift Lightspeed operator logs
oc logs -n openshift-lightspeed deployment/lightspeed-operator

# Check OLSConfig status
oc describe olsconfig cluster -n openshift-lightspeed
```

## Next Steps

Once deployed, you can:

1. **Access OpenShift Lightspeed**: Open the chat interface in your OpenShift console
2. **Ask Service Mesh Questions**: Use natural language to query your Istio service mesh
3. **Manage Istio Resources**: Create, modify, and validate Istio configuration objects
4. **Monitor Mesh Health**: Get real-time insights into your service mesh topology

The AI assistant will have access to all Kiali MCP server tools, enabling comprehensive service mesh management through conversational interfaces.
