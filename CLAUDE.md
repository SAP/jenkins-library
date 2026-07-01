# jenkins-library — Claude / AI Agent Context

## What this project is

jenkins-library is the open-source Project Piper CI/CD library (Apache 2.0). It provides generic pipeline steps and a Jenkins Groovy shared library that work with any enterprise — no SAP-proprietary logic. The compiled binary is called `piper`.

**Status: archived December 31 2026. Limited maintenance only (critical fixes). No new features accepted.**

SAP-specific extensions live in the InnerSource `piper-library`, which imports this project as a Go dependency (`github.com/SAP/jenkins-library`).

## Relationship to sibling projects

| Project | Relationship |
|---|---|
| `piper-library` | Consumer/extender — imports jenkins-library as base; adds SAP-specific steps on top. |
| `piper-pipeline-github` | Consumer — GitHub Actions GPP downloads and runs the `piper` binary produced here. |
| `piper-pipeline-azure` / `piper-pipeline-jenkins` | Consumers — legacy GPP templates for Azure Pipelines and Jenkins. |
| `engine` | Future replacement — binary-based step runner; jenkins-library is the catalog-driven baseline. |

## Tech stack

- **Go 1.25** — step implementations in `cmd/`, shared logic in `pkg/`
- **Groovy** — Jenkins shared library in `vars/`
- **YAML metadata** — in `resources/metadata/`, drives code generation and docs
- **Maven** — Jenkins library JAR artifacts (`pom.xml`)
- **GitHub Actions** — primary CI/CD; Jenkins and Azure Pipelines also supported

## Key directory map

```
cmd/                  Step implementations
  stepName.go         Business logic — edit this
  stepName_generated.go  AUTO-GENERATED from YAML — never edit
  stepName_test.go    Tests — edit this
pkg/                  Internal Go packages (abaputils, btp, build, docker, orchestrator, …)
vars/                 Jenkins Groovy shared library scripts
resources/metadata/   Step YAML definitions — source of truth for params and code gen
integration/          Docker-based integration tests (30+ test files)
.pipeline/config.yml  Pipeline config for jenkins-library's own build
.github/workflows/    CI/CD workflows
```

## Code generation — the golden rule

**Never edit `*_generated.go` files.** They are overwritten on every `go generate` run.

Workflow for any parameter change:

1. Edit `resources/metadata/stepName.yaml`
2. Run `go generate ./...`
3. Edit `cmd/stepName.go` and `cmd/stepName_test.go` as needed

## Naming conventions

- Generic steps: no prefix — `dockerBuild`, `golangBuild`, `sonarExecuteScan`
- ABAP steps: `abap` prefix — `abapEnvironmentBuild`, `abapAddonAssemblyKitCheck`
- Files: `stepName.go`, `stepName_generated.go`, `stepName_test.go`, `stepName.groovy`, `stepName.yaml`
- Config scopes: `GENERAL` → `STAGES` → `STEPS` → `PARAMETERS` (most specific wins)
- Orchestrator detection via `pkg/orchestrator` — never assume Jenkins; support Jenkins, GitHub Actions, Azure Pipelines

## Versioning model

- Binary version embedded at build time via ldflags (`GitCommit`, `GitTag`)
- SemVer tags: `v1.510.0` format
- Published as `piper` binary on GitHub releases
- Mirrored to `github.tools.sap` and `github.wdf.sap.corp`
- Renovate auto-updates Go deps and Docker images

## Active constraints

- **Archived Dec 31 2026.** Do not add new features. Critical fixes only.
- **Apache 2.0 license + DCO required.** All contributors must sign the Developer Certificate of Origin. Automated check on every PR.
- **No SAP-proprietary code.** Generic utilities only. SAP-specific logic belongs in `piper-library`.
- **No hardcoded secrets.** Use Vault, Jenkins credentials, GitHub secrets, or System Trust.
- **Orchestrator abstraction is mandatory.** Use `pkg/orchestrator` to detect the CI environment — never assume Jenkins.
- **Backward compatibility matters.** This library has external open-source users. Breaking changes need discussion.

## Testing conventions

```bash
# Unit tests for a single step
go test ./cmd/stepName*.go

# All unit tests
go test ./...

# Integration tests (Docker required)
go test -tags=integration ./integration/...
```

- Unit tests in `cmd/*_test.go`, use `testify` and `httpmock`
- Integration tests in `integration/` use `testcontainers-go` — run actual tool invocations in Docker
- Listed in `integration/github_actions_integration_test_list.yml` for selective CI runs
- Groovy tests in `test/groovy/`
- CI runs via `.github/workflows/verify-go.yml` — includes format check, code-gen check, SonarQube, CodeClimate

## Frequent task patterns

**Adding a generic step (rarely needed — project is frozen):**

1. `resources/metadata/newStep.yaml` — define parameters
2. `go generate ./...`
3. `cmd/newStep.go` — business logic
4. `cmd/newStep_test.go` — unit tests
5. `integration/integration_newTool_test.go` — integration test
6. `vars/newStep.groovy` — Jenkins wrapper
7. Register in `cmd/piper.go`

**Fixing a bug:** edit `cmd/stepName.go` and its test. If parameters change, update YAML and run `go generate`.

**Supporting multiple orchestrators:** use `orchestrator.GetOrchestratorType()` and conditionally execute orchestrator-specific logic.

## Common gotchas

- `*_generated.go` files include the step's `Command()` function and `Options` struct. If the build fails after a YAML change, you forgot `go generate`.
- DCO sign-off is enforced automatically on PRs. Use `git commit -s` or add `Signed-off-by` manually.
- Jenkins library changes in `vars/` are live immediately on merge; Go binary changes require a release build.
- The project is a base for `piper-library` — changes here affect both libraries. Coordinate with the piper-library team for anything ABAP- or SAP-related.
- Integration tests can be slow and require Docker. Run selectively using the test list in `github_actions_integration_test_list.yml`.
