# mctop

A terminal client for MCP servers. Connect to any server, browse its tools,
resources, and prompts, call them, and watch the result, without leaving the
shell. Then assert the server's contract in CI so a renamed tool or a drifted
schema fails the build instead of breaking an agent in production.

Think `curl` and `k9s`, but for the Model Context Protocol.

> Status: early. Building in the open. See [DESIGN.md](./DESIGN.md).

## What it does

- **Explore** a server in a TUI: tools, resources, and prompts, with their
  schemas, and call any tool with real arguments.
- **Script** it headless: `mctop ls` to list, `mctop call` for one-shot calls.
- **Test** it in CI: `mctop test spec.yaml` runs a contract and exits non-zero
  when it breaks.

Works over stdio (it spawns the server) and Streamable HTTP (it connects to a
URL).

## Install

```
go install github.com/aloki-alok/mctop@latest
```

## Usage

```
mctop <target>                 open the TUI against a server
mctop ls <target>              list tools, resources, and prompts
mctop call <target> <tool>     call one tool and print the result
mctop test <spec.yaml>         run a contract, exit 0 on pass, 1 on fail
```

A target is either a command to spawn (`"uvx mcp-server-time"`) or an
`http(s)://` URL.

## License

MIT.
