<div align="center">
  <img src=".github/assets/logo.svg" alt="ANBU Logo" width="250"/>

  <h1 align="center">Anbu</h1>

  <a href="https://github.com/tanq16/anbu/actions/workflows/release.yml"><img src="https://github.com/tanq16/anbu/actions/workflows/release.yml/badge.svg" alt="Release Build"></a>&nbsp;<a href="https://github.com/tanq16/anbu/releases/latest"><img src="https://img.shields.io/github/v/release/tanq16/anbu" alt="Latest Release"></a><br>

<p><b>Anbu</b> is a CLI tool that helps perform everyday tasks in an expert way. Just like the Anbu Black Ops division in Naruto, this tool helps carry out all the shadow-operations in your daily workflow.</p><br>

<a href="\#installation">Installation</a> • <a href="\#usage">Usage</a> • <a href="\#tips--tricks">Tips & Tricks</a><br>

</div>

A summary of everything that **Anbu** can perform:

| Operation | Details |
| --- | --- |
| **Time Operations** | Display current time in various formats, calculate time differences, and parse time strings |
| **Secrets Management** | Securely store and retrieve secrets with encryption at rest |
| **Network Tunneling** | Create TCP and SSH tunnels (forward and reverse) to securely access remote services |
| **Simple HTTP/HTTPS Server** | Host a simple webserver over HTTP/HTTPS or serve an upload page for text and file uploads |
| **Secrets Scan** | Find common secrets in file systems using regular expressions |
| **IP Information** | Display local and public IP details, including geolocation information |
| **Bulk Rename** | Batch rename files or directories using regular expression patterns, supporting capture groups |
| **Manual Rename** | Interactively rename files and directories one by one with TUI-style inline input |
| **Find Duplicates** | Find duplicate files by comparing file sizes and SHA256 hashes, with support for recursive search |
| **Bulk Sed (Regex Substitution)** | Apply regex pattern matching and replacement to file content, supporting capture groups |
| **Data & Encoding Conversion** | Convert between data formats (YAML/JSON), decode JWTs, and handle various encodings (Base64, Hex, URL) |
| **File Encryption/Decryption** | Secure file encryption and decryption with AES-256-GCM symmetric encryption |
| **RSA Key Pair Generation** | Create RSA key pairs for encryption or SSH authentication |
| **String Generation** | Generate random strings, UUIDs, passwords, and passphrases for various purposes |
| **Stash** | Persistent clipboard for files, folders, and text snippets with apply, pop, and clear operations, almost similar to `git` stash |
| **File System Synchronization** | One-shot bidirectional file synchronization between two machines over HTTP/HTTPS with decoupled send/receive and listen/connect roles |
| **Neo4j Database Interaction** | Execute Cypher queries against Neo4j databases from command line or YAML files |
| **Markdown Viewer** | Start a web server to view rendered markdown files with syntax highlighting, navigation, and Mermaid support |
| **AWS Helper Utilities** | Configure AWS SSO with IAM Identity Center for multi-role access and generate console URLs from CLI profiles |
| **Azure Helper Utilities** | Switch between Azure subscriptions interactively |

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
  anbu pass add API_KEY     # Create a new secret (encrypted with AES GCM at rest)
  anbu pass add API_KEY -m  # Create a new multi-line secret
  anbu pass get API_KEY     # Retrieve a secret (decrypted value)
  anbu pass delete API_KEY  # Delete a secret

  # Import and Export to file
  anbu pass export backup.json  # Export to a file (secrets are decrypted)
  anbu pass import backup.json  # Import from a file
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
  anbu http-server                     # Serves current directory on http://0.0.0.0:8080
  anbu http-server -l 0.0.0.0:8080 -t  # Serve HTTPS on given add:port with a self-signed cert
  anbu http-server -u                  # Serve simple upload page for text and files
  anbu http-server -u -t               # Serve upload page over HTTPS with self-signed cert
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
  anbu rename 'prefix_(.*)' 'new_\1'              # Rename files matching regex pattern
  anbu rename -d 'old_(.*)' 'new_\1'              # Rename directories instead of files
  anbu rename '(.*)\.(.*)' '\1_backup.\2'         # Add _backup before extension
  anbu rename 'image-(\d+).jpg' 'IMG_\1.jpeg' -r  # Perform a dry-run without renaming
  anbu rename '(.*)' '\1_\uuid'                    # Append UUID to filenames
  anbu rename '(.*)\.(.*)' '\1_\suid.\2'           # Insert short UUID before extension
  ```

- ***Manual Rename*** (alias: `mrename`)

  ```bash
  anbu manual-rename                    # Interactively rename files one by one
  anbu mrename -d                       # Include directories in the rename operation
  anbu mrename -H                       # Include hidden files and directories
  anbu mrename -x                       # Allow changing file extension
  anbu mrename -d -H -x                 # Include directories, hidden files, and allow extension changes
  ```

- ***Find Duplicates*** (alias: `dup`)

  ```bash
  anbu duplicates                 # Find duplicate files in the current directory
  anbu dup --recursive            # Find duplicate files recursively in subdirectories
  anbu dup --delete               # Find and delete duplicate files
  ```

- ***Sed (Regex Substitution)***

  ```bash
  anbu sed 'old' 'new' file.txt                            # Replace all occurrences of 'old' with 'new' in a file
  anbu sed '([a-z]+)@([a-z]+)\.com' '\1@***.com' file.txt  # Replace email patterns with masked version
  anbu sed 'foo' 'bar' ./directory                         # Apply substitution to all files in directory
  anbu sed 'foo' 'bar' ./directory -r                      # Perform a dry-run without modifying files
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

  # Docker command conversion
  anbu convert docker-compose "docker run -p 8080:80 nginx"  # Convert docker run command to docker-compose.yml
  anbu convert compose-docker docker-compose.yml             # Convert docker-compose.yml to docker run command
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

