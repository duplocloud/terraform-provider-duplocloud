# Terraform Provider duplocloud

## Building locally.

To build and install, run `make install`.

To generate or update documentation, run `go generate`.

## Running examples.

```
cd examples/service
terraform init && terraform apply

```

## Using debug output during execution.

``` 
export TF_LOG_PATH=duplo.log
export TF_LOG=TRACE
terraform init && terraform apply
```

## useful commands

```
######## list objects in state
terraform  state list
>>>>> duplocloud_tenant.tfc15
>>>>> duplocloud_tenant.tfc16

######## remove (could be deleted tenants from duplocloud) form state
terraform  state rm duplocloud_tenant.tfc15
>>>>> Removed duplocloud_tenant.tfc15
>>>>> Successfully removed 1 resource instance(s).

######## import tenant 
terraform import 'duplocloud_tenant.tfc16' 3a0c5cee-287f-4c6f-9c62-02e3c9d8f8ee
>>>>> duplocloud_tenant.tfc16: Import prepared!
>>>>> Prepared duplocloud_tenant for import
>>>>> duplocloud_tenant.tfc16: Refreshing state... [id=3a0c5cee-287f-4c6f-9c62-02e3c9d8f8ee]

######## remove and import again tenant ... (duplocloud exisiting tenant)
terraform  state rm duplocloud_tenant.tfc16
terraform import 'duplocloud_tenant.tfc16' v2/admin/TenantV2/3a0c5cee-287f-4c6f-9c62-02e3c9d8f8ee

##################crudemo
terraform import 'duplocloud_tenant.crud_demo' v2/admin/TenantV2/a5fe441c-beae-4c08-81fa-3e5d21b93481
terraform import 'duplocloud_aws_host.host2' v2/subscriptions/a5fe441c-beae-4c08-81fa-3e5d21b93481/NativeHostV2/i-01f1277e3cba865f8
######

##################compliance
terraform import 'duplocloud_tenant.compliance' v2/admin/TenantV2/a677df6e-4b89-44cb-8cd7-72a0d2ddb47d
terraform import 'duplocloud_aws_host.proxy' v2/subscriptions/a677df6e-4b89-44cb-8cd7-72a0d2ddb47d/NativeHostV2/i-0347b334c62864e32
######

## test 
terraform import 'duplocloud_tenant.tfc21'  v2/admin/TenantV2/88d3d58e-52cd-4320-95b0-ae32ea80fe70
terraform import 'duplocloud_aws_host.tfc22proxy1' v2/subscriptions/51099af9-8ce6-4d11-9522-d7f02df4148e/NativeHostV2/i-02686c3aeed775287
terraform import 'duplocloud_aws_host.tfc22proxy2' v2/subscriptions/51099af9-8ce6-4d11-9522-d7f02df4148e/NativeHostV2/i-0c657d4d9b908162b
terraform import 'duplocloud_aws_host.tfc22proxy3' v2/subscriptions/51099af9-8ce6-4d11-9522-d7f02df4148e/NativeHostV2/i-086275d1d7bb9f745
 
 
```