<div align="center">
  <img src=".github/assets/logo.png" alt="ANBU Logo" width="250"/>

  <h1 align="center">Anbu</h1>

  <a href="https://github.com/tanq16/anbu/actions/workflows/release.yml"><img src="https://github.com/tanq16/anbu/actions/workflows/release.yml/badge.svg" alt="Release Build"></a>&nbsp;<a href="https://github.com/tanq16/anbu/releases/latest"><img src="https://img.shields.io/github/v/release/tanq16/anbu" alt="Latest Release"></a><br>

  <p><b>Anbu</b> is a CLI tool that helps perform everyday tasks in an expert way. Just like the Anbu Black Ops division in Naruto, this tool helps carry out all the shadow-operations in your daily workflow.</p><br>
  
  <a href="#installation">Installation</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#acknowledgements">Acknowledgements</a> &bull; <a href="#tips--tricks">Tips & Tricks</a><br>
</div>

A summary of all capabilities that **Anbu** can perform:

- Time Operations
- Network Tunneling
- Command Template Execution
- Simple HTTP/HTTPS Server
- JWT Decode
- Secrets Scan
- IP Information
- Bulk Rename
- Data Conversion
- Encoding Conversion
- File Encryption/Decryption
- RSA Key Pair Generation
- Loop Command
- String Generation

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

- ***Time Operations***
  - ```bash
    anbu time          # prints time in various formats
    anbu time now      # prints time in various formats
    anbu time purple   # print time and public IP for purple teams
    ```
  - ```bash
    anbu time diff -e 1744192475 -e 1744497775  # print time difference between 2 epochs
    ```
  - ```bash
    anbu time parse -t "13 Apr 25 16:30 EDT"  # read given time and print in multiple formats
    anbu time until -t "13 Apr 25 16:30 EDT"  # read time and print difference from now
    ```
- ***Network Tunneling***
  - ```bash
    # forward TCP tunnels
    anbu tunnel tcp -l localhost:8000 -r example.com:80  # also supports --tls
    ```
  - ```bash
    # forward SSH tunnels
    anbu tunnel ssh -l localhost:8000 -r target.com:3306 -s ssh.vm.com:22 -u bob -p "builder"
    anbu tunnel ssh -l localhost:8000 -r target.com:3306 -s ssh.vm.com:22 -u bob -k ~/.ssh/mykey
    ```
  - ```bash
    # reverse SSH tunnels
    anbu tunnel rssh -l localhost:3389 -r 0.0.0.0:8080 -s ssh.vm.com:22 -u bob -p "builder"
    ```
- ***Command Template Execution***
  - ```bash
    anbu exec ./path/to/template.yaml  # Execute template file with commands as steps
    ```
  - ```bash
    anbu exec template.yam -v 'pass=P@55w0rd' -v 'uname=4.u53r'
    # Execute template file with custom variable replacement (see Tips for more information)
    ```
- ***Simple HTTP/HTTPS Server***
  - ```bash
    anbu http-server                     # Serves current directory on http://localhost:8000
    anbu http-server -l 0.0.0.0:8080 -t  # Serve HTTPS on given add:port with a self-signed cert
    anbu http-server -u                  # Enables file upload via PUT requests
    ```
- ***JWT Decode***
  - ```bash
    anbu jwt-decode "$TOKEN"  # Decodes and prints the headers and payload values in a table
    ```
- ***Secrets Scan***
  - ```bash
    anbu secrets                 # Scans current directory for secrets based on regex matches
    anbu secrets ./path/to/scan  # Scans path for secrets based on regex matches
    anbu secrets ./path -p       # Scans path with generic matches table (maybe false positive)
    ```
- ***IP Information***
  - ```bash
    anbu ipinfo       # Print local and public IP information
    anbu ipinfo ipv6  # Print local (IPv4 & IPv6) and public IP information
    ```
- ***Bulk Rename***
  - ```bash
    anbu rename 'prefix_(.*)' 'new_\1'        # Rename files matching regex pattern
    anbu rename -d 'old_(.*)' 'new_\1'        # Rename directories instead of files
    anbu rename '(.*)\\.(.*)' '\1_backup.\2'  # Add _backup before extension
    ```
- ***Data Conversion***
  - ```bash
    anbu convert yaml-json config.yaml  # Convert YAML file to JSON
    anbu convert json-yaml data.json    # Convert JSON file to YAML
    ```
