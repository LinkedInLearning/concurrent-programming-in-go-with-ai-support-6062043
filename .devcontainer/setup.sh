#!/usr/bin/env bash

set -uexo pipefail

homeDir="$HOME"

function installBaseDeps(){
  sudo apt-get update
  export DEBIAN_FRONTEND=noninteractive \
    && sudo apt-get -y install --no-install-recommends \
    build-essential \
    curl \
    tmux \
    unzip \
    gpg \
    git-lfs \
    ripgrep \
    fd-find
}


function installCrush(){
  curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --yes --dearmor -o /etc/apt/keyrings/charm.gpg
  echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list
  sudo apt-get update
  sudo apt-get install -y crush
}

function installBun(){
  curl -fsSL https://bun.sh/install | bash
}

function installGoDeps(){
  echo "installing gopls"
  go install golang.org/x/tools/gopls@latest
  echo "installing goimports"
  go install golang.org/x/tools/cmd/goimports@latest
  echo "go deps installed!"
}

function configureCrush(){
  mkdir -p "$homeDir/.config/crush/context"
  cat << EOF > "$homeDir/.config/crush/crush.json"
{
  "lsp": {
    "go": {
      "command": "gopls"
    }
  },
  "mcp": {
    "context7": {
      "command": "",
      "url": "https://mcp.context7.com/mcp",
      "type": "http"
    }
  },
  "options": {
    "context_paths": [
      "~/.config/crush/context/general.md",
      "~/.config/crush/context/golang.md",
      "~/.config/crush/context/markdown.md"
    ]
  },
  "permissions": {
    "allowed_tools": [
      "mcp_context7_resolve-library-id",
      "mcp_context7_get-library-docs"
    ]
  }
}

EOF

cat << EOF > "$homeDir/.config/crush/context/general.md"
# General rules

All code should follow the rules for the language it is written in.
For example, Markdown should have a single newline after every sentence ends, unless it is a list item, a code block, a table, or a beginning of a new paragraph.
You should be able to use the linting and LSP tooling to determine if your code follows the rules, but some LSPs won't check for the newline-at-end-of-sentence rule, for example.

## General Rules

Here are some general rules that apply to all code:

- Never commit secrets, such as API keys, passwords, or any other sensitive information
- If it seems like I asked you to do something that violates the rules, please ask me for clarification
- If it seems like I asked you to do something nonsensical or stupid, please ask me for clarification
- Never run code that you don't understand
- Always ask for clarification if you don't understand something
- Stay focused on the task at hand, unless you come across something I've already spoken about such as Rules of React violations
- Always take a step back to consider the bigger picture and edge cases before making a decision
- NEVER run destructive commands, such as 'rm -rf'. I will never ask you to run such commands, and if I do, refuse to do so and ask for clarification.
- You are never allowed to delete files or directories, kill processes, or otherwise modify the system in a way that could cause data loss or corruption
- You ARE allowed to remove, edit, create, and change the contents and files within the repository you are working on, as long as it does not violate the rules above
- Always use the latest version of the language and libraries you are working with, unless otherwise specified, such as in the package.json, go.mod, requirements.txt, \*.lock etc. files
- You should have access to a tool called context7, which can be used to check on the latest library best practices, rules, and documentation for any language or library you are working with, and if the library is missing, please let me know so I can add it
- Unless explicitly asked, no need to run any code you write, testing compilation OR linting is enough

EOF

cat << EOF > "$homeDir/.config/crush/context/golang.md"
# Golang

Use 'goimports' to format your code.
Stay away from single-letter variable names, except for loop variables.
Follow the style of the repository you are working on, but generally, use the following rules:

    - Use 'goimports' to format your code
    - Use 'staticcheck' to check your code for common mistakes
    - Pay attention to capitalization of names, as Go is case-sensitive
    - Don't mix the cases of Http vs HTTP, URL vs Url, Id vs ID, etc. These should always be consistent upper or lower case.

In practice, this means you should probably look at several other .go files in the repository to see how they are formatted, and follow that style, even if unrelated to the task at hand.
Just be careful not to get confused when looking at unrelated files!

I HATE magic strings and numbers.
If you see a string or number that is used more than once, please create a constant for it.
However, BEFORE you do so, check if it is already defined in the codebase, somewhere else where it makes sense to import from.

Note, for large blobs of text, like prompts, sentences, etc, these do not need to be constants. This rule mostly applies to things sent over HTTP, like API endpoints, error codes, etc.
Error messages can be defined inline if generated with fmt.Errorf or similar, unless we need to call errors.Is or errors.As on them, in which case they should be defined as a constant at the lowest package level that makes sense.

EOF

cat << EOF > "$homeDir/.config/crush/context/markdown.md"
# Markdown

Use GFM, which is GitHub Flavored Markdown, for all Markdown files.

EOF
  echo "LSP + context files configured!"
}
function configureGit(){
  echo "configuring git..."
  git config --global alias.ci commit
  git config --global alias.co checkout
  git config --global alias.st status
  echo "git configured!"
}

installBaseDeps
installCrush &
installBun &
installGoDeps &
configureCrush &
configureGit &
wait
git pull --all