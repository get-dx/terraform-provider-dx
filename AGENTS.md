# AGENTS.md — terraform-provider-dx

Terraform provider for DX, backed by the DX Web API.

## Commands (see `GNUmakefile`)

| Task                                     | Command                                                     |
| ---------------------------------------- | ----------------------------------------------------------- |
| Build                                    | `make build`                                                |
| Format                                   | `make fmt`                                                  |
| Lint                                     | `make lint` (uses golangci-lint)                            |
| Unit tests                               | `make test`                                                 |
| Acceptance tests                         | `make testacc` (needs a live DX account; not run by agents) |
| Regenerate docs                          | `make generate`                                             |
| Full default (fmt+lint+install+generate) | `make`                                                      |

Run `make fmt`, `make lint`, `make test`, and `make generate` before committing.

## Critical details

- **CI fails on a dirty `make generate` diff.** If you change any resource/data-source schema, run `make generate` and commit the resulting changes in `docs/`.
- Add changelog entries to the `## Unreleased` section of `CHANGELOG.md`.
- Examples under `examples/resources/*/resource.tf` are embedded into the generated docs — only add to them for commonly-used features.
