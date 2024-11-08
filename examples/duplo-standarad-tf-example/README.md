## Infrastructure Automation with Terraform

Welcome to the infrastructure automation repository using Terraform. This repository houses all the code needed to automate infrastructure provisioning for any Application using DuploCloud platform. For more information on Terraform, please refer to the [official documentation](https://developer.hashicorp.com/terraform/docs).

### Directory Structure

```
terraform-repository/
│
├── scripts/
│   ├── apply.sh
│   └── plan.sh
│
├── config/
│   ├── prod/
│   │   ├── admin.tfvars
│   │   |── aws-services.tfvars
│   │   |── app.tfvars
│   ├── customer1/
│   │   ├── admin.tfvars
│   │   |── aws-service.tfvars
│   │   |── app.tfvars
│   └── test/
│   │   ├── admin.tfvars
│   │   |── aws-service.tfvars
│   │   |── app.tfvars
└── terraform/
    ├── admin/
    │   ├── main.tf
    │   └── vars.tf
    ├── aws-servces/
    │   ├── main.tf
    │   ├── rds.tf
    │   └── vars.tf
    └── app/
        ├── main.tf
        ├── service1.tf
        └── variables.tf
```
It root level there are three directories as below

1. **scripts:** This directory contains the wrapper scripts to run terraform `plan`, `apply` and `destroy` commands. These wrapper scripts take care of setting up the workspace for environment, setting AWS credentials using DuploCloud JIT, s3 as backend to save the state file and more.

2. **configs:** Configs directory allows user to override the default values for terraform modules for different environments. One has to follow the naming convention as `config/<env>/<terraform_module>.tfvars` to override the default values for the terraform project.

3. **terraform:** This directory has subdirectories for different terraform modules. To keep the modularity, Terraform code is divided into three different terraform modules as below. 

   - **admin**: This module take care of all the administrative tasks needs to be done in DuploCloud like creating infra, tenants, security rules and outputs the common variables used by subsequent terraform projects. 
   - **aws-services**: As name suggest this layer is used for deploying AWS or Cloud specific services and required dependencies like RDS, Elasticache cluster, s3 buckets etc.
   - **app**: This module is used for deploying application specific servics

### Running Terraform on your local machine

#### Prerequisites

Before getting started, ensure you have the following prerequisites:

- Install JQ:
  - Ubuntu: `apt-get install jq -y`
  - RHEL: `yum install jq -y`
  - macOS: `brew install jq`
- Install Terraform. Follow the steps outlined in the [official Terraform download guide](https://developer.hashicorp.com/terraform/downloads).

#### Terraform Plan

Before deploying any infrastructure changes, it's crucial to run a plan and review the output. The `terraform plan` command provides insights into the proposed changes.

##### Planning for All Projects

To plan changes for all Terraform modules, run the following command:

```bash
./scripts/plan.sh <env_name>
```

For example, to plan all projects for the test environment:

```bash
./scripts/plan.sh test
```

##### Planning for Specific Project

To plan changes for a specific terraform, use the following command:

```bash
./scripts/plan.sh <env_name> <terraform_module>
```

For example, to plan changes only for the app module in the sandbox environment:

```bash
./scripts/plan.sh test app
```

#### Terraform Apply

Once you've verified the Terraform plan, you can apply the changes to deploy the infrastructure.

##### Applying for all terraform projects

To apply changes for all Terraform projects, execute:

```bash
./scripts/apply.sh <env_name>
```

For instance, to apply all projects for the `test` environment:

```bash
./scripts/apply.sh test
```

##### Applying for Specific Project

To apply changes for a specific module, run:

```bash
./scripts/apply.sh <env_name> <module_name>
```

For example, to apply changes only for the `mcb` project in the `test` environment:

```bash
./scripts/apply.sh test app
```

### Notes

- Always review the Terraform plan before applying changes to understand the proposed modifications.
- Ensure all prerequisites are met before running Terraform commands.
