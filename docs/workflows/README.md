# Workflows

How alGOrithm's processes actually run, one document per process: adding an entry, testing and benchmarking it, the development loop, and the checks that run locally, in CI and at release. Each follows the real files, says why each step is built the way it is, and shows real output.

| Workflow | In one line |
|----------|-------------|
| [adding-an-entry](adding-an-entry.md) | An algorithm, data structure or simulation from a new file to a registered, tested, documented entry |
| [testing-and-benchmarking](testing-and-benchmarking.md) | The test layers (shared suites, package tests, registry, command, generated docs, site), how to write a test, benchmarks |
| [dev-loop](dev-loop.md) | `make dev`: build, serve, watch, rebuild, test, reload |
| [ci-and-checks](ci-and-checks.md) | `make ci`, the hooks, the Tests, Lint, Web and Release workflows, common failures, releasing |

## Where to start

| If you want to | Read |
|----------------|------|
| Add an algorithm, structure or simulation | [adding-an-entry](adding-an-entry.md) |
| Work on code and see it in the site at once | [dev-loop](dev-loop.md) |
| Fix or write tests | [testing-and-benchmarking](testing-and-benchmarking.md) |
| Fix a red CI run, or cut a release | [ci-and-checks](ci-and-checks.md) |

## Conventions

These documents describe how; the rules are in [rules](../rules/README.md). Paths are relative to the repository root, markdown is standard, diagrams are plain text ([rule](../rules/diagrams-are-plain-text.md)), spelling is US and there are no em dashes. When a change alters a process described here, update the document in the same change.