- ***Stash***

  ```bash
  # Stash a file or folder (keeps original, unlike git stash)
  anbu stash fs /path/to/file.txt
  anbu stash fs ./my-folder

  # Stash text from stdin
  echo "my text" | anbu stash text my-snippet
  anbu stash text notes <<EOF
  Multi-line text
  goes here
  EOF

  # List all stashed entries
  anbu stash list

  # Apply a stash without removing it (text prints to stdout, files/folders extracted to current directory)
  anbu stash apply 1

  # Apply a stash and remove it (pop operation)
  anbu stash pop 1

  # Remove a stash without applying it
  anbu stash clear 1
  ```

- ***File System Synchronization***

  Send and receive roles are decoupled from listen and connect roles, so either side can be the sender or receiver regardless of which side listens.

  ```bash
  # Sender listens, receiver connects (sender has open port)
  anbu fs-sync send --listen -p 8080 -d /path/to/sync/dir --ignore ".git,node_modules"
  anbu fs-sync receive --connect http://sender.example.com:8080 -d /path/to/local/dir

  # Receiver listens, sender connects (receiver has open port)
  anbu fs-sync receive --listen -p 8080 -d /path/to/local/dir
  anbu fs-sync send --connect http://receiver.example.com:8080 -d /path/to/sync/dir --ignore ".git,node_modules"

  # With TLS (listener enables TLS, connector skips verification for self-signed certs)
  anbu fs-sync send --listen -p 8443 -d ./sync-dir -t
  anbu fs-sync receive --connect https://sender.com:8443 -d ./local-dir -k

  # Dry run and delete options (receiver-side flags)
  anbu fs-sync receive --connect http://sender.com:8080 -d ./local-dir -r       # Dry run
  anbu fs-sync receive --connect http://sender.com:8080 -d ./local-dir --delete  # Delete extra local files
  anbu fs-sync receive --listen -p 8080 -d ./local-dir --delete --dry-run        # Dry run with delete preview
  ```

- ***Neo4j Database Interaction***

  ```bash
  # Execute a single Cypher query
  anbu neo4j -q "MATCH (n) RETURN n LIMIT 5"
  anbu neo4j -r neo4j://localhost:7687 -u neo4j -p password -d neo4j -q "MATCH (n) RETURN count(n)"

  # Execute queries from a YAML file (multi-line queries supported using '|' in YAML)
  anbu neo4j --query-file ./queries.yaml --output-file results.json

  # Execute write queries (CREATE, UPDATE, DELETE, etc.)
  anbu neo4j --write -q "CREATE (n:Person {name: 'Alice'}) RETURN n"

  # Custom connection settings
  anbu neo4j -r neo4j+s://example.com:7687 -u admin -p secret -d mydb -q "MATCH (n) RETURN n" -o output.json
  ```

- ***Markdown Viewer*** (alias: `md`)

  ```bash
  anbu markdown          # Start markdown viewer on default address (0.0.0.0:8080)
  anbu md -l :3000       # Start on port 3000 on all interfaces
  ```

- ***AWS Helper Utilities***

  ```bash
  # Configure AWS SSO with IAM Identity Center for multi-role access
  # This will create profiles in ~/.aws/config for all accounts and roles
  anbu aws iidc-login -u https://my-sso.awsapps.com/start -r us-east-1
  anbu aws iidc-login --start-url https://my-sso.awsapps.com/start --sso-region us-east-1 --cli-region us-west-2 --session-name my-sso

  # Generate AWS console URL from a local CLI profile (valid for up to 12 hours)
  anbu aws cli-ui -p my-profile
  ```

- ***Azure Helper Utilities*** (alias: `az`)

  ```bash
  # Switch between Azure subscriptions interactively
  anbu azure switch-sub
  anbu az switch
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

