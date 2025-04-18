<div align="center">
  <img src=".github/assets/logo.png" alt="ANBU Logo" width="250"/>

  <h1 align="center">Anbu</h1>

  <a href="https://github.com/tanq16/anbu/actions/workflows/release.yml"><img src="https://github.com/tanq16/anbu/actions/workflows/release.yml/badge.svg" alt="Release Build"></a>&nbsp;<a href="https://github.com/tanq16/anbu/releases/latest"><img src="https://img.shields.io/github/v/release/tanq16/anbu" alt="Latest Release"></a><br>

  <p><b>Anbu</b> is a CLI tool that helps perform everyday tasks in an expert way. Just like the Anbu Black Ops division in Naruto, this tool helps carry out all the shadow-operations in your daily workflow.</p><br>
  
  <a href="#installation">Installation</a> &bull; <a href="#usage">Usage</a> &bull; <a href="#acknowledgements">Acknowledgements</a> &bull; <a href="#tips--tricks">Tips & Tricks</a><br>
</div>

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
    anbu tunnel tcp -l localhost:8000 -r example.com:80  # forward TCP tunnel, also supports --tls
    ```
  - ```bash
    anbu tunnel rtcp -l localhost:3000 -r public-server.com:8000  # reverse TCP Tunnel (for NAT traversal), also supports --tls
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

***Connecting Two NAT-hidden Machines via Public VPS***

*Machine A* &rarr;
```bash
anbu tunnel rssh -l localhost:3389 -r 0.0.0.0:8001 -s vps.example.com:22 -u bob -p builder
```

*Machine B* &rarr;
```bash
anbu tunnel ssh -l localhost:3389 -r localhost:8001 -s vps.example.com:22 -u bob -p builder
```

Now, connecting to `localhost:3389` on Machine B will allow access to Machine A's 3389.

## Acknowledgements

Anbu takes inspiration from the following projects:

- [GoST](https://github.com/ginuerzh/gost)
- [SimpleHTTPServer](https://github.com/projectdiscovery/simplehttpserver)
