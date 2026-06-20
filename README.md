# cayswap 🔑

Automated, painless WireGuard public-key exchange for hub-and-spoke networks.

[![Go Reference](https://img.shields.io/badge/GoDoc-reference-007d9c)](https://pkg.go.dev/github.com/taigrr/cayswap)
[![Go Report Card](https://goreportcard.com/badge/github.com/taigrr/cayswap)](https://goreportcard.com/report/github.com/taigrr/cayswap)

## What it does

Bringing a new node into a WireGuard mesh means copying public keys back and
forth by hand: generate a key on the spoke, paste it into the hub's config,
copy the hub's key back to the spoke, and reload both interfaces. It's tedious
and easy to get wrong, especially when machines are provisioned automatically.

**cayswap** automates that handshake. The hub runs a short-lived server; each
spoke runs a single `swap` command. They authenticate with a shared key,
exchange public keys over HTTP, write each other into the local
`/etc/wireguard/<device>.conf`, and reload `wg-quick`. The tunnel comes up with
no manual editing.

It's designed to be driven from provisioning tooling such as Terraform or
cloud-init: stand the server up for a few minutes, point your new nodes at it,
and tear it down.

## How it works

```
                       shared auth key
                              │
   ┌──────────┐  POST /key  ┌─┴────────┐
   │  spoke   │ ───────────▶│   hub    │   cayswap serve
   │  (swap)  │◀─────────── │ (server) │   - validates the auth key
   └──────────┘  hub pubkey └──────────┘   - records the spoke as a /32 peer
        │                         │        - replies with its own pubkey + subnet
        ▼                         ▼
  wg0.conf updated          wg0.conf updated
  wg-quick@wg0 reloaded     wg-quick@wg0 reloaded
```

1. The spoke reads its local WireGuard config, derives its public key, and posts
   it (plus its address as a single `/32` host) to the hub.
2. The hub authenticates the request with a constant-time key comparison,
   refuses duplicates, and appends the spoke as a peer.
3. The hub replies with its own public key and its subnet (reduced to the
   network address, e.g. `10.0.0.0/24`) so the spoke routes the whole network
   through the hub.
4. Both ends reload `wg-quick@<device>` and the tunnel establishes.
5. The server shuts itself down after 15 minutes to keep the exchange window
   small.

## Install

```bash
go install github.com/taigrr/cayswap/cmd/cayswap@latest
```

cayswap manages files in `/etc/wireguard/` and reloads systemd units, so it
must run as **root**.

## Usage

### On the hub

```bash
cayswap serve --auth "$SHARED_KEY" --device wg0 --interface 0.0.0.0:5150
```

| Flag | Default | Description |
|------|---------|-------------|
| `-k`, `--auth` | – | shared authentication key (required) |
| `-d`, `--device` | `wg0` | WireGuard interface to manage |
| `-i`, `--interface` | `0.0.0.0:5150` | address to listen on |
| `--restart` | `true` | reload `wg-quick@` after updates |

The server exits automatically after 15 minutes.

### On each spoke

```bash
cayswap swap \
  --auth "$SHARED_KEY" \
  --server-endpoint hub.example.com:5150 \
  --wireguard-endpoint hub.example.com:51820 \
  --device wg0
```

| Flag | Default | Description |
|------|---------|-------------|
| `-k`, `--auth` | – | shared authentication key (required) |
| `-s`, `--server-endpoint` | – | hub's cayswap API address |
| `-w`, `--wireguard-endpoint` | – | hub's WireGuard endpoint to dial |
| `-d`, `--device` | `wg0` | WireGuard interface to manage |
| `--restart` | `true` | reload `wg-quick@` after updates |

## Configuration

The auth key may be supplied on the command line, via environment variable, or
through a config file at `/etc/cayswap/cayswap.yaml` (override with `--config`):

```yaml
auth: your-shared-key
```

Environment variables are read automatically (e.g. `AUTH=your-shared-key`).
Precedence is: `--auth` flag → environment → config file.

## Security notes

- The shared key is compared in constant time to avoid timing attacks.
- Keep the auth key secret and rotate it between provisioning runs; it is only
  meant to be valid for the few minutes the server is up.
- The exchange currently happens over plain HTTP — run it on a trusted/private
  network or behind a tunnel, and bring the server down as soon as your nodes
  have swapped.

## Development

```bash
go build ./...
go test -race ./...
go vet ./...
staticcheck ./...
```

The CLI lives in [`cmd/cayswap`](cmd/cayswap); the WireGuard config
parser/writer is in [`wg/parser`](wg/parser), and the HTTP key-exchange handler
is in [`api`](api).

## License

See the repository for license details.
