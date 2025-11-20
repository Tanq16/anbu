#!/bin/bash
set -e

curl -L -o marked.min.js "https://cdn.jsdelivr.net/npm/marked/marked.min.js"
echo "Downloaded marked.min.js"

curl -L -o mermaid.min.js "https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"
echo "Downloaded mermaid.min.js"

curl -L -o highlight.min.js "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/highlight.min.js"
echo "Downloaded highlight.min.js"

curl -L -o github-dark.min.css "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/styles/github-dark.min.css"
echo "Downloaded github-dark.min.css"
