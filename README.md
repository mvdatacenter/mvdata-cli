# mvdata-cli

Command-line interface for managing MV Data cloud resources.

## Goal

This CLI provides 1:1 coverage of every resource in the MV Data console API. It is a thin wrapper around [mvdata-sdk-go](https://github.com/mvdatacenter/mvdata-sdk-go) using [cobra](https://github.com/spf13/cobra). Every resource the SDK exposes gets a corresponding CLI command. When the SDK adds a new resource, the CLI adds a matching command.

### Coverage

| Console API resource | CLI command | Operations |
|---------------------|-------------|------------|
| VPC | `mvdata vpc` | create, get, delete |
| Subnet | `mvdata subnet` | create, get, delete |
| Instance | `mvdata instance` | create, get, delete |
| SSH Key | `mvdata key` | create, get, delete |
| Kubernetes Cluster | `mvdata kubernetes` | create, get, update, delete |
| Instance Type | `mvdata instance-types` | list |

## Installation

### From source

```bash
go install github.com/mvdatacenter/mvdata-cli@latest
```

### From release

Download the latest binary from the [releases page](https://github.com/mvdatacenter/mvdata-cli/releases). Builds are available for Linux, macOS, and Windows (amd64 + arm64).

## Authentication

The CLI resolves credentials in this order:

1. **Flags** — `--api-url` and `--api-token`
2. **Environment variables** — `MVDATA_API_URL` and `MVDATA_API_TOKEN`
3. **Config file** — `~/.mvdata/config.json`

```json
{
  "api_url": "https://console.mvdatacenter.com",
  "api_token": "your-api-token"
}
```

## Commands

```
mvdata vpc create      --name <name>
mvdata vpc get         --name <name>
mvdata vpc delete      --name <name>

mvdata subnet create   --name <name> --vpc <vpc> --cidr <cidr>
mvdata subnet get      --name <name>
mvdata subnet delete   --name <name>

mvdata instance create --name <name> --vpc <vpc> --type <type> --key <key>
mvdata instance get    --name <name> --vpc <vpc>
mvdata instance delete --name <name> --vpc <vpc>

mvdata key create      --name <name> --key <public-key>
mvdata key get         --name <name>
mvdata key delete      --name <name>

mvdata kubernetes create  --name <name> --version <v> --node-type <type> --node-count <n>
mvdata kubernetes get     --name <name>
mvdata kubernetes update  --name <name> --node-count <n>
mvdata kubernetes delete  --name <name>

mvdata instance-types list

mvdata version
```

### Global flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `table` | Output format: `table` or `json` |
| `--api-url` | | | MV Data API URL |
| `--api-token` | | | MV Data API token |

## Development

```bash
git clone git@github.com:mvdatacenter/mvdata-cli.git
cd mvdata-cli
make build
make test
```

For local development against an unreleased SDK, add a replace directive:

```
go mod edit -replace github.com/mvdatacenter/mvdata-sdk-go=../mvdata-sdk-go
```

Remove it before committing — CI requires deps resolve from the module proxy.

## License

[The Unlicense](LICENSE)
