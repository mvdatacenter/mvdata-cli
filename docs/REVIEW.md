# Review Guidelines — mvdata-cli

## Design goal

The CLI provides 1:1 coverage of every resource in the MV Data console API. Every object the SDK exposes should have a corresponding CLI command. When a new resource is added to `mvdata-sdk-go`, a matching command must be added here.

### Current coverage

| Console API resource | SDK type | CLI command | Operations |
|---------------------|----------|-------------|------------|
| VPC | `VPC` | `mvdata vpc` | create, get, delete |
| Subnet | `Subnet` | `mvdata subnet` | create, get, delete |
| Instance | `Instance` | `mvdata instance` | create, get, delete |
| SSH Key | `Key` | `mvdata key` | create, get, delete |
| Kubernetes Cluster | `KubernetesCluster` | `mvdata kubernetes` | create, get, update, delete |
| Instance Type | `InstanceType` | `mvdata instance-types` | list |

If a resource appears in `mvdata-sdk-go/mvdata/types.go` but has no CLI command, that is a gap.

## Adding a new resource command

1. Create `internal/cmd/<resource>.go` with subcommands matching the SDK methods (create/get/delete, plus update if the SDK has it)
2. Wire required flags — every SDK field that is part of a create request needs a `--flag`
3. Add a `<resource>Output()` function returning `*output.Table` for table format and the SDK struct directly for JSON format
4. Register the command in `internal/cmd/root.go` via `root.AddCommand()`
5. Write tests in `internal/cmd/<resource>_test.go` using the httptest helpers in `helpers_test.go`
6. Update README.md command list

## Review checklist

- [ ] Every required flag uses `cmd.MarkFlagRequired()`
- [ ] All commands go through `newClient()` for auth resolution
- [ ] Table output includes all user-relevant fields (skip internal IDs)
- [ ] JSON output passes the SDK struct directly (no re-serialization)
- [ ] Delete commands print a confirmation message, not a table
- [ ] Tests verify HTTP method, path, and output for each subcommand
- [ ] Tests verify missing required flags produce errors
- [ ] No `replace` directives in go.mod (CI must resolve deps from the module proxy)
