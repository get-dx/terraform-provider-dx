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
      source  = "registry.terraform.io/local/dx" # TODO: replace after publishing
      version = "0.1.0"
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

## Developing the Provider

See [CONTRIBUTING.md](./CONTRIBUTING.md).
