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
| **Time Operations** | Display current time in various formats, calculate time differences, and parse time strings |
| **Secrets Management** | Securely store and retrieve secrets with encryption at rest |
| **Key Pair Generation** | Generate RSA key pairs in PEM or OpenSSH format with strict permissioning |
| **Network Tunneling** | Create TCP and SSH tunnels (forward and reverse) to securely access remote services |
| **WireGuard Proxy** | Userspace WireGuard HTTP CONNECT and SOCKS5 proxy on one port, with no root, TUN device, or host routing changes |
| **Simple HTTP/HTTPS Server** | Host a simple webserver over HTTP/HTTPS or serve an upload page for text and file uploads |
| **IP Information** | Display local and public IP details, including geolocation information |
| **Bulk Rename** | Batch rename files or directories using regular expression patterns, supporting capture groups |
| **Find Duplicates** | Find duplicate files by comparing file sizes and SHA256 hashes, with support for recursive search |
| **String Generation** | Generate random strings, UUIDs, passwords, and passphrases for various purposes |
| **Stash** | Persistent clipboard for files, folders, and text snippets with apply, pop, and clear operations, almost similar to `git` stash |
| **AWS Helper Utilities** | Configure AWS SSO with IAM Identity Center, SAML direct login, and generate console URLs from CLI profiles |

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

- ***Time Operations*** (alias: `t`)

  ```bash
  anbu time          # prints time in various formats
  anbu t now         # prints time in various formats
  anbu t iso         # prints current time in ISO format for script-usage
  anbu time purple   # print time and public IP for purple teams
  anbu t diff -e 1744192475 -e 1744497775   # print time difference between 2 epochs
  anbu t parse -t "13 Apr 25 16:30 EDT"     # read given time and print in multiple formats
  anbu time until -t "13 Apr 25 16:30 EDT"  # read time and print difference from now
  anbu t parse -t "13 Apr 25 16:30 EDT" -p purple  # parse time and print in purple team format
  ```

- ***Secrets Management*** (alias: `p`)

  ```bash
  anbu pass list  # List all secrets

  # Managing Secrets (Default password used or provide yours with --password)
  anbu pass add API_KEY                 # Create a new secret (encrypted with AES GCM at rest)
  anbu pass add API_KEY --multiline     # Create a new multi-line secret via the editor
  anbu pass add API_KEY --value sk-1234 # Set the value from a flag
  echo "sk-1234" | anbu pass add API_KEY --value -  # Add from piped stdin
  anbu pass add API_KEY --value-file ./key.pem      # Set the value from a file
  anbu pass get API_KEY     # Retrieve a secret (decrypted value)
  anbu pass delete API_KEY  # Delete a secret

  # Import and Export to file
  anbu pass export backup.json  # Export to a file (secrets are decrypted)
  anbu pass import backup.json  # Import from a file
  ```

- ***Key Pair Generation*** (alias: `kp`)

  ```bash
  anbu key-pair                      # Generate a 2048-bit RSA PEM key pair (anbu-key, anbu-key.pub)
  anbu kp -o ~/.ssh/id_rsa -k 4096   # Generate a 4096-bit RSA PEM key pair at specified output path
  anbu kp --ssh -o ~/.ssh/id_ed25519 # Generate key pair in OpenSSH format
  ```

- ***Network Tunneling***

  ```bash
  # forward TCP tunnels
  anbu tunnel tcp -l localhost:8000 -r example.com:80
  anbu tunnel tcp -l localhost:4430 -r example.com:443 --tls --insecure

  # forward SSH tunnels
  anbu tunnel ssh -l localhost:8000 -r target.com:3306 -s ssh.vm.com:22 -u bob -p "builder"
  anbu tunnel ssh -l localhost:8000 -r target.com:3306 -s ssh.vm.com:22 -u bob -k ~/.ssh/mykey

  # reverse SSH tunnels
  anbu tunnel rssh -l localhost:3389 -r 0.0.0.0:8080 -s ssh.vm.com:22 -u bob -p "builder"
  ```

- ***WireGuard Proxy*** (alias: `wgp`)

  Runs entirely in userspace. It does not create a `utun`/`wg0` interface and does not change host routes or DNS. Only clients that use the local proxy are sent through the tunnel.

  Keys may be standard WireGuard Base64 or 64-character hex.

  ```bash
  anbu wg-proxy -k "$WG_PRIVATE" -p "$WG_PEER" -e vpn.example.com:51820 -a 10.0.0.2
  anbu wgp --private-key "$WG_PRIVATE" --peer-key "$WG_PEER" --endpoint 203.0.113.10:51820 --address 10.0.0.2 --listen 127.0.0.1:1080

  # HTTP CONNECT
  curl -x http://127.0.0.1:1080 https://icanhazip.com

  # SOCKS5 (same port)
  curl -x socks5h://127.0.0.1:1080 https://icanhazip.com
  yt-dlp --proxy socks5://127.0.0.1:1080 "https://www.youtube.com/watch?v=..."
  ```

- ***Simple HTTP/HTTPS Server***

  ```bash
  anbu http-server                           # Serves current directory on http://0.0.0.0:8080
  anbu http-server -l 0.0.0.0:8080 --tls     # Serve HTTPS on given addr:port with a self-signed cert
  anbu http-server --upload                  # Serve simple upload page for text and files
  anbu http-server --upload --tls            # Serve upload page over HTTPS with self-signed cert
  ```

