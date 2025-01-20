# hledger-language-server

This is a hobby project to bring language support for [hledger](https://hledger.org/) files to editors. I am mainly working with neovim, so integration is not tested with other editors.

## Features

- Completion for account names

## Note

This collects telemetry data using open telemetry. By default it sends this data to an open telemetry collector at localhost, which you probably don't have. If you don't set this up and don't provide a collector via environmont variables, no telemetry data will be collected. I don't collect your data.
I'm doing this just for fun and out of curiosity with my own data.

## How to use

1. Install hledger-language-server in a way that is appropriate for your OS. Since this software is not distributed to anywhere yet, your best bet in cloning the code, compiling it yourself and putting it in the appropriate location in your system.
2. Configure your editor to know where the binary is and to use it for ledger files. I have this in my neovim lps configuration (using nvim-lspconfig):
```lua
if not lspConfigurations.hledger_ls then
  lspConfigurations.hledger_ls = {
    default_config = {
      cmd = { "/home/yeldir/querbeet/workspace/private/projects/hledger-language-server/hledger-language-server" },
      filetypes = { "ledger" },
      root_dir = require("lspconfig.util").root_pattern(".git", "*.journal"),
      settings = {},
    },
  }
end

add_lsp(lspconfig.hledger_ls, {})
```
3. You might need to tell your editor to recognize ledger files.

## Development

If you want to make contributions, please first talk to me. For development and experimentation, clone the repo, install devbox, cd into the project directory and run `devbox shell`. Run `go build .` or `go test ./...` to get all dependencies and see that the setup works.
