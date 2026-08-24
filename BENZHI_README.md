# MineAir

MineAir coordinates mine ventilation, gas sampling, extraction, sectional power,
and evacuation state. It persists control events to local files and serves four
operator pages backed by live HTTP APIs.

Run with Go 1.26.2:

```text
go run ./cmd/mineair -addr 127.0.0.1:21206 -data ./data
```

The service exposes `/healthz`, `/ventilation`, `/gas-zones`, `/extraction`, and
`/incidents`. Offline builds use the checked-in `vendor` directory.
