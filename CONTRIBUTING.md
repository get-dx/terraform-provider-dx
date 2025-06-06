# Contributing

Thank you for being interested in contributing to the Terraform provider!

## Developing locally

### Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23

### First-time setup

Follow these steps to run the locally-built provider rather than the one from the TF registry:

- Clone the repository.

- Enter the repository directory.

- Build the provider using the Go `install` command:

  ```shell
  go install
  ```

  This will put the binary in the `$GOPATH/bin` directory. Make sure the directory is in your system `$PATH`.

- Create a `~/.terraformrc` file that lets you override the default provider location:

  ```terraform
  provider_installation {
    dev_overrides {
      "registry.terraform.io/get-dx/dx" = "/path/to/your/go/bin"
    }
    direct {}
  }
  ```

- In your TF module, use the `provider` as:

  ```terraform
  terraform {
    required_providers {
      dx = {
        source = "registry.terraform.io/get-dx/dx"
      }
    }
  }
  ```

- Optionally change your log level:

  ```shell
  export TF_LOG=DEBUG
  ```

- Optionally change the DX Web API URL:

  ```shell
  export DX_WEB_API_URL="https://api.getdx.com"
  ```

### Iteration loop

Build the provider and put it in the go path:

```shell
go install
```

Then in your TF module, with the `source` override from before, run Terraform:

```shell
terraform plan
terraform apply
```

### Generating documentation

Run the following:

```shell
go generate
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to the provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.
