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
      "registry.terraform.io/local/dx" = "/path/to/your/go/bin"
    }
    direct {}
  }
  ```

  Note that this references `local/dx` instead of `get-dx/dx`.

- In your TF module, use the `provider` as:

  ```terraform
  terraform {
    required_providers {
      dx = {
        source = "registry.terraform.io/local/dx"
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
make generate
```

### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to the provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Publishing (for maintainers)

- Update [CHANGELOG.md](./CHANGELOG.md) with release notes.

- If applicable, update the version range in `README.md` and the examples:

  ```diff
   terraform {
     required_providers {
       dx = {
         source  = "registry.terraform.io/get-dx/dx"
  -      version = "~> 0.3.0"
  +      version = "~> 0.4.0"
       }
     }
   }
  ```

  (Make sure to check in any changes to generated docs afterward.)

- Create a tag and push it:

  ```shell
  git tag v0.1.0         # Replace with the new version
  git push --tags
  ```

- The [Release action](https://github.com/get-dx/terraform-provider-dx/actions/workflows/release.yml) will run based off of the tag trigger. The action will build the provider for several architectures, then create a [release](https://github.com/get-dx/terraform-provider-dx/releases) and attach the assets to it.

- Once the release has been published, within a minute or two, the release will automatically propagate to the Terraform registry and will be listed as the `latest` version on the [provider overview page](https://registry.terraform.io/providers/get-dx/dx/latest).
