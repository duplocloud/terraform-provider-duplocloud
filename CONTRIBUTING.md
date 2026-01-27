# Contributing to terraform-provider-duplocloud

Thank you for your interest in contributing to the DuploCloud Terraform Provider! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style Guidelines](#code-style-guidelines)
- [Documentation](#documentation)
- [Reporting Issues](#reporting-issues)
- [Community](#community)

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to support@duplocloud.com.

## Getting Started

### Prerequisites

- **Go**: Version 1.23.0 or higher
- **Terraform**: Version 1.0 or higher
- **Git**: For version control
- **Make**: For running build tasks
- **DuploCloud Account**: For testing (optional but recommended)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/terraform-provider-duplocloud.git
   cd terraform-provider-duplocloud
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/duplocloud/terraform-provider-duplocloud.git
   ```

## Development Setup

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Verify dependencies
go mod verify
```

### Build the Provider

```bash
# Build the provider binary
make build

# Install the provider locally
make install
```

### Verify Installation

```bash
# Check if the provider is installed
ls -la ~/.terraform.d/plugins/registry.terraform.io/duplocloud/duplocloud/
```

## Making Changes

### Create a Branch

Always create a new branch for your changes. **Include the ClickUp issue ID** in your branch name:

```bash
git checkout -b feature/DUPLO-123-your-feature-name
# or
git checkout -b fix/DUPLO-456-issue-description
```

**Branch Naming Convention:**
- `feature/DUPLO-XXX-description` - New features
- `fix/DUPLO-XXX-description` - Bug fixes
- `docs/DUPLO-XXX-description` - Documentation changes
- `refactor/DUPLO-XXX-description` - Code refactoring
- `test/DUPLO-XXX-description` - Test additions or modifications

**Example:** `feature/DUPLO-789-add-lambda-event-config`

### Development Workflow

1. **Make your changes** in the appropriate files
2. **Write or update tests** for your changes
3. **Update documentation** if needed
4. **Run tests** to ensure nothing breaks
5. **Commit your changes** with clear commit messages

### Commit Message Guidelines

Follow conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Example:**
```
feat(aws): add support for Lambda function event config

- Add new resource duplocloud_aws_lambda_function_event_config
- Implement CRUD operations
- Add acceptance tests
- Update documentation

Closes #123
```

## Testing

### Run Unit Tests

```bash
# Run all unit tests
make test

# Run tests for a specific package
go test ./duplocloud/... -v

# Run specific tests by name
go test ./duplocloud -run TestDuploService -v
```

### Run Acceptance Tests

Acceptance tests require a real DuploCloud environment:

```bash
# Set environment variables
export duplo_host="https://your-duplocloud-instance.com"
export duplo_token="your-api-token"

# Run acceptance tests
make testacc

# Run specific acceptance test
TF_ACC=1 go test ./duplocloud/... -v -run=TestAccDuploService_basic -timeout 120m
```

**Note:** Acceptance tests create real resources and may incur costs.

### Test Coverage

Aim for at least 70% test coverage for new code:

```bash
# Generate coverage report
go test ./duplocloud/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Submitting Changes

### Before Submitting

1. **Ensure all tests pass:**
   ```bash
   make test
   ```

2. **Format your code:**
   ```bash
   go fmt ./...
   ```

3. **Run linters:**
   ```bash
   golangci-lint run
   ```

4. **Update documentation:**
   ```bash
   make doc
   ```

5. **Commit and push your changes:**
   ```bash
   git add .
   git commit -m "feat: your feature description"
   git push origin feature/your-feature-name
   ```

### Create a Pull Request

1. Go to the [repository on GitHub](https://github.com/duplocloud/terraform-provider-duplocloud)
2. Click "New Pull Request"
3. Select your branch
4. Fill out the PR template with:
   - **ClickUp Ticket**: Link to the related ClickUp ticket (DUPLO-XXXXX) - **Required**
   - **Overview**: Brief description of changes
   - **Summary of changes**: Detailed list of modifications
   - **Testing performed**: How you tested your changes
   - **Breaking changes**: Any breaking changes (if applicable)

**Note:** All PRs must reference a ClickUp ticket. The PR validation workflow checks for the ClickUp ticket ID (DUPLO-XXXXX) in the PR title or description.

### PR Review Process

- A maintainer will review your PR
- Address any feedback or requested changes
- Once approved, your PR will be merged
- Your contribution will be included in the next release

## Code Style Guidelines

### Go Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions small and focused

### Resource Implementation

When adding a new resource:

```go
func resourceDuploExample() *schema.Resource {
    return &schema.Resource{
        Description: "Manages an example resource in DuploCloud",
        
        CreateContext: resourceDuploExampleCreate,
        ReadContext:   resourceDuploExampleRead,
        UpdateContext: resourceDuploExampleUpdate,
        DeleteContext: resourceDuploExampleDelete,
        
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        
        Timeouts: &schema.ResourceTimeout{
            Create: schema.DefaultTimeout(15 * time.Minute),
            Update: schema.DefaultTimeout(15 * time.Minute),
            Delete: schema.DefaultTimeout(15 * time.Minute),
        },
        
        Schema: map[string]*schema.Schema{
            "tenant_id": {
                Description: "The GUID of the tenant",
                Type:        schema.TypeString,
                Required:    true,
                ForceNew:    true,
            },
            // ... more schema fields
        },
    }
}
```

### Error Handling

- Always return meaningful error messages
- Include context in error messages
- Use `diag.Diagnostics` for errors in Terraform resources

```go
if err != nil {
    return diag.Errorf("failed to create resource: %s", err)
}
```

### SDK Methods

When adding SDK methods in `duplosdk/`:

```go
// ExampleGet retrieves an example resource
func (c *Client) ExampleGet(tenantID, resourceID string) (*DuploExample, ClientError) {
    rp := fmt.Sprintf("subscriptions/%s/ExampleGet/%s", tenantID, resourceID)
    
    var resp DuploExample
    err := c.getAPI(rp, &resp)
    if err != nil {
        return nil, err
    }
    
    return &resp, nil
}
```

## Documentation

### Auto-Generated Documentation

Documentation is auto-generated from code comments:

```bash
# Generate documentation
go generate
# or
make doc
```

### Resource Documentation

Add descriptions to schema fields:

```go
"name": {
    Description: "The name of the resource. Must be unique within the tenant.",
    Type:        schema.TypeString,
    Required:    true,
},
```

### Examples

Add examples in the `examples/` directory:

```
examples/
└── resources/
    └── duplocloud_example/
        └── resource.tf
```

## Reporting Issues

**Note:** We use **ClickUp** for issue tracking. Please create issues in ClickUp rather than GitHub Issues.

### Bug Reports

When reporting bugs in ClickUp, include:

- **Description**: Clear description of the issue
- **Steps to Reproduce**: Detailed steps to reproduce
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**: Provider version, Terraform version, OS
- **Configuration**: Relevant Terraform configuration (sanitized)
- **Logs**: Relevant logs with `TF_LOG=DEBUG`

### Feature Requests

When requesting features in ClickUp, include:

- **Use Case**: Why you need this feature
- **Proposed Solution**: How you envision it working
- **Alternatives**: Other solutions you've considered
- **Additional Context**: Any other relevant information

## Community

### Getting Help

- **ClickUp**: For bugs, feature requests, and task tracking
- **GitHub Discussions**: For questions and community discussions
- **Email**: support@duplocloud.com for general support

### Stay Updated

- Watch the repository for updates
- Check the [CHANGELOG](CHANGELOG.md) for release notes
- Review the [Architecture Documentation](ARCHITECTURE.md) for technical details

## Development Tips

### Debugging

Use VS Code debug configuration:

1. Set breakpoints in your code
2. Start debug session (F5)
3. Copy `TF_REATTACH_PROVIDERS` from debug console
4. Run Terraform with the environment variable

### Local Testing

Test against a local DuploCloud instance:

```bash
export duplo_host="http://localhost:60020"
export duplo_token="FAKE"
terraform init && terraform apply
```

### Common Issues

**Issue: Provider not found**
```bash
# Reinstall the provider
make install
terraform init -upgrade
```

**Issue: Import errors**
```bash
# Clean and rebuild
go clean -cache
go mod tidy
make build
```

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).

## Questions?

If you have questions about contributing, feel free to:
- Open a GitHub Discussion
- Email us at support@duplocloud.com
- Check existing ClickUp tickets and PRs for similar questions

Thank you for contributing to terraform-provider-duplocloud! 🎉
