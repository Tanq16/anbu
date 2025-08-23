<div align="center">
  <img src=".github/assets/logo.png" alt="ANBU Logo" width="250"/>

  <h1 align="center">Anbu</h1>

  <a href="https://github.com/tanq16/anbu/actions/workflows/release.yml"><img src="https://github.com/tanq16/anbu/actions/workflows/release.yml/badge.svg" alt="Release Build"></a>&nbsp;<a href="https://github.com/tanq16/anbu/releases/latest"><img src="https://img.shields.io/github/v/release/tanq16/anbu" alt="Latest Release"></a><br>

<p><b>Anbu</b> is a CLI tool that helps perform everyday tasks in an expert way. Just like the Anbu Black Ops division in Naruto, this tool helps carry out all the shadow-operations in your daily workflow.</p><br>

<a href="\#installation">Installation</a> • <a href="\#usage">Usage</a> • <a href="\#acknowledgements">Acknowledgements</a> • <a href="\#tips--tricks">Tips & Tricks</a><br>

</div>

A summary of all capabilities that **Anbu** can perform:

| Operation | Details |
| --- | --- |
| **Time Operations** | Display current time in various formats, calculate time differences, and parse time strings |
| **Secrets Management** | Securely store, retrieve, and serve secrets with encryption at rest |
| **Network Tunneling** | Create TCP and SSH tunnels (forward and reverse) to securely access remote services |
| **Simple HTTP/HTTPS Server** | Host a simple webserver over HTTP/HTTPS with optional file upload capability |
| **Secrets Scan** | Find common secrets in file systems using regular expressions |
| **IP Information** | Display local and public IP details, including geolocation information |
| **Bulk Rename** | Batch rename files or directories using regular expression patterns |
| **Data & Encoding Conversion** | Convert between data formats (YAML/JSON), decode JWTs, and handle various encodings (Base64, Hex, URL) |
| **File Encryption/Decryption** | Secure file encryption and decryption with AES-256-GCM symmetric encryption |
| **RSA Key Pair Generation** | Create RSA key pairs for encryption or SSH authentication |
| **String Generation** | Generate random strings, UUIDs, passwords, and passphrases for various purposes |

## Installation

