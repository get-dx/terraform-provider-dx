# DX Terraform Provider

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23

## Using the provider

Add the following block into your Terraform files:

```terraform
terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/get-dx/dx"
      version = "~> 0.6.0"
    }
  }
}
```

Configure the provider:

```terraform
provider "dx" {
  # Define your Web API token here, or set `DX_WEB_API_TOKEN` in your environment.
  #
  # To manage scorecards, the token must have the following scopes:
  #
  # - scorecards:read
  # - scorecards:write
  #
  api_token = "<your api token>"
}
```

### Resource Types

This provider comes with the following resource types:

- [`dx_scorecard`](docs/resources/scorecard.md)

See the [examples/](examples/) directory for example usage.

### Notes about compatibility with the DX Web API

- The DX API will convert some optional string attributes from `null` to `""`. Therefore, to avoid inconsistent plan results in Terraform state, we treat `null` and `""` as semantically equal in the provider for these attributes. If you do not want to set a value for one of these attributes, please use `null` or omit the attribute. This applies to the following:
  - `dx_scorecard.description`
  - `dx_scorecard.checks.output_type`
  - `dx_scorecard.checks.output_aggregation`

## Developing the Provider

See [CONTRIBUTING.md](./CONTRIBUTING.md).
