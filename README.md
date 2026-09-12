<div align="center">
  <img src=".github/assets/logo.svg" alt="ANBU Logo" width="250"/>

  <h1>Anbu</h1>

  <a href="https://github.com/tanq16/anbu/actions/workflows/release.yaml"><img src="https://github.com/tanq16/anbu/actions/workflows/release.yaml/badge.svg" alt="Release Build"></a>&nbsp;<a href="https://github.com/tanq16/anbu/releases/latest"><img src="https://img.shields.io/github/v/release/tanq16/anbu" alt="Latest Release"></a><br>

<p><b>Anbu</b> is a CLI tool that helps perform everyday tasks in an expert way. Just like the Anbu Black Ops division in Naruto, this tool helps carry out all the shadow-operations in your daily workflow.</p><br>

<a href="#capabilities">Capabilities</a> &bull; <a href="#installation">Installation</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#tips-and-notes">Tips & Notes</a><br>

</div>

## Capabilities

A summary of everything that **Anbu** can perform:

| Operation | Details |
| --- | --- |
| **Time** | Current time via `now`, parse a timestamp, and epoch diffs, in one table of epoch, RFC 822 local, ISO 8601 local, ISO 8601 UTC, and human UTC |
| **Secrets** | Encrypted store for named secrets, with list, get, add, delete, import, and export |
| **SSH Sessions** | Named Ed25519 sessions under `~/.config/anbu/ssh/`, with list, setup, exec, and delete |
| **WireGuard Proxy** | Userspace WireGuard SOCKS5 proxy, with no root, TUN device, or host routing changes |
| **HTTP Download** | Multi-connection HTTP download with automatic fallback to a single connection |
| **GitHub Release** | Latest-release asset download, with platform auto-select or an explicit asset name |
| **HTTP Server** | Serve the current directory, or an upload page for text and files |
| **IP Information** | Local and public IP details, including geolocation |
| **Archive** | Zip files with include/exclude regex, optional AES-GCM encryption, and a wrap vs bare extract |
| **Bulk Rename** | Batch rename files or directories with regular expressions and capture groups |
| **Find Duplicates** | Duplicate files by size and SHA256, with optional recursive search |
| **Passphrase** | Diceware-style hyphenated phrase, with one capital letter and one digit by default |
| **UUID** | UUID v7 by default, v4 with `--v4`, and a short 18-character form that keeps only random bits |
| **Random String** | Cryptographic random string; alphanumeric by default, or hex, digits, letters, or all printable ASCII |

## Installation