- ***IP Information*** (alias: `ip`)

  ```bash
  anbu ip-info         # Print local and public IP information
  anbu ip-info --ipv6  # Print local (IPv4 & IPv6) and public IP information
  ```

- ***Bulk Rename***

  ```bash
  anbu rename 'prefix_(.*)' 'new_\1'                     # Rename files matching regex pattern
  anbu rename --directories 'old_(.*)' 'new_\1'          # Rename directories instead of files
  anbu rename '(.*)\.(.*)' '\1_backup.\2'                # Add _backup before extension
  anbu rename 'image-(\d+).jpg' 'IMG_\1.jpeg' --dry-run  # Perform a dry-run without renaming
  anbu rename '(.*)' '\1_\uuid'                    # Append UUID to filenames
  anbu rename '(.*)\.(.*)' '\1_\suid.\2'           # Insert short UUID before extension
  ```

- ***Find Duplicates*** (alias: `dup`)

  ```bash
  anbu duplicates                 # Find duplicate files in the current directory
  anbu dup --recursive            # Find duplicate files recursively in subdirectories
  anbu dup --delete               # Find and delete duplicate files
  ```

- ***String Generation*** (alias: `s`)

  ```bash
  anbu string random 23        # generate 23 (100 if not specified) random alphanumeric chars
  anbu string seq 29           # prints "abcdefghijklmnopqrstuvwxyz" until desired length
  anbu string rep 23 str2rep   # prints "str2repstr2rep...23 times"

  anbu string uuid     # generates a uuid
  anbu string ruid 16  # generates a short uuid of length b/w 1-30
  anbu string suid     # generates a short uuid of length 18

  anbu string password           # generate a 12-character complex password
  anbu string password 16        # generate a 16-character complex password
  anbu string password 8 --simple  # generate an 8-letter lowercase password

  anbu string passphrase               # generate a 3-word passphrase with hyphens
  anbu string passphrase 5             # generate a 5-word passphrase with hyphens
  anbu string passphrase 4 --separator '@'  # generate a 4-word passphrase with a custom separator
  ```

- ***Stash***

  ```bash
  # Stash a file or folder (keeps original, unlike git stash)
  anbu stash fs /path/to/file.txt
  anbu stash fs ./my-folder

  # Stash text interactively (multiline TUI editor)
  anbu stash text my-snippet

  # Stash text from a flag
  anbu stash text my-snippet --text "the snippet"

  # List all stashed entries
  anbu stash list

  # Apply a stash without removing it (text prints to stdout, files/folders extracted to current directory)
  anbu stash apply 1

  # Apply a stash and remove it (pop operation)
  anbu stash pop 1

  # Remove a stash without applying it
  anbu stash clear 1
  ```

- ***AWS Helper Utilities***

  ```bash
  # Configure AWS SSO with IAM Identity Center for multi-role access
  # This will create profiles in ~/.aws/config for all accounts and roles
  anbu aws iidc-login -u https://my-sso.awsapps.com/start -r us-east-1
  anbu aws iidc-login --start-url https://my-sso.awsapps.com/start --sso-region us-east-1 --cli-region us-west-2 --session-name my-sso

  # Login with a SAML response captured from a browser session
  anbu aws saml-direct-login -r ROLE_ARN -i PRINCIPAL_ARN -p my-profile
  anbu aws saml-direct-login -r ROLE_ARN -i PRINCIPAL_ARN --file assertion.xml

  # Generate AWS console URL from a local CLI profile (valid for up to 12 hours)
  anbu aws cli-ui -p my-profile
  ```

## Tips and Notes

<details>
<summary><b>Connecting Two NAT-hidden Machines via Public VPS</b></summary>

*Machine A* &rarr;
```bash
anbu tunnel rssh -l localhost:3389 -r 0.0.0.0:8001 -s vps.example.com:22 -u bob -p builder
```

*Machine B* &rarr;
```bash
anbu tunnel ssh -l localhost:3389 -r localhost:8001 -s vps.example.com:22 -u bob -p builder
```

Now, connecting to `localhost:3389` on Machine B will allow access to Machine A's 3389.

</details>

<details>
<summary><b>Creating a Secure Database (or service) Connection Tunnel</b></summary>

When working with remote databases or services that don't allow direct access, this method can enable connections. Create an SSH tunnel to the database server:

```bash
anbu tunnel ssh -l localhost:3306 -r db.internal.network:3306 -s jumpbox.vpn.com:22 -u bob -p builder
```

Now, connect your database client to localhost:3306, which will forward requests via the SSH forward proxy through the jumphost:

```bash
mysql -u dbuser -p -h localhost -P 3306
```

This allows a connection to restricted databases while maintaining security best practices.

</details>

<details>
<summary><b>Use Anbu within Shell Commands</b></summary>

It's quite helpful to use Anbu within shell commands for simple things like UUIDs or for more sensitive things like secrets. Imagine a shell script that requires a username and password:

```bash
hypothetical --username admin --password sensitive
```

Using such commands leaves credentials within the shell history and is not safe for screen sharing. Instead of exposing secrets here, we can use `anbu`:

```bash
hypothetical --username $(anbu pass get myuser) --password $(anbu pass get mypw)
```

Furthermore, you can create an alias for `anbu` as `a` and use it to say generate a UUID like so:

```bash
hypothetical_command --uuid $(a s uuid)
```

</details>
