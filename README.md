<div align="center">

# DevDoctor

**Find out why a project won't run on your machine.**

[![Release](https://img.shields.io/github/v/release/paldentitung/devdoctor?style=flat-square)](https://github.com/paldentitung/devdoctor/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/paldentitung/devdoctor?style=flat-square)](https://go.dev)
[![Go Reference](https://img.shields.io/badge/go.dev-reference-007d9c?style=flat-square&logo=go&logoColor=white)](https://pkg.go.dev/github.com/paldentitung/devdoctor)
[![Go Report Card](https://goreportcard.com/badge/github.com/paldentitung/devdoctor?style=flat-square)](https://goreportcard.com/report/github.com/paldentitung/devdoctor)
[![License: MIT](https://img.shields.io/github/license/paldentitung/devdoctor?style=flat-square)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen?style=flat-square)](#contributing)

</div>

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

**Using Go**

Requires Go installed:

```bash
go install github.com/paldentitung/devdoctor@latest
```

**Download a prebuilt binary**

Don't have Go installed? Download the binary for your operating system and architecture from [GitHub Releases](https://github.com/paldentitung/devdoctor/releases).

Extract the archive and place the `devdoctor` executable somewhere in your system's `PATH`.

[View all releases](https://github.com/paldentitung/devdoctor/releases)

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

Contributions are welcome! Whether you're fixing a bug, improving documentation, or adding a new diagnostic check, your help makes DevDoctor better.

### Adding a new check

Every diagnostic check implements the `Check` interface.

To add one:

1. Create your check in the appropriate package under `internal/`.
2. Register it in `internal/registry`.
3. Add table-driven tests covering the expected behavior.
4. Run the test suite to make sure everything works:

   ```bash
   go test ./...
   ```

### Before submitting a PR

- Keep checks focused and consistent with existing implementations.
- Include tests for new functionality.
- Make sure all existing tests pass.
- Keep diagnostics read-only; DevDoctor should diagnose problems, not modify the user's environment.

### Submit your contribution

1. Fork the repository.
2. Create a branch for your changes.
3. Commit your changes.
4. Open a pull request with a clear description of what you changed and why.

Found a bug or have an idea? Feel free to open an issue before starting work.

Thanks for helping make DevDoctor better!

## License

Distributed under the [MIT License](LICENSE).

---

<div align="center">
  <sub>Built with Go · Licensed under MIT</sub>
</div>
