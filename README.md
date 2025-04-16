<div align="center">
  <img src=".github/assets/logo.png" alt="ANBU Logo" width="250"/>

  <h1 align="center">Anbu</h1>

  <a href="https://github.com/tanq16/anbu/actions/workflows/release.yml"><img src="https://github.com/tanq16/anbu/actions/workflows/release.yml/badge.svg" alt="Release Build"></a>&nbsp;<a href="https://github.com/tanq16/anbu/releases/latest"><img src="https://img.shields.io/github/v/release/tanq16/anbu" alt="Latest Release"></a><br>

  <p><b>Anbu</b> is a CLI tool that helps perform everyday tasks in an expert way. Just like the Anbu Black Ops division in Naruto, this tool helps carry out all the shadow-operations in your daily workflow.</p><br>
  
  <a href="#installation">Installation</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#acknowledgements">Acknowledgements</a><br>
</div>

## Installation

Download directly from [releases](https://github.com/Tanq16/anbu/releases). Anbu is available for AMD64 and ARM64 for Linux and MacOS only. Determine the version with `anbu -v`.

To build latest commit directly via Go, use:

```bash
go install github.com/tanq16/anbu@latest
```

To clone and build locally for development, use:

```bash
git clone https://github.com/tanq16/anbu.git && cd anbu
go build .
```

## Usage

```
anbu is a tool for performing various everyday tasks with ease

Usage:
  anbu [command]

Available Commands:
  filecrypt   Encryption/decryption on files using AES-256-GCM symmetric encryption
  httpserver  Start a simple HTTP/HTTPS file server
  key-pair    Generate RSA key pairs for encryption
  loop        execute a command for each number range in a range
  string      generate a random string, a sequence, a repetition, or password/passphrase
  time        time related function: use `now`, `purple`, `diff` (calculate epoch diff), `parse` (ingest a time str & print)
  tunnel      Create tunnels between local and remote endpoints

Flags:
      --debug     enable debug logging
  -h, --help      help for anbu
  -v, --version   version for anbu

Use "anbu [command] --help" for more information about a command.
```

Anbu supports a large number of operations across the board. The specific details of each are:

- ***File Encryption/Decryption***
  - ```bash
    anbu filecrypt encrypt /path/to/file.zip -p "P@55w0rd"  # Encrypt a file
    anbu filecrypt decrypt ./encrypted.enc -p "P@55w0rd"    # Decrypt a file
    ```
- ***RSA Key Pair Generation***
  - ```bash
    anbu key-pair -o mykey -k 4096  # 4096 bit RSA key pair
    ```
  - ```bash
    anbu key-pair --ssh  # 2048 bit RSA SSH key pair called anbu-key.*
    ```
- ***Loop Command***
  - ```bash
    anbu loop 03-112 'echo "$i"' -p 2  # run command for index 3 to 112 as 003, 004, ...
    ```
  - ```bash
    anbu loop 20 'echo justprintme'  # run command 20 times linearly
    ```
- ***String Generation***
  - ```bash
    anbu string 23  # generate 23 (100 if not specified) random alphanumeric chars
    ```
  - ```bash
    anbu string seq 29  # prints "abcdefghijklmnopqrstuvxyz" back to back until desired length
    ```
  - ```bash
    anbu string rep 23 stringToRepeat  # prints "stringToRepeatstringToRepeat...23 times"
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
- ***Time Operations***
  - ```bash
    anbu time  # prints time in various formats
    ```
  - ```bash
    anbu time -a purple  # print time and public IP for purple teams
    ```
  - ```bash
    anbu time -a diff -e 1744192475 -e 1744497775  # print human readable diff between 2 epochs
    ```
  - ```bash
    anbu time -a parse -t "13 Apr 25 16:30 EDT"
    # read time in a format and print equivalent in multiple formats
    ```
- ***Network Tunneling***
  - ```bash
    anbu tunnel tcp -l localhost:8000 -r example.com:80
    # TCP Tunnel: forward local port 8000 to example.com:80
    ```
  - ```bash
    anbu tunnel tcp -l localhost:8000 -r example.com:443 --tls
    # TLS TCP Tunnel: forwards local port 8000 to example.com:443 using TLS
    ```
  - ```bash
    anbu tunnel ssh -l localhost:8000 -r internal.example.com:3306 -s ssh.example.com:22 -u username -p password
    # SSH Tunnel: establishes SSH tunnel from localhost:8000 to internal.example.com:3306 via SSH server
    ```
  - ```bash
    anbu tunnel ssh -l localhost:8000 -r internal.example.com:3306 -s ssh.example.com:22 -u username -k ~/.ssh/mykey
    # SSH Tunnel with Key Authentication: uses SSH key authentication instead of password
    ```
- ***Simple HTTP/HTTPS Server***
  - ```bash
    anbu httpserver  # Serves current directory on http://localhost:8000
    ```
  - ```bash
    anbu httpserver -l 0.0.0.0:8080 -t  # Serve HTTPS on given add:port with a self-signed cert
    ```
  - ```bash
    anbu httpserver -u  # Enables file upload via PUT requests
    ```

## Acknowledgements

Anbu takes inspiration from the following projects:

- [GoST](https://github.com/ginuerzh/gost)
- [SimpleHTTPServer](https://github.com/projectdiscovery/simplehttpserver)
