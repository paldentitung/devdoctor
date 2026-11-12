# DevDoctor

**Find out why a project won't run on your machine.**

Run `devdoctor` in a project folder. It detects the project type, reads the versions the project requires (`go.mod`, `package.json`, ...), compares them with what is installed, and tells you exactly what to fix.

> **Status:** early development. Flags and output may change.

## Example

```text
$ devdoctor

Detected project: go
  Requires: go 1.27.3  (go.mod)

Runtimes
✔ Go is installed: go version go1.27.3 linux/amd64

Tools
! Git 2.43.0 is older than the required 2.45.0.
    → Upgrade Git to 2.45.0 or newer: https://git-scm.com

Databases
! MongoDB is installed, but no server is listening on port 27017.
    → Start MongoDB: sudo systemctl start mongod

Passed: 4  Warnings: 2  Failed: 0
```

## Install

```bash
go install github.com/paldentitung/devdoctor@latest
```

## Usage

```bash
devdoctor                      # auto-detect the project and run checks
devdoctor --type node,docker   # override detection
devdoctor --only runtimes      # run a single group
devdoctor --json               # machine-readable output
```

| Flag         | Description                                         |
| ------------ | --------------------------------------------------- |
| `--type`     | Project types: node, typescript, python, go, docker |
| `--only`     | Run only the named group                            |
| `--json`     | Print results as JSON                               |
| `--no-color` | Disable colored output                              |
| `--version`  | Print version and exit                              |

## What it checks

| Group     | Checks                                        |
| --------- | --------------------------------------------- |
| Runtimes  | Go, Node, Python (versions vs. project files) |
| Tools     | Git, curl, Docker                             |
| Databases | PostgreSQL, MongoDB (installed and reachable) |
| Project   | Port availability                             |

## Exit codes

| Code | Meaning                             |
| ---- | ----------------------------------- |
| 0    | No failed checks (warnings allowed) |
| 1    | At least one check failed           |
| 2    | Invalid usage (e.g. bad `--type`)   |

## Contributing

A check is any type that implements the `Check` interface. To add one, create it in the matching package under `internal/`, register it in `internal/registry`, and add a table-driven test. Run `go test ./...` before opening a PR.

## License

MIT. See [LICENSE](LICENSE).
