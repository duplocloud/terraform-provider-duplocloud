# Architecture Documentation

## Overview

The `terraform-provider-duplocloud` is a custom Terraform provider that enables Infrastructure as Code (IaC) management for the DuploCloud platform. This provider follows the standard Terraform plugin architecture and uses the Terraform Plugin SDK v2.

## Table of Contents

- [High-Level Architecture](#high-level-architecture)
- [Project Structure](#project-structure)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Resource Lifecycle](#resource-lifecycle)
- [SDK Layer](#sdk-layer)
- [Design Patterns](#design-patterns)
- [Multi-Cloud Support](#multi-cloud-support)
- [Testing Strategy](#testing-strategy)
- [Extension Points](#extension-points)

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Terraform CLI                            │
│                  (User Interface)                            │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            │ gRPC/Plugin Protocol
                            │
┌───────────────────────────▼─────────────────────────────────┐
│              Terraform Provider (main.go)                    │
│                   Plugin Entry Point                         │
└───────────────────────────┬─────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                  Provider Core (provider.go)                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Schema Definition & Configuration                   │   │
│  │  - Authentication (duplo_host, duplo_token)         │   │
│  │  - SSL Configuration                                 │   │
│  │  - HTTP Timeout Settings                            │   │
│  └─────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────┘
                            │
        ┌───────────────────┴───────────────────┐
        │                                       │
┌───────▼──────────┐                   ┌───────▼──────────┐
│   Resources      │                   │  Data Sources    │
│   (duplocloud/)  │                   │  (duplocloud/)   │
│                  │                   │                  │
│  - AWS Resources │                   │  - Tenant Data   │
│  - Azure Res.    │                   │  - Plan Data     │
│  - GCP Resources │                   │  - Service Data  │
│  - K8s Resources │                   │  - Infra Data    │
└───────┬──────────┘                   └───────┬──────────┘
        │                                       │
        └───────────────────┬───────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                    DuploCloud SDK (duplosdk/)                │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Client Layer (client.go)                           │   │
│  │  - HTTP Client Management                           │   │
│  │  - Authentication & Headers                         │   │
│  │  - Retry Logic & Rate Limiting                      │   │
│  └─────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  API Abstraction Layer                              │   │
│  │  - Tenant APIs (tenant.go)                          │   │
│  │  - Infrastructure APIs (infrastructure.go)          │   │
│  │  - Service APIs (replication_controller.go)         │   │
│  │  - Cloud-specific APIs (aws_*, azure_*, gcp_*)      │   │
│  └─────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            │ REST API (HTTP/HTTPS)
                            │
┌───────────────────────────▼─────────────────────────────────┐
│                   DuploCloud Platform                        │
│                    (REST API Server)                         │
└─────────────────────────────────────────────────────────────┘
```

## Project Structure

```
terraform-provider-duplocloud/
├── main.go                      # Provider entry point
├── go.mod                       # Go module dependencies
├── go.sum                       # Dependency checksums
├── Makefile                     # Build and development tasks
│
├── duplocloud/                  # Provider implementation
│   ├── provider.go              # Provider schema and configuration
│   ├── provider_test.go         # Provider tests
│   │
│   ├── resource_duplo_*.go      # Resource implementations
│   ├── data_source_duplo_*.go   # Data source implementations
│   │
│   ├── schema_k8s_*.go          # Kubernetes schema definitions
│   ├── structure_*.go           # Data structure converters
│   ├── utils.go                 # Utility functions
│   ├── validators.go            # Input validators
│   ├── diff_utils.go            # Diff suppression functions
│   └── *_test.go                # Unit tests
│
├── duplosdk/                    # DuploCloud SDK
│   ├── client.go                # HTTP client and base methods
│   ├── client_test.go           # Client tests
│   ├── types.go                 # Common type definitions
│   │
│   ├── tenant.go                # Tenant management APIs
│   ├── infrastructure.go        # Infrastructure APIs
│   ├── plan.go                  # Plan management APIs
│   ├── native_hosts.go          # Host management APIs
│   │
│   ├── aws_*.go                 # AWS-specific APIs
│   ├── azure_*.go               # Azure-specific APIs
│   ├── gcp_*.go                 # GCP-specific APIs
│   ├── k8*.go                   # Kubernetes APIs
│   ├── ecs_*.go                 # ECS APIs
│   └── utils.go                 # SDK utilities
│
├── docs/                        # Auto-generated documentation
│   ├── resources/               # Resource documentation
│   └── data-sources/            # Data source documentation
│
├── examples/                    # Example configurations
│   ├── service/                 # Service examples
│   ├── infrastructure/          # Infrastructure examples
│   └── ...                      # Other examples
│
├── templates/                   # Documentation templates
│
├── scripts/                     # Helper scripts
│
├── .github/                     # GitHub workflows
│   └── workflows/               # CI/CD pipelines
│
└── internal/                    # Internal packages
```

## Core Components

### 1. Provider Entry Point (`main.go`)

The entry point initializes the Terraform plugin and registers the provider:

```go
func main() {
    opts := &plugin.ServeOpts{
        Debug:        debug,
        ProviderAddr: "registry.terraform.io/duplocloud/duplocloud",
        ProviderFunc: func() *schema.Provider {
            return duplocloud.Provider()
        },
    }
    plugin.Serve(opts)
}
```

**Key Responsibilities:**
- Initialize the plugin server
- Register the provider function
- Enable debug mode support for development

### 2. Provider Core (`duplocloud/provider.go`)

The provider core defines the provider schema, resources, data sources, and configuration:

**Configuration Schema:**
- `duplo_host`: DuploCloud API endpoint
- `duplo_token`: Authentication token
- `ssl_no_verify`: SSL verification toggle
- `http_timeout`: HTTP request timeout

**Key Functions:**
- `Provider()`: Returns the provider schema with all resources and data sources
- `providerConfigure()`: Initializes the DuploCloud SDK client with user configuration

### 3. Resources (`duplocloud/resource_duplo_*.go`)

Resources represent infrastructure components that can be created, read, updated, and deleted (CRUD operations).

**Resource Structure:**
```go
func resourceDuploService() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceDuploServiceCreate,
        ReadContext:   resourceDuploServiceRead,
        UpdateContext: resourceDuploServiceUpdate,
        DeleteContext: resourceDuploServiceDelete,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: map[string]*schema.Schema{
            // Schema definition
        },
    }
}
```

**Categories:**
- **AWS Resources**: EC2 hosts, Lambda functions, RDS, DynamoDB, ECS, etc.
- **Azure Resources**: Virtual machines, databases, storage accounts, etc.
- **GCP Resources**: Compute instances, Cloud Functions, storage buckets, etc.
- **Kubernetes Resources**: ConfigMaps, Secrets, Ingress, Jobs, etc.
- **DuploCloud Resources**: Tenants, services, infrastructure, plans, etc.

### 4. Data Sources (`duplocloud/data_source_duplo_*.go`)

Data sources provide read-only access to existing infrastructure:

```go
func dataSourceTenant() *schema.Resource {
    return &schema.Resource{
        ReadContext: dataSourceTenantRead,
        Schema: map[string]*schema.Schema{
            // Schema definition
        },
    }
}
```

### 5. DuploCloud SDK (`duplosdk/`)

The SDK layer abstracts the DuploCloud REST API and provides Go-native interfaces.

**Client Structure:**
```go
type Client struct {
    HTTPClient *http.Client
    HostURL    string
    Token      string
}
```

**Key Features:**
- HTTP client management with configurable timeouts
- Authentication via bearer tokens
- Retry logic for rate-limited requests
- Type-safe API methods
- Error handling and logging

## Data Flow

### Create Resource Flow

```
1. User runs `terraform apply`
   ↓
2. Terraform calls CreateContext function
   ↓
3. Extract configuration from ResourceData
   ↓
4. Call SDK method (e.g., client.TenantCreate())
   ↓
5. SDK constructs HTTP request
   ↓
6. Send POST/PUT request to DuploCloud API
   ↓
7. Parse API response
   ↓
8. Update Terraform state with resource ID and attributes
   ↓
9. Return success or error diagnostics
```

### Read Resource Flow

```
1. Terraform calls ReadContext function
   ↓
2. Extract resource ID from state
   ↓
3. Call SDK method (e.g., client.TenantGet())
   ↓
4. SDK sends GET request to DuploCloud API
   ↓
5. Parse API response
   ↓
6. Update Terraform state with current attributes
   ↓
7. Detect drift if attributes changed
   ↓
8. Return success or error diagnostics
```

### Update Resource Flow

```
1. Terraform detects configuration changes
   ↓
2. Calls UpdateContext function
   ↓
3. Extract old and new configuration
   ↓
4. Determine which attributes changed
   ↓
5. Call SDK update method
   ↓
6. SDK sends PUT/PATCH request
   ↓
7. Update Terraform state
   ↓
8. Return success or error diagnostics
```

### Delete Resource Flow

```
1. User runs `terraform destroy`
   ↓
2. Terraform calls DeleteContext function
   ↓
3. Extract resource ID from state
   ↓
4. Call SDK delete method
   ↓
5. SDK sends DELETE request
   ↓
6. Wait for resource deletion (if async)
   ↓
7. Remove resource from Terraform state
   ↓
8. Return success or error diagnostics
```

## Resource Lifecycle

### State Management

Terraform maintains state for each resource, which includes:
- **Resource ID**: Unique identifier in DuploCloud
- **Attributes**: All configured and computed values
- **Metadata**: Timeouts, dependencies, etc.

### Import Support

Most resources support `terraform import` to bring existing infrastructure under Terraform management:

```bash
terraform import duplocloud_tenant.myapp tenant-id
```

### Timeouts

Resources can define custom timeouts for operations:
```go
Timeouts: &schema.ResourceTimeout{
    Create: schema.DefaultTimeout(15 * time.Minute),
    Update: schema.DefaultTimeout(15 * time.Minute),
    Delete: schema.DefaultTimeout(15 * time.Minute),
}
```

## SDK Layer

### Client Initialization

```go
client, err := duplosdk.NewClient(host, token)
```

### API Method Pattern

```go
func (c *Client) TenantCreate(rq *DuploTenant) (*DuploTenant, ClientError) {
    // 1. Construct API endpoint
    // 2. Marshal request body
    // 3. Send HTTP request
    // 4. Handle response
    // 5. Unmarshal response body
    // 6. Return result or error
}
```

### Error Handling

The SDK uses custom error types for better error handling:
```go
type ClientError interface {
    error
    Status() int
}
```

### Retry Logic

The SDK implements retry logic for rate-limited requests:
- Exponential backoff
- Configurable retry attempts
- Rate limit detection

## Design Patterns

### 1. Schema-Driven Development

All resources and data sources are defined using declarative schemas:
```go
Schema: map[string]*schema.Schema{
    "name": {
        Type:     schema.TypeString,
        Required: true,
    },
}
```

### 2. Separation of Concerns

- **Provider Layer**: Terraform-specific logic, schema definitions
- **SDK Layer**: API communication, data marshaling
- **Utilities**: Shared helper functions

### 3. Diff Suppression

Custom diff suppression functions handle API inconsistencies:
```go
DiffSuppressFunc: diffSuppressWhenNotChange,
```

### 4. Validation

Input validation at multiple levels:
- Schema-level validation (Required, Optional, Computed)
- Custom validators for complex rules
- API-level validation

### 5. Structure Converters

Separate functions convert between Terraform schemas and SDK types:
```go
func expandK8sPodSpec(in []interface{}) *v1.PodSpec
func flattenK8sPodSpec(spec *v1.PodSpec) []interface{}
```

## Multi-Cloud Support

The provider supports multiple cloud platforms through DuploCloud:

### AWS
- EC2, ECS, Lambda, RDS, DynamoDB
- Load Balancers, Target Groups
- CloudWatch, SNS, SQS
- Elasticsearch, EMR

### Azure
- Virtual Machines, Scale Sets
- SQL Database, PostgreSQL, MySQL
- Storage Accounts, Cosmos DB
- Key Vault, Redis Cache

### GCP
- Compute Instances, Node Pools
- Cloud Functions, Cloud SQL
- Storage Buckets, Firestore
- Pub/Sub, Redis

### Kubernetes
- ConfigMaps, Secrets
- Ingress, Services
- Jobs, CronJobs
- Persistent Volume Claims

## Testing Strategy

### Unit Tests
- Test individual functions and utilities
- Mock SDK responses
- Validate schema definitions

### Acceptance Tests
- Test full resource lifecycle (CRUD)
- Require `TF_ACC=1` environment variable
- Use real or mock DuploCloud API

### Integration Tests
- Test resource interactions
- Validate dependencies
- Test error scenarios

### Test Execution
```bash
# Unit tests
make test

# Acceptance tests
TF_ACC=1 go test ./duplocloud/... -v -timeout 120m
```

## Extension Points

### Adding a New Resource

1. Create `resource_duplo_<name>.go` in `duplocloud/`
2. Implement CRUD functions
3. Define schema
4. Add SDK methods in `duplosdk/`
5. Register in `provider.go`
6. Add tests
7. Generate documentation

### Adding a New Data Source

1. Create `data_source_duplo_<name>.go`
2. Implement Read function
3. Define schema
4. Add SDK methods if needed
5. Register in `provider.go`
6. Add tests
7. Generate documentation

### Adding SDK Methods

1. Add method to appropriate SDK file
2. Define request/response types
3. Implement HTTP communication
4. Add error handling
5. Add unit tests

## Performance Considerations

### Caching
- Provider configuration is cached per Terraform run
- SDK client is reused across resources

### Parallel Operations
- Terraform handles parallelism automatically
- Provider must be thread-safe
- SDK client uses concurrent-safe HTTP client

### Rate Limiting
- SDK implements retry logic for rate limits
- Exponential backoff prevents API overload

## Security

### Authentication
- Token-based authentication
- Tokens stored securely in Terraform state
- Support for environment variables

### SSL/TLS
- HTTPS by default
- Optional SSL verification bypass for development

### Sensitive Data
- Tokens marked as sensitive in schema
- Secrets handled securely
- No logging of sensitive values

## Documentation Generation

Documentation is auto-generated using `terraform-plugin-docs`:

```bash
go generate
```

This generates:
- Resource documentation in `docs/resources/`
- Data source documentation in `docs/data-sources/`
- Index and guides

## Future Enhancements

- Enhanced error messages with troubleshooting guides
- Provider-level caching for improved performance
- Webhook support for async operations
- Enhanced testing framework
- Metrics and telemetry
- Provider-level validation

## References

- [Terraform Plugin SDK v2](https://github.com/hashicorp/terraform-plugin-sdk)
- [Terraform Plugin Development](https://developer.hashicorp.com/terraform/plugin)
- [DuploCloud Documentation](https://duplocloud.com/docs)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on contributing to this provider.

## Support

For issues and questions:
- ClickUp: Report bugs and feature requests
- DuploCloud Support: Platform-specific questions (support@duplocloud.com)
- GitHub Discussions: Community questions and best practices
