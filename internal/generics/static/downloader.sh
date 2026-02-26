#!/bin/bash
set -e

mkdir -p css js

curl -L -o js/marked.min.js "https://cdn.jsdelivr.net/npm/marked/marked.min.js"
echo "Downloaded js/marked.min.js"

curl -L -o js/mermaid.min.js "https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"
echo "Downloaded js/mermaid.min.js"

curl -L -o js/highlight.min.js "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/highlight.min.js"
echo "Downloaded js/highlight.min.js"

curl -L -o css/github-dark.min.css "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.7.0/styles/github-dark.min.css"
echo "Downloaded css/github-dark.min.css"

curl -L -o js/tailwindcss.js "https://cdn.tailwindcss.com/3.4.16"
echo "Downloaded js/tailwindcss.js"