- ***Encoding Conversion***
  - ```bash
    anbu convert b64 "Hello World"              # Convert text to base64
    anbu convert b64d "SGVsbG8gV29ybGQ="        # Decode base64 to text
    anbu convert hex "Hello World"              # Convert text to hex
    anbu convert hexd "48656c6c6f20576f726c64"  # Decode hex to text
    anbu convert url "Hello World"              # URL encode text
    anbu convert urld "Hello%20World"           # URL decode text
    ```
  - ```bash
    anbu convert b64-hex "SGVsbG8gV29ybGQ="        # Convert base64 to hex
    anbu convert hex-b64 "48656c6c6f20576f726c64"  # Convert hex to base64
    ```
- ***File Encryption/Decryption***
  - ```bash
    anbu file-crypt encrypt /path/to/file.zip -p "P@55w0rd"  # Encrypt a file
    anbu file-crypt decrypt ./encrypted.enc -p "P@55w0rd"    # Decrypt a file
    ```
- ***RSA Key Pair Generation***
  - ```bash
    anbu key-pair -o mykey -k 4096  # 4096 bit RSA key pair
    anbu key-pair --ssh             # 2048 bit RSA SSH key pair called anbu-key.*
    ```
- ***Loop Command***
  - ```bash
    anbu loop 03-112 'echo "$i"' -p 2  # run command for index 3 to 112 as 003, 004, ...
    anbu loop 20 'echo justprintme'    # run command 20 times linearly
    ```
- ***String Generation***
  - ```bash
    anbu string 23               # generate 23 (100 if not specified) random alphanumeric chars
    anbu string seq 29           # prints "abcdefghijklmnopqrstuvxyz" until desired length
    anbu string rep 23 str2rep   # prints "str2repstr2rep...23 times"
    ```
  - ```bash
    anbu string uuid     # generates a uuid
    anbu string ruid 16  # generates a short uuid of length b/w 1-32
    anbu string suid     # generates a short uuid of length 18
    ```
  - ```bash
    anbu string password           # generate a 12-character complex password
    anbu string password 16        # generate a 16-character complex password
    anbu string password 8 simple  # generate an 8-letter lowercase password
    ```
  - ```bash
    anbu string passphrase               # generate a 3-word passphrase with hyphens
    anbu string passphrase 5             # generate a 5-word passphrase with hyphens
    anbu string passphrase 4 '@'         # generate a 4-word passphrase with period separators
    anbu string passphrase 4 '-' simple  # generate a simple 4-word lowercase passphrase
    anbu string passphrase 4 '.' simple  # generate a simple 4-word passphrase with numbers and capitalization
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
<summary><b>Running a Command Template</b></summary>

A command template needs to be in the form of a YAML template. Variables can be declared inline as well as within the template. A variable `var` for example, should be used as `{{.var}}` in the command string.

An example of a template is as follows:

```yaml
name: "Project Backup"
description: "Creates a timestamped backup of a project directory"
variables:
  backup_dir: "/home/tanq/testrepo"
  exclude_patterns: ".git,*.log"

steps:
  - name: "Clone project"
    command: "git clone https://github.com/tanq16/danzo {{.backup_dir}}"

  - name: "Build Project"
    command: "cd {{.backup_dir}} && go build -ldflags='-s -w' ."

  - name: "Get current timestamp"
    command: "timestamp=$(date +%Y%m%d_%H%M%S) && echo $timestamp > {{.backup_dir}}/anbu_backup_timestamp.txt && echo '{{.exclude_patterns}}' > {{.backup_dir}}/anbu_exclude_patterns.txt"

  - name: "Create backup archive"
    command: "cd {{.project_dir}} && tar --exclude-from={{.backup_dir}}/anbu_exclude_patterns.txt -czf backup_$(cat {{.backup_dir}}/anbu_backup_timestamp.txt).tar.gz {{.backup_dir}}"

  - name: "List created backup"
    command: "ls -lh {{.project_dir}}/*.tar.gz"

  - name: "Cleanup temporary files"
    command: "rm /tmp/anbu_backup_timestamp.txt /tmp/anbu_exclude_patterns.txt" # used non-existent for demo
    ignore_errors: true

```

The template can then be executed as:

```bash
anbu exec template.yaml -v 'project_dir=/opt/backups'
```

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

## Acknowledgements

Anbu takes inspiration from the following projects:

- [GoST](https://github.com/ginuerzh/gost)
- [SimpleHTTPServer](https://github.com/projectdiscovery/simplehttpserver)
- [TruffleHog](https://github.com/trufflesecurity/trufflehog), [GitLeaks](https://github.com/gitleaks/gitleaks), and [NoseyParker](https://github.com/praetorian-inc/noseyparker) for secret regular expressions
