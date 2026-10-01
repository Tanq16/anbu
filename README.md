<div align="center">
  <img src=".github/assets/logo.svg" alt="Anbu Logo" width="200">
  <h1>Anbu</h1>

  <a href="https://github.com/tanq16/anbu/actions/workflows/release.yaml"><img alt="Build Workflow" src="https://github.com/tanq16/anbu/actions/workflows/release.yaml/badge.svg"></a>&nbsp;<a href="https://github.com/tanq16/anbu/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/tanq16/anbu"></a>&nbsp;<a href="https://hub.docker.com/r/tanq16/anbu"><img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/tanq16/anbu"></a><br><br>
  <a href="#features">Features</a> &bull; <a href="#install">Install</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#notes">Notes</a>
</div>

---

Anbu is a self-hosted IT hub: an encrypted secrets vault, a task list, a web SSH terminal, AWS access, and EC2 workstation management behind one web UI and one REST API.

It runs as a single binary for one person or a small team behind a forward-auth proxy. It has no login of its own and is not a multi-tenant password manager.

## Features

| Area | What it does |
|---|---|
| Vault | Typed secrets (login, SSH key, AWS static keys, AWS SSO, GitHub PAT, generic) with TOTP codes, custom fields, Ed25519 key generation, and plaintext export and import |
| Tasks | Single-line tasks with priority, due date, and overdue tracking |
| SSH | Stored hosts and EC2 machines in a browser terminal, with trust-on-first-use host keys |
| AWS | Static keys, ad-hoc keys, and SSO profiles through the device flow, plus an `aws` CLI runner |
| Machines | EC2 workstations on a per-account scaffold, with create, start, stop, resize, remove, live pricing, and a bootstrap probe |
| Tools | Hashes, YAML and JSON conversion, time parsing, UUIDs, passphrases, random strings, JWT, Base64, URL, case, and text stats |

## Install

### Docker

```bash
mkdir -p $HOME/.anbu && sudo chown 10001:10001 $HOME/.anbu
```
```bash
docker run -d --name anbu \
  -p 8080:8080 \
  -v $HOME/.anbu:/data \
  tanq16/anbu:latest
```

Available at `http://localhost:8080`. The same setup as a compose file:

```yaml
services:
  anbu:
    image: tanq16/anbu:latest
    container_name: anbu
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data # change as needed
```

The container runs as UID and GID `10001`, so the mounted directory must be writable by that user. The image includes the AWS CLI v2 for the command runner.

### Binary

Download a binary from [releases](https://github.com/tanq16/anbu/releases) for Linux or macOS on AMD64 or ARM64, then run `anbu serve`. The command runner needs the `aws` CLI on the server's `PATH`.

### From source

Requires Go 1.27 or newer, `curl`, and `uv` (for the Nerd Font asset).

```bash
git clone https://github.com/tanq16/anbu.git && cd anbu && make build
```

## Usage

```bash
anbu serve                 # http://0.0.0.0:8080, data in ~/.config/anbu/data
anbu serve -p 9000 -d /srv/anbu
```

### Data directory

`-d` names the data directory. It is created at `0700`, and every file in it is `0600`.

| Path | Holds |
|---|---|
| `password` | the vault password in plaintext, generated on first start |
| `vault.json` | every secret, AES-256-GCM encrypted under a PBKDF2 key from the password |
| `tasks.json`, `settings.json`, `hosts.json` | tasks, settings, and stored SSH hosts in plaintext |
| `known_hosts` | host keys for every SSH target |
| `aws/` | empty AWS config files and the `HOME` of the `aws` CLI |

Back up the whole directory. The vault cannot be read without the `password` file next to it. Rotate the password under Settings in the UI.

### Authentication

Anbu has no users, sessions, or tokens. Put it behind a forward-auth proxy that also passes WebSocket upgrades through for `/ws/terminal`. The UI needs HTTPS or `localhost` for its copy buttons to work.

### `anbu api`

The `api` command group calls the REST API and prints the raw JSON response.

```bash
anbu api setup https://anbu.example.com -H "X-Proxy-Token: <token>"
anbu api secrets list
anbu api secrets get github
anbu api secrets totp google-work
anbu api ssh targets
anbu api machines list corp:admin
```

- `setup` writes `~/.config/anbu/api.json`, and its headers ride on every call so the proxy admits the CLI.
- Without `api.json`, calls go to `http://localhost:8080`. `ANBU_URL` overrides the URL for one invocation.
- An HTTP error or a connection failure is logged and exits 1.

### AWS SSO in a local profile

An SSO profile is referenced as `<secret>:<profile>`. Once the SSO session is logged in through the UI, `anbu api secrets get` prints temporary credentials in the `credential_process` format:

```ini
[profile corp-admin]
credential_process = anbu api secrets get corp:admin
region = us-west-2
```

## Notes

- **SSO sessions** live in memory only. A restart needs a new device login, and an expired session returns `401` with `sso login required`.
- **EC2 jobs** run one at a time in a queue held in memory. A restart drops queued jobs and job history.
- **Scaffold keys** are vault secrets named `sharingan-<account>-<region>`. Only scaffold teardown deletes one. A sharingan-created scaffold is adopted by importing `~/.config/sharingan/id_ed25519` under that name.
- **Pricing** for machine options and the machine list is cached per account and region for an hour, so the first request in an hour is slow.
- **Host key changes** fail the connection. Clear the old key with Forget host key in the SSH view.
