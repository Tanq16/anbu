<div align="center">
  <img src=".github/assets/logo.png" alt="ANBU Logo" width="250"/>

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
| **Simple HTTP/HTTPS Server** | Host a simple webserver over HTTP/HTTPS with optional file upload capability |
| **Secrets Scan** | Find common secrets in file systems using regular expressions |
| **IP Information** | Display local and public IP details, including geolocation information |
| **Bulk Rename** | Batch rename files or directories using regular expression patterns |
| **Data & Encoding Conversion** | Convert between data formats (YAML/JSON), decode JWTs, and handle various encodings (Base64, Hex, URL) |
| **File Encryption/Decryption** | Secure file encryption and decryption with AES-256-GCM symmetric encryption |
| **RSA Key Pair Generation** | Create RSA key pairs for encryption or SSH authentication |
| **String Generation** | Generate random strings, UUIDs, passwords, and passphrases for various purposes |
| **Stash** | Persistent clipboard for files, folders, and text snippets with apply, pop, and clear operations, almost similar to `git` stash |
| **Google Drive Interaction** | Interact with Google Drive to list, upload, download, and sync files and folders |
| **Box.com Interaction** | Interact with Box.com to list, upload, download, and sync files and folders |
| **GitHub Interaction** | Interact with GitHub to list issues, PRs, workflow runs, add comments, create issues/PRs, and download files/folders |
| **File System Synchronization** | Synchronize files between client and server using WebSocket with real-time change propagation |
| **Neo4j Database Interaction** | Execute Cypher queries against Neo4j databases from command line or YAML files |
| **Markdown Viewer** | Start a web server to view rendered markdown files with syntax highlighting, navigation, and Mermaid support |

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

  # Managing Secrets (Password asked or from ANBUPW env var)
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
  anbu http-server                     # Serves current directory on http://localhost:8000
  anbu http-server -l 0.0.0.0:8080 -t  # Serve HTTPS on given add:port with a self-signed cert
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

- ***Google Drive Interaction*** (alias: `gd`)

  ```bash
  # Set up credentials by placing credentials.json at ~/.anbu-gdrive-credentials.json or pass with --credentials flag
  anbu gdrive -c path/to/credentials.json list

  # List files and folders (defaults to root 'My Drive')
  anbu gdrive list
  anbu gd ls "My Folder"
  anbu gd ls "My Folder/file.txt"  # Shows info for a specific file

  # Upload a file or folder (defaults uploading to root 'My Drive' when not specified)
  anbu gdrive upload local-file.txt "My Folder"
  anbu gd up local-folder  # uploads folder recursively to root

  # Download a file or folder to the current working directory
  anbu gdrive download "My Drive Folder/remote-file.txt"
  anbu gd dl "My Drive Folder/remote-folder"  # downloads folder recursively

  # Sync local directory with remote directory (uploads missing, deletes remote-only, updates changed)
  anbu gdrive sync ./local-dir "My Drive Folder"
  anbu gd sync ./local-dir "Backup"  # syncs using MD5 hashes (yes, Google uses this) for comparison
  ```

- ***Box Interaction***

  ```bash
  # Set up credentials by placing credentials.json at ~/.anbu-box-credentials.json or pass with --credentials flag
  anbu box -c path/to/credentials.json list

  # List files and folders (defaults to root folder)
  anbu box list
  anbu box ls "My Folder"
  anbu box ls "My Folder/file.txt"  # Shows info for a specific file

  # Upload a file or folder (defaults uploading to root when not specified)
  anbu box upload local-file.txt "My Folder"
  anbu box up local-folder  # uploads folder recursively to root

  # Download a file or folder to the current working directory
  anbu box download "My Folder/remote-file.txt"
  anbu box dl "My Folder/remote-folder"  # downloads folder recursively

  # Sync local directory with remote directory (uploads missing, deletes remote-only, updates changed)
  anbu box sync ./local-dir "Test Folder"
  anbu box sync ./local-dir ""  # syncs to root folder using SHA1 hashes (yes, Box uses this) for comparison
  ```

- ***GitHub Interaction*** (alias: `gh`)

  ```bash
  # Put OAuth app client ID at ~/.anbu-github-credentials.json or pass another file with --credentials flag
  anbu github -c path/to/credentials.json list owner/repo/i

  anbu gh ls owner/repo/i # List issues
  anbu gh ls owner/repo/i/23 # List comments for an issue
  anbu gh ls owner/repo/pr # List pull requests
  anbu gh ls owner/repo/pr/24 # List comments for a PR
  anbu gh ls owner/repo/a # List workflows
  anbu gh ls owner/repo/a/3 # List jobs in a workflow run (3rd run)
  
  # Get info about a job (4th job in 3rd workflow run)
  anbu gh ls owner/repo/a/3/4

  # Download logs for a job to anbu-github.log
  anbu gh ls owner/repo/a/3/4/logs

  # Add comment to an issue or pr (end with 'EOF' on a new line)
  anbu gh add owner/repo/i/23
  anbu gh add owner/repo/pr/24

  anbu gh make owner/repo/i # Create a new issue
  anbu gh make owner/repo/pr/newfeat # Create a PR from branch to main
  anbu gh make owner/repo/pr/newfeat/develop # Create a PR from branch to base branch

  # Download files or folders from a repository
  anbu gh download owner/repo/tree/main/src/file.go      # Download a single file
  anbu gh dl owner/repo/tree/feature/src                 # Download folder from feature branch
  anbu gh download owner/repo/tree/abc123def/path/to/dir # Download from specific commit
  ```

