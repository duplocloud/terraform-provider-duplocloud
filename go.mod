module terraform-provider-duplocloud

go 1.16

require (
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/hashicorp/go-getter v1.6.1 // indirect
	github.com/hashicorp/terraform-plugin-docs v0.4.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.5.0
	github.com/ucarion/jcs v0.1.2
	golang.org/x/mod v0.4.0 // indirect
	golang.org/x/tools v0.1.1-0.20210201201750-4d4ee958a9b7 // indirect
)

replace github.com/hashicorp/terraform-plugin-docs v0.4.0 => github.com/nmuesch/terraform-plugin-docs v0.4.1-0.20210304202717-40b0963b9557

replace github.com/hashicorp/go-getter v1.4.0 => github.com/hashicorp/go-getter v1.6.1

replace github.com/hashicorp/go-getter v1.5.0 => github.com/hashicorp/go-getter v1.6.1