- Download directly from [RELEASES](https://github.com/Tanq16/anbu/releases). Anbu is available for AMD64 and ARM64 for Linux and MacOS.
- To clone and build locally for development (requires Go 1.27 or newer), use:
  ```bash
  git clone https://github.com/tanq16/anbu.git && \
  cd anbu && \
  go build .
  ```

## Usage

Anbu supports a large number of operations across the board. All commands support the `--debug` flag to enable debug logging.

The specific details of each are:

- ***HTTP Download*** (alias: `dl`)

  Uses multiple connections when the server supports byte ranges and the file is large enough. Otherwise it falls back to a single connection. A partial `.anbu-temp` file is resumed automatically.

  ```bash
  anbu download https://example.com/file.tar.gz
  anbu dl https://example.com/file.tar.gz -o package.tar.gz
  anbu dl https://example.com/file.tar.gz -c 16
  anbu dl https://example.com/file.tar.gz -H "Authorization: Bearer token"
  anbu dl https://example.com/file.tar.gz --proxy http://127.0.0.1:8080
  ```

- ***GitHub Release*** (alias: `ghr`)

  Resolves `owner/repo`, a `github.com` URL, or `github.com/owner/repo`. Auto-selects the asset for this OS and architecture. `--manual` picks from a list; `--asset` names one for scripts. `GITHUB_TOKEN` is used when set.

  ```bash
  anbu github-release tanq16/anbu
  anbu ghr https://github.com/tanq16/anbu
  anbu ghr tanq16/anbu --asset anbu-linux-amd64
  anbu ghr tanq16/anbu --manual
  anbu ghr tanq16/anbu -o anbu.bin
  ```

- ***Time*** (alias: `t`)

  ```bash
  anbu time now                             # table for now
  anbu t parse "13 Apr 25 16:30 EDT"        # parse a timestamp into the same table
  anbu t until "13 Apr 25 16:30 EDT"        # how far that time is from now
  anbu t diff 1744192475 1744497775         # difference between two epochs
  anbu t diff 1744192475                    # difference between that epoch and now
  ```

- ***Secrets*** (alias: `p`)

  ```bash
  anbu secrets list
  anbu secrets list -f 'api'   # names matching the regex

  anbu secrets add API_KEY
  anbu secrets add API_KEY --multiline
  anbu secrets add API_KEY --value sk-1234
  echo "sk-1234" | anbu secrets add API_KEY --value-file -
  anbu secrets add API_KEY --value-file ./key.pem
  anbu secrets get API_KEY
  anbu secrets delete API_KEY

  anbu secrets export backup.json
  anbu secrets import backup.json
  ```

- ***SSH Sessions***

  Sessions and keys live under `~/.config/anbu/ssh/`. Exec uses the system `ssh` binary with a private known_hosts file and `StrictHostKeyChecking=accept-new`. A host already verified in `~/.ssh/known_hosts` is trust-on-first-use again here. Remote stderr is forwarded, and a remote command's exit code is preserved.

  ```bash
  anbu ssh setup prod --host 203.0.113.10 -u ubuntu
  anbu ssh setup lab --host lab.internal -u bob -p 2222
  anbu ssh list
  anbu ssh exec prod
  anbu ssh exec prod -c "uname -a"
  anbu ssh delete lab
  ```

- ***WireGuard Proxy*** (alias: `wgp`)

  Runs entirely in userspace. It does not create a `utun`/`wg0` interface and does not change host routes or DNS. Only clients that use the local proxy are sent through the tunnel.

  Keys may be standard WireGuard Base64 or 64-character hex. A wg-quick `.conf` supplies them as a file; flags override file values.

  ```bash
  anbu wgp --config-file ./wg.conf
  anbu wg-proxy -k "$WG_PRIVATE" -p "$WG_PEER" -e vpn.example.com:51820 -a 10.0.0.2

  curl -x socks5h://127.0.0.1:8888 https://icanhazip.com
  yt-dlp --proxy socks5://127.0.0.1:8888 "https://www.youtube.com/watch?v=..."
  ```

- ***HTTP Server***

  ```bash
  anbu http-server                  # current directory on http://0.0.0.0:8080
  anbu http-server -l 0.0.0.0:8080
  anbu http-server --upload         # upload page for text and files
  ```

- ***IP Information*** (alias: `ip`)

  ```bash
  anbu ip-info         # local and public IP information
  anbu ip-info --ipv6  # include IPv6
  ```

- ***Archive***

  Default create wraps entries in a folder named after the output file. Default extract writes those stored paths into the current directory. `--bare` on create skips the wrapper. `--bare` on extract strips the first path component. `--encrypt` wraps the zip in AES-GCM and writes a `.enc` file; that is not `zip -e`.

  ```bash
  anbu archive create ./src ./docs
  anbu archive c ./src -o backup.zip
  anbu archive create ./src --include '\.go$' --exclude '_test\.go$'
  anbu archive create ./src --bare
  echo pw | anbu archive create ./src --encrypt -
  anbu archive extract archive.zip.enc --password pw
  anbu archive e backup.zip
  anbu archive extract backup.zip --bare
  ```

- ***Bulk Rename***

  ```bash
  anbu rename 'prefix_(.*)' 'new_\1'
  anbu rename --directories 'old_(.*)' 'new_\1'
  anbu rename '(.*)\.(.*)' '\1_backup.\2'
  anbu rename 'image-(\d+).jpg' 'IMG_\1.jpeg' --dry-run
  anbu rename '(.*)' '\1_\uuid'
  anbu rename '(.*)\.(.*)' '\1_\suid.\2'
  ```

- ***Find Duplicates*** (alias: `dup`)

  ```bash
  anbu duplicates
  anbu dup --recursive
  anbu dup --delete
  ```

- ***Passphrase***

  Default is three hyphenated words, then one of those words is capitalized and one (possibly the same) gets a trailing digit. `--simple` is words and hyphens only.

  ```bash
  anbu passphrase
  anbu passphrase -l 5
  anbu passphrase --simple
  ```

- ***UUID***

  ```bash
  anbu uuid
  anbu uuid --v4
  anbu uuid --short      # 18-character form from v7 random bits
  anbu uuid --v4 --short # 18-character form from v4 random bits
  ```

- ***Random String***

  ```bash
  anbu random-string
  anbu random -l 32
  anbu random --hex
  anbu random --digits
  anbu random --alpha
  anbu random --all
  ```

## Tips and Notes

<details>
<summary><b>Use Anbu within Shell Commands</b></summary>

A command that takes a username and password leaves those values in shell history:

```bash
hypothetical --username admin --password sensitive
```

Pull them from the secrets store instead:

```bash
hypothetical --username $(anbu secrets get myuser) --password $(anbu secrets get mypw)
```

An alias of `anbu` as `a` keeps generators short:

```bash
hypothetical_command --uuid $(a uuid)
```

</details>

<details>
<summary><b>Copy an SSH session public key</b></summary>

`anbu ssh setup` prints the public key. That line is what goes into the server's `authorized_keys`. Exec then uses the matching private key from `~/.config/anbu/ssh/keys/`.

</details>