- ***File System Synchronization***

  ```bash
  # Run the fs-sync server (maintains source of truth)
  anbu fs-sync server -p 8080 -d /path/to/sync/dir
  anbu fs-sync server --port 8080 --dir ./sync-dir --ignore ".git/*,*.tmp"

  # Run the fs-sync client (connects to server and syncs)
  anbu fs-sync client -a ws://server.example.com:8080/ws -d /path/to/local/dir
  anbu fs-sync client --addr wss://file.sync.com/ws --dir ./local-dir --ignore "node_modules/*,*.log"
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

<details>
<summary><b>Creating Google Drive API Credentials</b></summary>

To use the `gdrive` command, you need to create OAuth 2.0 credentials. Here’s how to do it:

1. **Go to the Google Cloud Console:** Navigate to [https://console.cloud.google.com/](https://console.cloud.google.com/) and create a new project (or select an existing one).
2. **Enable the Google Drive API:** In the project dashboard, search for "Google Drive API" and enable it.
3. **Create OAuth Credentials:**
    - Go to **APIs & Services > Credentials**.
    - Click **Create Credentials > OAuth client ID**.
    - Select **Desktop app** as the application type.
    - Give it a name and click **Create**.
4. **Download Credentials:**
    - After creation, a dialog will show your client ID and secret. Click **Download JSON** to save the `credentials.json` file.
    - Place this file at `~/.anbu-gdrive-credentials.json` or provide its path using the `--credentials` flag.
5. **Publish the App:**
    - In the **OAuth consent screen** tab, you need to configure the consent screen. For personal use, you can keep it in "Testing" mode and add your Google account as a test user.
    - If you want to allow other users, you must publish the app, which may require verification by Google.

**Authentication Flow:**

When you run a `gdrive` command for the first time, `anbu` will:
1. Print a URL to your console.
2. Open this URL in your web browser. You will be prompted to sign in with your Google account.
3. Since the app is not verified by Google, you will see a warning screen. Click **Advanced** and then **Go to (unsafe)** to proceed.
4. After you grant permission, the browser will redirect to a `localhost` address, which will likely fail to load. This is expected.
5. Copy the authorization code from the URL in your browser’s address bar (it will be a long string in the `code` parameter).
6. Paste this code back into the terminal where `anbu` is waiting.

`anbu` will then use this code to get an access token and refresh token, which it will store for future use.

</details>

<details>
<summary><b>Creating Box API Credentials</b></summary>

To use the `box` command, you need to create OAuth 2.0 credentials. Here's how to do it:

1. **Go to the Box Developer Console:** Navigate to [Developer Console](https://app.box.com/developers/console) and sign in with your Box account.
2. **Create a New App:**
    - Click **Create Platform App** and then **Custom App**,
    - Name it and choose **User Authentication (OAuth 2.0)**.
3. **Configure OAuth Settings:**
    - Go to the **Configuration** tab of the new app.
    - Under **OAuth 2.0 Credentials**, copy and store your **Client ID** and **Client Secret**.
    - Add `http://localhost:8080` as a **Redirect URI**.
    - Enable the option for **Write all files and folders stored in Box**.
4. **Create Credentials File:**
    - Create a JSON file with the following structure:
      ```json
      {
        "client_id": "your_client_id",
        "client_secret": "your_client_secret"
      }
      ```
    - Save this file as `~/.anbu-box-credentials.json` or provide its path using the `--credentials` flag.

**Authentication Flow:**

When you run a `box` command for the first time, `anbu` will:
1. Print a URL to your console. Open this in your web browser. You will be prompted to sign in with Box.
2. After granting permission, the browser will redirect to a `localhost` address, which will fail to load. This is expected.
3. Copy the full redirect URL from the address bar (it looks like `http://localhost:8080/?code=...&state=...`).
4. Paste the URL back into the terminal where `anbu` is waiting.

`anbu` will then use this URL to extract the authorization code and exchange it for an access token and refresh token, which will be stored for future use. Future commands will not require the credentials flag.

</details>

<details>
<summary><b>Creating GitHub API Credentials</b></summary>

To use the `github` command, you need to create a GitHub OAuth App. Here's how to do it:

1. Navigate to [GitHub Developer Settings](https://github.com/settings/developers).
2. **Create a New OAuth App:**
   - Click **OAuth Apps** in the left sidebar, then click **New OAuth App**.
   - Fill in the application details:
     - **Application name:** Anbu
     - **Homepage URL:** `http://localhost`
     - **Authorization callback URL:** `http://localhost:8080` (not used for device flow, but required)
   - Click **Register application**.
3. Get Client ID for the app and copy the value. Client Secret is not required for device code login.
4. **Create Credentials File:**
   - Put the client ID in as follows:
     ```json
     {
       "client_id": "your_client_id"
     }
     ```
   - Save this file as `~/.anbu-github-credentials.json` or provide its path using the `--credentials` flag.

**Authentication Flow:**

When you run a `github` command for the first time, `anbu` will:
1. Print a verification URL and a user code to your console.
2. Open the URL in your web browser and enter the user code when prompted.
3. Authorize the application in your browser, then return to the terminal and press Enter.
4. `anbu` will then check authorization and retrieve an access token.

The access token is saved at `~/.anbu-github-token.json` and will be reused for subsequent commands.

</details>

<details>
<summary><b>Path Shortcuts for Box and Google Drive</b></summary>

You can define path shortcuts to simplify remote paths. For Box, create `~/.anbu-box-shortcuts.json`, and for Google Drive, create `~/.anbu-gdrive-shortcuts.json`. Use `%shortcut%` syntax in remote paths (not local paths for Google Drive) to automatically expand shortcuts. Use `%%` for a literal percent sign.

```json
{
  "project": "MyProject/2024/Documents",
  "reports": "Shared/Reports"
}
```

Example: `anbu box download %project%/file.pdf` or `anbu gdrive upload local.txt %reports%` will expand to the full paths.

</details>
