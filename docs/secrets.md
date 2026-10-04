# Secrets

`mvdata secret-store` and `mvdata secret` manage an account's secret stores and the secrets in
them. `secret-stores` and `secrets` work as aliases.

## Reading a value

`secret get <store>/<name>` prints a secret's metadata: its store, name, version, sync state and
who last changed it. Only `--reveal` prints its value, and the console records each reveal in the
account's audit trail. In table output `--reveal` prints the value alone, followed by a newline;
under `-o json` it prints `{ storeName, name, version, value }`.

## Writing a value

`secret put <store>/<name>` creates the secret, or replaces its value if it exists. Every argument
and flag `put` takes names the secret or a file, and the value itself arrives by stdin, by a prompt
that hides what is typed when stdin is a terminal, or from `--from-file <path>`, so it stays out of
shell history and the process list. One trailing newline is dropped, because `echo` and most
editors add one; `--keep-newline` keeps it.

```
printf %s "$DB_PASSWORD" | mvdata secret put app/db
mvdata secret put app/db --from-file ./db-password
```

## Exit status

Every `mvdata` command that fails prints `Error: <message>` to stderr and exits with the status of
the console's error code:

| Status | Meaning |
|--------|---------|
| 1 | any failure the console does not classify |
| 3 | `not_found` |
| 4 | `forbidden` |
| 5 | `limit_exceeded`; the message names the limit and its current and maximum values |
| 6 | `conflict` |
| 7 | `invalid` |
| 8 | `unavailable`, including a console that cannot be reached |
