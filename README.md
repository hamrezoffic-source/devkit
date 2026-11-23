# devkit

`devkit` is a fast, portable developer toolkit for formatting JSON, converting
between JSON and YAML, and checking TCP ports.

## Install

Download a binary for Windows, Linux, or macOS from the GitHub Releases page,
or build it from source:

```sh
go build -o devkit .
```

## JSON

Format JSON from a file or standard input:

```sh
devkit json format input.json
cat input.json | devkit json format -
devkit format input.json
```

Convert JSON to YAML:

```sh
devkit json to-yaml input.json
cat input.json | devkit json-to-yaml -
```

Large JSON integers are preserved without floating-point rounding.

## YAML

Convert YAML to formatted JSON:

```sh
devkit yaml to-json input.yaml
cat input.yaml | devkit yaml-to-json -
```

Empty input converts to `null`. Multiple YAML documents become a JSON array.
YAML mapping keys must be strings, duplicate keys are rejected, and non-finite
numbers cannot be represented by JSON.

## TCP ports

Check whether a TCP port accepts connections:

```sh
devkit port check example.com 443
devkit port check 127.0.0.1 8080 --timeout 5
devkit check-port example.com 443
```

The command prints `open` on success and exits with a non-zero status on failure.
