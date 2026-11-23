<div align="center">

# devkit

**A fast, portable, single-binary toolkit for everyday developer chores.**

Format JSON · Convert JSON ⇄ YAML · Check TCP ports

[![Test](https://github.com/hamrezoffic-source/devkit/actions/workflows/test.yml/badge.svg)](https://github.com/hamrezoffic-source/devkit/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/hamrezoffic-source/devkit?sort=semver)](https://github.com/hamrezoffic-source/devkit/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/hamrezoffic-source/devkit)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

</div>

---

## Why devkit?

- **One binary, no runtime.** Written in Go, statically built for Linux, macOS and Windows (amd64 and arm64).
- **Correct by default.** Large integers stay exact, numbers keep their type across JSON ⇄ YAML, duplicate JSON and YAML keys are rejected by validation/conversion.
- **Pipe-friendly.** Every command reads a file or standard input and writes clean output to standard output.
- **Script-friendly exit codes.** `0` on success, `1` on any error.

## Installation

**Pre-built binaries:** download the file for your platform from the [Releases page](https://github.com/hamrezoffic-source/devkit/releases).

**With Go (1.22+):**

```sh
go install github.com/hamrezoffic-source/devkit@latest
```

**From source:**

```sh
git clone https://github.com/hamrezoffic-source/devkit.git
cd devkit
go build -o devkit .
```

Check the installation:

```sh
devkit --version
```

## Quick start

```sh
# Validate a JSON config (silent on success; nonzero exit on invalid input)
devkit json check config.json

# Pretty-print JSON
echo '{"name":"devkit","tags":["go","cli"]}' | devkit json format

# JSON -> YAML
devkit json to-yaml config.json > config.yaml

# YAML -> JSON
devkit yaml to-json config.yaml

# Is the port open?
devkit port check example.com 443
```

## Commands

| Command | Description | Short alias |
| --- | --- | --- |
| `devkit json check [file\|-]` | Validate JSON and reject duplicate object keys | — |
| `devkit json format [file\|-]` | Pretty-print JSON with 2-space indentation | `devkit format` |
| `devkit json to-yaml [file\|-]` | Convert JSON to YAML | `devkit json-to-yaml` |
| `devkit yaml to-json [file\|-]` | Convert YAML to formatted JSON | `devkit yaml-to-json` |
| `devkit port check <host> <port> [--timeout s]` | Test whether a TCP port accepts connections | `devkit check-port` |
| `devkit --version` | Print the version | `devkit -v` |
| `devkit --help` | Show usage | `devkit -h` |

Omit the file argument, or pass `-`, to read from standard input.

### JSON

```sh
# Check a file or a stream; valid input produces no output and exits 0
devkit json check config.json
cat config.json | devkit json check -

# Format a file or a stream
devkit json format input.json
cat input.json | devkit json format -
```

`json check` rejects malformed input, multiple top-level values, and duplicate object keys (including nested objects). It produces no output on success and exits nonzero on failure, so it works in CI and shell scripts. Large integers such as `9007199254740993` are preserved without floating-point rounding by the formatter and converter. Input containing more than one JSON value is rejected.

### JSON ⇄ YAML

```console
$ echo '{"port":8080,"debug":true,"name":"api"}' | devkit json to-yaml
debug: true
name: api
port: 8080
```

- Numbers, booleans and `null` keep their types in both directions.
- Empty YAML input converts to `null`.
- Multiple YAML documents (`---`) become a JSON array.
- YAML keys must be strings; duplicate keys and non-finite numbers (`.inf`, `.nan`) are rejected because JSON cannot represent them.

### TCP port check

```console
$ devkit port check example.com 443
open
$ devkit port check 127.0.0.1 9 --timeout 1
closed
error: dial tcp 127.0.0.1:9: connect: connection refused
```

The default timeout is 3 seconds.

## Roadmap

- [x] JSON formatting, strict JSON validation, and JSON ⇄ YAML conversion
- [x] TCP port check
- [x] `--version` flag and cross-platform releases
- [ ] Split into `cmd/` and `internal/` packages
- [ ] Hashing (`sha256`, `sha1`, `md5`), Base64, UUID, JWT decode, Unix timestamps
- [ ] Concurrent multi-port scan (`80,443,8000-8010`)
- [ ] Colored output with `NO_COLOR` support
- [ ] GoReleaser, Homebrew and Scoop packages

Ideas and feedback are welcome in [Issues](https://github.com/hamrezoffic-source/devkit/issues).

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

## License

[MIT](LICENSE) © hamrezoffic-source
