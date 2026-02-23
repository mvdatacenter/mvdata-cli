# mvdata-cli

Command-line interface for managing MV Data cloud resources.

## Installation

### From source

```bash
go install github.com/mvdatacenter/mvdata-cli@latest
```

### From release

Download the latest binary from the [releases page](https://github.com/mvdatacenter/mvdata-cli/releases).

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

## License

[The Unlicense](LICENSE)
