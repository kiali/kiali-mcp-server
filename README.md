# Kiali MCP Server

[![GitHub License](https://img.shields.io/github/license/kiali/kiali-mcp-server)](https://github.com/kiali/kiali-mcp-server/blob/main/LICENSE)
[![npm](https://img.shields.io/npm/v/kiali-mcp-server)](https://www.npmjs.com/package/kiali-mcp-server)
[![PyPI - Version](https://img.shields.io/pypi/v/kiali-mcp-server)](https://pypi.org/project/kiali-mcp-server/)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/kiali/kiali-mcp-server?sort=semver)](https://github.com/kiali/kiali-mcp-server/releases/latest)
[![Build](https://github.com/kiali/kiali-mcp-server/actions/workflows/build.yaml/badge.svg)](https://github.com/kiali/kiali-mcp-server/actions/workflows/build.yaml)

https://github.com/user-attachments/assets/376f2137-43ee-4fe0-810d-39028a97ac47

**Kiali MCP Server** is a specialized Model Context Protocol (MCP) server that brings powerful Istio service mesh management capabilities to AI assistants like Claude, Cursor, and others. Built as an extension of the upstream Kubernetes MCP Server, it provides seamless integration with Kiali for service mesh observability and management.

- 🌐 **Native Kiali Integration**: Direct access to service mesh topology, validations, and health data
- 🔧 **Built on Kubernetes MCP**: Inherits all Kubernetes/OpenShift management capabilities  
- 🚀 **AI-First Design**: Optimized for AI assistant workflows and natural language interactions
- 📊 **Real-time Mesh Insights**: Live service mesh topology, traffic flows, and health status

For the complete set of Kubernetes tools and capabilities, see the upstream documentation: [openshift/openshift-mcp-server README](https://github.com/openshift/openshift-mcp-server/blob/main/README.md)

[✨ Features](#features) | [🚀 Getting Started](#getting-started) | [🎥 Demos](#demos) | [⚙️ Configuration](#configuration) | [🛠️ Tools](#tools-and-functionalities) | [🧑‍💻 Development](#development)

## ✨ Features <a id="features"></a>

### 🎯 Service Mesh Management
- **📊 Mesh Topology Visualization**: Real-time service graph with traffic flows, health status, and connectivity
- **🔍 Configuration Validation**: Comprehensive Istio object validation across namespaces
- **🌐 Multi-Namespace Support**: Work with single namespaces or multiple namespaces simultaneously
- **⚡ Live Data**: Direct integration with running Kiali instances for up-to-date mesh insights

### 🤖 AI-Optimized Experience  
- **Natural Language Queries**: Ask questions like "Check my bookinfo mesh status" or "Show validations for istio-system"
- **Intelligent Context**: Tools designed for AI understanding and optimal prompt engineering
- **Flexible Parameters**: Support both single and multiple namespace operations
- **Rich Responses**: Structured JSON data perfect for AI analysis and interpretation

### 🔧 Built on Kubernetes MCP
Inherits all capabilities from the upstream Kubernetes MCP Server including pod management, resource operations, Helm integration, and more.

## 🚀 Getting Started <a id="getting-started"></a>

### Requirements

- **Kubernetes/OpenShift Cluster**: Access via kubeconfig or in-cluster service account
- **Kiali Instance**: A running and accessible Kiali server 
- **Network Access**: Connectivity between the MCP server and your Kiali instance
- **AI Assistant**: Claude Desktop, Cursor, or any MCP-compatible AI tool

### Quick Start with Claude Desktop

https://github.com/user-attachments/assets/376f2137-43ee-4fe0-810d-39028a97ac47

#### Using npx (Recommended)

If you have npm installed, this is the fastest way to get started with Kiali MCP Server on Claude Desktop.

Open your `claude_desktop_config.json` and add the MCP server to the list of `mcpServers`:

```json
{
  "mcpServers": {
    "kiali-mcp-server": {
      "command": "npx",
      "args": [
        "-y",
        "kiali-mcp-server@latest"
      ]
    }
  }
}
```

### Quick Start with Cursor

Install the Kiali MCP server extension in Cursor by clicking the button below:

[![Install MCP Server](https://cursor.com/deeplink/mcp-install-dark.svg)](https://cursor.com/install-mcp?name=kiali-mcp-server&config=eyJjb21tYW5kIjoibnB4IC15IGtpYWxpLW1jcC1zZXJ2ZXJAbGF0ZXN0In0%3D)

Alternatively, you can install the extension manually by editing the `mcp.json` file:

```json
{
  "mcpServers": {
    "kiali-mcp-server": {
      "command": "npx",
      "args": ["-y", "kiali-mcp-server@latest"]
    }
  }
}
```

https://github.com/user-attachments/assets/d88a3b72-980c-43db-a69a-a19ad564cf49

### Configuring Kiali Connection

**Note:** You must specify the Kiali endpoint if the MCP server cannot auto-detect it. You may also need to configure TLS settings.

```json
{
  "mcpServers": {
    "kiali-mcp-server": {
      "command": "npx",
      "args": [
        "-y",
        "kiali-mcp-server@latest",
        "--kiali-server-url",
        "https://kiali-istio-system.apps-crc.testing/",
        "--kiali-insecure"
      ]
    }
  }
}
```

### Common Configuration Examples

<details>
<summary>🔒 Secure Kiali with Valid TLS</summary>

```json
{
  "mcpServers": {
    "kiali-mcp-server": {
      "command": "npx",
      "args": [
        "-y",
        "kiali-mcp-server@latest",
        "--kiali-server-url",
        "https://kiali.example.com/"
      ]
    }
  }
}
```
</details>

<details>
<summary>🔓 Local Development with Self-Signed Certificates</summary>

```json
{
  "mcpServers": {
    "kiali-mcp-server": {
      "command": "npx",
      "args": [
        "-y",
        "kiali-mcp-server@latest",
        "--kiali-server-url",
        "https://kiali-istio-system.apps-crc.testing/",
        "--kiali-insecure"
      ]
    }
  }
}
```
</details>

<details>
<summary>🎯 Kiali-Only Mode</summary>

```json
{
  "mcpServers": {
    "kiali-mcp-server": {
      "command": "npx",
      "args": [
        "-y",
        "kiali-mcp-server@latest",
        "--toolsets",
        "kiali",
        "--kiali-server-url",
        "https://kiali.example.com/"
      ]
    }
  }
}
```
</details>

<details>
<summary>🚀 OpenShift Lightspeed Integration</summary>

For deploying the Kiali MCP Server within OpenShift Lightspeed for AI-powered service mesh management directly in your OpenShift cluster, see the comprehensive guide:

**[📖 Deploying with OpenShift Lightspeed](docs/LIGHTSPEED.md)**

This guide covers:
- Prerequisites and cluster setup
- Step-by-step deployment instructions
- LLM provider configuration
- Troubleshooting and verification
- Integration with OpenShift Lightspeed chat interface

</details>

## ⚙️ Configuration <a id="configuration"></a>

Kiali MCP Server extends the upstream Kubernetes MCP Server with additional Kiali-specific configuration options.

### Kiali-Specific Flags

| Flag | Type | Description | Example |
|------|------|-------------|---------|
| `--kiali-server-url` | `string` | URL of the Kiali server | `https://kiali-istio-system.apps-crc.testing/` |
| `--kiali-insecure` | `boolean` | Skip TLS verification when connecting to Kiali | Use for self-signed certificates |

### Toolset Configuration

By default, both Kubernetes and Kiali tools are available. Use `--toolsets` to control which tool groups are enabled:

```bash
# Kiali tools only
--toolsets kiali

# All available toolsets (default)
--toolsets core,config,helm,kiali
```

### Command Line Examples

**Using npx:**
```bash
npx -y kiali-mcp-server@latest \
  --kiali-server-url "https://kiali-istio-system.apps-crc.testing/" \
  --kiali-insecure \
  --toolsets kiali
```

**Using compiled binary:**
```bash
./kiali-mcp-server \
  --kiali-server-url "https://kiali-istio-system.apps-crc.testing/" \
  --kiali-insecure \
  --port 8080
```

### Additional Configuration

For comprehensive configuration options including authentication, ports, read-only mode, and output formats, refer to the upstream documentation: [openshift/openshift-mcp-server README](https://github.com/openshift/openshift-mcp-server/blob/main/README.md)

## 🛠️ Tools and Functionalities <a id="tools-and-functionalities"></a>

The Kiali MCP server supports enabling or disabling specific groups of tools and functionalities (tools, resources, prompts, and so on) via the `--toolsets` command-line flag or `toolsets` configuration option.
This allows you to control which Kubernetes functionalities are available to your AI tools.
Enabling only the toolsets you need can help reduce the context size and improve the LLM's tool selection accuracy.

### Available Toolsets

The following sets of tools are available (only Kiali by default):

<!-- AVAILABLE-TOOLSETS-START -->

| Toolset | Description                          |
|---------|--------------------------------------|
| kiali   | Most common tools for managing Kiali |

<!-- AVAILABLE-TOOLSETS-END -->

### Tools

<!-- AVAILABLE-TOOLSETS-TOOLS-START -->

<details>

<summary>kiali</summary>

- **graph** - Check the status of my mesh by querying Kiali graph
  - `namespace` (`string`) - Optional single namespace to include in the graph (alternative to namespaces)
  - `namespaces` (`string`) - Optional comma-separated list of namespaces to include in the graph

- **mesh_status** - Get the status of mesh components including Istio, Kiali, Grafana, Prometheus and their interactions, versions, and health status

- **istio_config** - Get all Istio configuration objects in the mesh including their full YAML resources and details

- **istio_object_details** - Get detailed information about a specific Istio object including validation and help information
  - `group` (`string`) **(required)** - API group of the Istio object (e.g., 'networking.istio.io', 'gateway.networking.k8s.io')
  - `kind` (`string`) **(required)** - Kind of the Istio object (e.g., 'DestinationRule', 'VirtualService', 'HTTPRoute', 'Gateway')
  - `name` (`string`) **(required)** - Name of the Istio object
  - `namespace` (`string`) **(required)** - Namespace containing the Istio object
  - `version` (`string`) **(required)** - API version of the Istio object (e.g., 'v1', 'v1beta1')

- **istio_object_patch** - Modify an existing Istio object using PATCH method. The JSON patch data will be applied to the existing object.
  - `group` (`string`) **(required)** - API group of the Istio object (e.g., 'networking.istio.io', 'gateway.networking.k8s.io')
  - `json_patch` (`string`) **(required)** - JSON patch data to apply to the object
  - `kind` (`string`) **(required)** - Kind of the Istio object (e.g., 'DestinationRule', 'VirtualService', 'HTTPRoute', 'Gateway')
  - `name` (`string`) **(required)** - Name of the Istio object
  - `namespace` (`string`) **(required)** - Namespace containing the Istio object
  - `version` (`string`) **(required)** - API version of the Istio object (e.g., 'v1', 'v1beta1')

- **istio_object_create** - Create a new Istio object using POST method. The JSON data will be used to create the new object.
  - `group` (`string`) **(required)** - API group of the Istio object (e.g., 'networking.istio.io', 'gateway.networking.k8s.io')
  - `json_data` (`string`) **(required)** - JSON data for the new object
  - `kind` (`string`) **(required)** - Kind of the Istio object (e.g., 'DestinationRule', 'VirtualService', 'HTTPRoute', 'Gateway')
  - `namespace` (`string`) **(required)** - Namespace where the Istio object will be created
  - `version` (`string`) **(required)** - API version of the Istio object (e.g., 'v1', 'v1beta1')

- **istio_object_delete** - Delete an existing Istio object using DELETE method.
  - `group` (`string`) **(required)** - API group of the Istio object (e.g., 'networking.istio.io', 'gateway.networking.k8s.io')
  - `kind` (`string`) **(required)** - Kind of the Istio object (e.g., 'DestinationRule', 'VirtualService', 'HTTPRoute', 'Gateway')
  - `name` (`string`) **(required)** - Name of the Istio object
  - `namespace` (`string`) **(required)** - Namespace containing the Istio object
  - `version` (`string`) **(required)** - API version of the Istio object (e.g., 'v1', 'v1beta1')

- **validations_list** - List all the validations in the current cluster from all namespaces
  - `namespace` (`string`) - Optional single namespace to retrieve validations from (alternative to namespaces)
  - `namespaces` (`string`) - Optional comma-separated list of namespaces to retrieve validations from

- **namespaces** - Get all namespaces in the mesh that the user has access to

- **services_list** - Get all services in the mesh across specified namespaces with health and Istio resource information
  - `namespaces` (`string`) - Comma-separated list of namespaces to get services from (e.g. 'bookinfo' or 'bookinfo,default'). If not provided, will list services from all accessible namespaces

- **service_details** - Get detailed information for a specific service in a namespace, including validation, health status, and configuration
  - `namespace` (`string`) **(required)** - Namespace containing the service
  - `service` (`string`) **(required)** - Name of the service to get details for

- **service_metrics** - Get metrics for a specific service in a namespace. Supports filtering by time range, direction (inbound/outbound), reporter, and other query parameters
  - `byLabels` (`string`) - Comma-separated list of labels to group metrics by (e.g., 'source_workload,destination_service'). Optional
  - `direction` (`string`) - Traffic direction: 'inbound' or 'outbound'. Optional, defaults to 'outbound'
  - `duration` (`string`) - Duration of the query period in seconds (e.g., '1800' for 30 minutes). Optional, defaults to 1800 seconds
  - `namespace` (`string`) **(required)** - Namespace containing the service
  - `quantiles` (`string`) - Comma-separated list of quantiles for histogram metrics (e.g., '0.5,0.95,0.99'). Optional
  - `rateInterval` (`string`) - Rate interval for metrics (e.g., '1m', '5m'). Optional, defaults to '1m'
  - `reporter` (`string`) - Metrics reporter: 'source', 'destination', or 'both'. Optional, defaults to 'source'
  - `requestProtocol` (`string`) - Filter by request protocol (e.g., 'http', 'grpc', 'tcp'). Optional
  - `service` (`string`) **(required)** - Name of the service to get metrics for
  - `step` (`string`) - Step between data points in seconds (e.g., '15'). Optional, defaults to 15 seconds

- **workloads_list** - Get all workloads in the mesh across specified namespaces with health and Istio resource information
  - `namespaces` (`string`) - Comma-separated list of namespaces to get workloads from (e.g. 'bookinfo' or 'bookinfo,default'). If not provided, will list workloads from all accessible namespaces

- **workload_details** - Get detailed information for a specific workload in a namespace, including validation, health status, and configuration
  - `namespace` (`string`) **(required)** - Namespace containing the workload
  - `workload` (`string`) **(required)** - Name of the workload to get details for

- **workload_metrics** - Get metrics for a specific workload in a namespace. Supports filtering by time range, direction (inbound/outbound), reporter, and other query parameters
  - `byLabels` (`string`) - Comma-separated list of labels to group metrics by (e.g., 'source_workload,destination_service'). Optional
  - `direction` (`string`) - Traffic direction: 'inbound' or 'outbound'. Optional, defaults to 'outbound'
  - `duration` (`string`) - Duration of the query period in seconds (e.g., '1800' for 30 minutes). Optional, defaults to 1800 seconds
  - `namespace` (`string`) **(required)** - Namespace containing the workload
  - `quantiles` (`string`) - Comma-separated list of quantiles for histogram metrics (e.g., '0.5,0.95,0.99'). Optional
  - `rateInterval` (`string`) - Rate interval for metrics (e.g., '1m', '5m'). Optional, defaults to '1m'
  - `reporter` (`string`) - Metrics reporter: 'source', 'destination', or 'both'. Optional, defaults to 'source'
  - `requestProtocol` (`string`) - Filter by request protocol (e.g., 'http', 'grpc', 'tcp'). Optional
  - `step` (`string`) - Step between data points in seconds (e.g., '15'). Optional, defaults to 15 seconds
  - `workload` (`string`) **(required)** - Name of the workload to get metrics for

- **health** - Get health status for apps, workloads, and services across specified namespaces in the mesh. Returns health information including error rates and status for the requested resource type
  - `namespaces` (`string`) - Comma-separated list of namespaces to get health from (e.g. 'bookinfo' or 'bookinfo,default'). If not provided, returns health for all accessible namespaces
  - `queryTime` (`string`) - Unix timestamp (in seconds) for the prometheus query. If not provided, uses current time. Optional
  - `rateInterval` (`string`) - Rate interval for fetching error rate (e.g., '10m', '5m', '1h'). Default: '10m'
  - `type` (`string`) - Type of health to retrieve: 'app', 'service', or 'workload'. Default: 'app'

- **workload_logs** - Get logs for a specific workload's pods in a namespace. Only requires namespace and workload name - automatically discovers pods and containers. Optionally filter by container name, time range, and other parameters. Container is auto-detected if not specified.
  - `container` (`string`) - Optional container name to filter logs. If not provided, automatically detects and uses the main application container (excludes istio-proxy and istio-init)
  - `namespace` (`string`) **(required)** - Namespace containing the workload
  - `previous` (`boolean`) - Whether to include logs from previous terminated containers (default: false)
  - `since` (`string`) - Time duration to fetch logs from (e.g., '5m', '1h', '30s'). If not provided, returns recent logs
  - `tail` (`integer`) - Number of lines to retrieve from the end of logs (default: 100)
  - `workload` (`string`) **(required)** - Name of the workload to get logs for

</details>


<!-- AVAILABLE-TOOLSETS-TOOLS-END -->
## 🎥 Demos <a id="demos"></a>

In this video, we explore how the Mesh Control Plane (MCP) in Kubernetes/OpenShift works together with Kiali to validate Istio configuration objects directly in your editor (_Cursor_).

<a href="https://youtu.be/1l9m1B5uEPw" target="_blank">
 <img src="docs/images/kiali_mcp_cursor.png" alt="Cursor: Kiali-mcp-server running" width="240"  />
</a>


## 🧑‍💻 Development <a id="development"></a>

### Running with mcp-inspector

Compile the project and run the Kiali MCP server with [mcp-inspector](https://modelcontextprotocol.io/docs/tools/inspector) to inspect the MCP server.

```shell
# Compile the project
make build
# Run the Kubernetes MCP server with mcp-inspector
npx @modelcontextprotocol/inspector@latest $(pwd)/kiali-mcp-server --kiali-server-url "https://kiali-istio-system.apps-crc.testing/" --kiali-insecure
```
