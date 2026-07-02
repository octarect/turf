# turf

[![Go Report Card](https://goreportcard.com/badge/github.com/octarect/turf)](https://goreportcard.com/report/github.com/octarect/turf)

CLI tool for fetching JRA (Japan Racing Association) horse racing data.

## Installation

```sh
go install github.com/octarect/turf/cmd/turf@latest
```

Or download a pre-built binary from [GitHub Releases](https://github.com/octarect/turf/releases).

## Usage

### List fixtures

```sh
turf fixtures --month 2025-06
turf fixtures --date 2025-06-01 --course tokyo
```

### List races

```sh
turf races --date 2025-06-01 --course tokyo
```

### List races

```sh
turf races --date 2025-06-01 --course tokyo
```

### Get race result

```sh
turf result --date 2025-06-01 --course tokyo --race 11
```

## Output Formats

| Format | Flag | Description |
|---|---|---|
| Table | (default) | Human-readable table |
| JSON | `-o json` | JSON array or object |
| Custom columns | `-o custom-columns=HEADER:.path,...` | Pick specific fields |

### Custom columns example

```sh
turf result --date 2025-06-01 --course tokyo --race 11 \
  -o custom-columns="FP:.entries[*].finish.position,HORSE:.entries[*].horse.nameEN"
```

## Courses

`tokyo`, `kyoto`, `hanshin`, `nakayama`, `chukyo`, `kokura`, `sapporo`, `hakodate`, `fukushima`, `niigata`