- Download directly from [RELEASES](https://github.com/Tanq16/anbu/releases). Anbu is available for AMD64 and ARM64 for Linux and MacOS.
- To build latest commit directly via Go, use:
  ```bash
  go install github.com/tanq16/anbu@latest
  ```
- To clone and build locally for development, use:
  ```bash
  git clone https://github.com/tanq16/anbu.git && \
  cd anbu && \
  go build .
  ```

## Usage

Anbu supports a large number of operations across the board. The specific details of each are:

- ***Time Operations*** (alias: `t`)

  ```bash
  anbu time          # prints time in various formats
  anbu time now      # prints time in various formats
  anbu time purple   # print time and public IP for purple teams
  anbu time diff -e 1744192475 -e 1744497775  # print time difference between 2 epochs
  anbu time diff -e 1744192475 -e 1744497775  # print time difference between 2 epochs
  ```

  ```bash
  anbu time diff -e 1744192475 -e 1744497775  # print time difference between 2 epochs
  ```

  ```bash
  anbu time parse -t "13 Apr 25 16:30 EDT"  # read given time and print in multiple formats
  anbu time until -t "13 Apr 25 16:30 EDT"  # read time and print difference from now
  ```

- ***Secrets Management*** (alias: `p`)

  ```bash
  anbu pass list  # List all secrets

  # Managing Secrets (Password asked or from ANBUPW env var)
  anbu pass add API_KEY     # Create a new secret (encrypted with AES GCM at rest)
  anbu pass add API_KEY -m  # Create a new multi-line secret
  anbu pass get API_KEY     # Retrieve a secret (decrypted value)
  anbu pass delete API_KEY  # Delete a secret

  # Import and Export to file
  anbu pass export backup.json  # Export to a file (secrets are decrypted)
  anbu pass import backup.json  # Import from a file

  # Serve secrets over an API
  anbu pass serve
  # Interact with a remote server
  anbu pass --remote http://127.0.0.1:8080 list
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

- ***Simple HTTP/HTTPS Server***

  ```bash
  anbu http-server                     # Serves current directory on http://localhost:8000
  anbu http-server -l 0.0.0.0:8080 -t  # Serve HTTPS on given add:port with a self-signed cert
  anbu http-server -t --domain my.dev  # Serve HTTPS with a specific domain in the cert
  anbu http-server -u                  # Enables file upload via PUT requests
  ```

- ***Secrets Scan***

  ```bash
  anbu secret-scan                 # Scans current directory for secrets based on regex matches
  anbu secret-scan ./path/to/scan  # Scans path for secrets based on regex matches
  anbu secret-scan ./path -p       # Scans path with generic matches table (maybe false positive)
  ```

- ***IP Information*** (alias: `ip`)

  ```bash
  anbu ip-info      # Print local and public IP information
  anbu ip-info -6   # Print local (IPv4 & IPv6) and public IP information
  ```

- ***Bulk Rename***

  ```bash
  anbu rename 'prefix_(.*)' 'new_\1'        # Rename files matching regex pattern
  anbu rename -d 'old_(.*)' 'new_\1'        # Rename directories instead of files
  anbu rename '(.*)\.(.*)' '\1_backup.\2'   # Add _backup before extension
  anbu rename 'image-(\d+).jpg' 'IMG_\1.jpeg' -r # Perform a dry-run without renaming
  ```

- ***Data & Encoding Conversion*** (alias: `c`)

  ```bash
  # File format conversion
  anbu convert yaml-json config.yaml  # Convert YAML file to JSON
  anbu convert json-yaml data.json    # Convert JSON file to YAML

  # Encoding conversion
  anbu convert b64 "Hello World"              # Convert text to base64
  anbu convert b64d "SGVsbG8gV29ybGQ="        # Decode base64 to text
  anbu convert hex "Hello World"              # Convert text to hex
  anbu convert hexd "48656c6c6f20576f726c64"  # Decode hex to text
  anbu convert url "Hello World"              # URL encode text
  anbu convert urld "Hello%20World"           # URL decode text

  # Cross-encoding conversion
  anbu convert b64-hex "SGVsbG8gV29ybGQ="        # Convert base64 to hex
  anbu convert hex-b64 "48656c6c6f20576f726c64"  # Convert hex to base64

  # JWT Decoding
  anbu convert jwtd "$TOKEN"  # Decodes and prints the headers and payload
  ```

- ***File Encryption/Decryption*** (alias: `fc`)

  ```bash
  anbu file-crypt /path/to/file.zip -p "P@55w0rd"    # Encrypt a file
  anbu file-crypt /path/to/encrypted.enc -p "P@55w0rd" -d # Decrypt a file
  ```

- ***RSA Key Pair Generation***

  ```bash
  anbu key-pair -o mykey -k 4096  # 4096 bit RSA key pair in PEM format
  anbu key-pair -s -o anbu-key    # 2048 bit RSA key pair in OpenSSH format
  ```

- ***String Generation*** (alias: `s`)

  ```bash
  anbu string 23               # generate 23 (100 if not specified) random alphanumeric chars
  anbu string seq 29           # prints "abcdefghijklmnopqrstuvxyz" until desired length
  anbu string rep 23 str2rep   # prints "str2repstr2rep...23 times"

  anbu string uuid     # generates a uuid
  anbu string ruid 16  # generates a short uuid of length b/w 1-32
  anbu string suid     # generates a short uuid of length 18

  anbu string password           # generate a 12-character complex password
  anbu string password 16        # generate a 16-character complex password
  anbu string password 8 simple  # generate an 8-letter lowercase password

  anbu string passphrase               # generate a 3-word passphrase with hyphens
  anbu string passphrase 5             # generate a 5-word passphrase with hyphens
  anbu string passphrase 4 '@'         # generate a 4-word passphrase with a custom separator
  ```

## Tips & Tricks

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
hypothetical --neo4j-username neo4j --neo4j-password sensitive
```

Using such commands leaves credentials within the shell history and is not safe for screen sharing. Instead of exposing secrets here, we can use `anbu`:

```bash
hypothetical --neo4j-username $(anbu pass get n4jun) --neo4j-password $(anbu pass get neo4jpw)
```

Furthermore, you can create an alias for `anbu` as `a` and use it to say generate a UUID like so:

```bash
hypothetical_command --uuid $(a s uuid)
```

</details>

## Acknowledgements

Anbu takes inspiration from the following projects:

- [GoST](https://github.com/ginuerzh/gost)
- [SimpleHTTPServer](https://github.com/projectdiscovery/simplehttpserver)
- [TruffleHog](https://github.com/trufflesecurity/trufflehog), [GitLeaks](https://github.com/gitleaks/gitleaks), and [NoseyParker](https://github.com/praetorian-inc/noseyparker) for secret regular expressions
