# mctop: design doc

Status: APPROVED (problem, non-goals, rollback signed off 2026-06-18). Building.
Date: 2026-06-18. Owner: Ryu (Alok). Personal project.

Name: **mctop** (chosen 2026-06-18), in the `htop` / `k9s` lineage: a terminal
tool to watch and drive an MCP server. The binary is `mctop`.

## 1. Problem

MCP is now everywhere (≈97M monthly SDK downloads, 10k+ servers, a Linux
Foundation project), but the way devs poke at a server is still bad. The official
Inspector is a web app you launch in a browser. There is no terminal-native way
to connect to a server, see its tools/resources/prompts, call one with real
arguments, and watch the result, the way `curl`, Postman, or `k9s` let you do for
HTTP and Kubernetes. Worse, there is no easy way to **assert an MCP server's
contract in CI**: tools get renamed, input schemas drift, a call starts erroring,
and nothing catches it until an agent breaks in production. Devs who live in the
terminal want a fast TUI to explore a server and a one-command check they can
gate CI on.

## 2. Non-goals (v1)

- **Not a gateway/proxy/aggregator.** We connect to one server at a time, we do
  not route or multiplex many. (That is a separate idea.)
- **Not a server framework / scaffolder.** We do not generate MCP servers. (That
  is `mcpify`, a separate idea.)
- **Not a security scanner.** We do not judge tool descriptions for prompt
  injection. (Separate idea; mctop may later feed it.)
- **No MCP Apps UI rendering.** We show tool results as text/JSON, not as
  server-rendered UI.
- **No full OAuth flow in v1.** HTTP servers get a bearer token via flag/env;
  the browser OAuth dance is later.
- **No sampling / elicitation / resource-subscription** round-trips in v1.

## 3. Alternatives considered & rejected

- **Extend the web MCP Inspector**. Lost because the open gap is terminal-native
  + CI-gateable. A web app does not serve the `lazygit`/`k9s` crowd or run in CI.
- **A Go library only, no TUI**. Lost because the TUI is the viral hook (devs
  star tools they can *see*); a library alone does not spread.
- **Node/TypeScript**. Lost because single static binary beats `npx` friction
  for adoption, matches Ryu's ethos (crwl, scrub), and Go's Charm ecosystem
  (Bubble Tea) is the best TUI toolkit going.
- **Hand-roll the MCP protocol vs use the official Go SDK**. Leaning SDK (see
  Tradeoffs); the spec moves fast and tracking it by hand is a maintenance tax.

## 4. Chosen approach + system diagram

A single Go binary with one interactive mode (TUI) and several non-interactive
subcommands that share one MCP client core.

```
        mctop <target>            target = stdio command  OR  http(s) URL
              │
        ┌─────┴───────────────────────────────────────────┐
        │                 mcp client core                  │
        │  transport: stdio (spawn) | streamable-http      │
        │  methods: initialize, tools/list, tools/call,    │
        │           resources/list/read, prompts/list/get  │
        └─────┬───────────────────────────┬────────────────┘
              │                            │
   ┌──────────┴─────────┐      ┌───────────┴──────────────┐
   │   TUI (Bubble Tea) │      │  headless subcommands     │
   │  browse tools/res/ │      │  ls    list, pipeable     │
   │  prompts, fill args│      │  call  one-shot call      │
   │  from JSON Schema, │      │  test  run a spec, assert,│
   │  see result + raw  │      │        exit 0/1, --report │
   │  record -> spec    │      │  record  write a spec     │
   └────────────────────┘      └───────────────────────────┘
```

**The CI contract (`mctop test spec.yaml`).** Strict YAML (unknown keys are
errors, like callwright). Declares the server and expectations:

```yaml
server:
  command: ["python", "server.py"]   # stdio;  OR  url: https://host/mcp
expect:
  tools: [search, fetch]             # these tools must exist
calls:
  - tool: search
    args: { q: "hello" }
    assert:
      not_error: true                # call must not return isError
      contains: "result"             # substring in the text content
```

`test` connects, checks `expect`, runs each `call`, evaluates `assert`, prints a
readable report, exits 0 (all pass) or 1 (any fail), `--report json` for CI.

**Build order (small, verified, committed one by one):**
1. Repo skeleton + DESIGN + README + module.
2. MCP client core over **stdio**, verified against a real public server.
3. `ls` + `call` headless subcommands (prove the core end to end).
4. `test` (spec parse + assertions + exit codes), verified to pass AND fail.
5. Streamable-HTTP transport + bearer token, verified against a real HTTP server.
6. TUI: connect, browse, schema-driven arg form, call, result view.
7. `record` -> spec. Polish, `--report json`, docs.
8. Release: GoReleaser binaries, `go install`, installer script.

## 5. Tradeoffs

- **Deps, unlike crwl/scrub.** A TUI (Bubble Tea) and an MCP client are exactly
  where third-party code is justified (genuine complexity behind a seam). This
  tool will NOT be zero-dep. Accepted, with the dep set kept small and vetted.
- **Official Go SDK vs hand-rolled.** SDK = faster, tracks the spec, less protocol
  risk, but a moving dependency. Hand-rolled = zero protocol deps and full control,
  but a maintenance tax as the spec evolves. Default: SDK; revisit if it is heavy
  or drifts.
- **Spec moves fast.** MCP added Tasks, MCP Apps, stateless core in 2026. v1
  targets the stable core (tools/resources/prompts over stdio + HTTP); newer
  surfaces are explicit non-goals to avoid chasing a moving target.

## 6. Success + failure metrics

- **Success (v1 done):** connect to a real public MCP server over BOTH stdio and
  HTTP; browse and call a tool in the TUI; `mctop test` correctly passes on a
  good contract and fails on a broken one, all verified against a real server,
  not a mock. Adoption signal afterward: GitHub stars + a third party using
  `test` in their CI.
- **Failure (the cardinal sin):** `test` reports green when a call actually
  errored or the contract drifted (a false pass in CI). This mirrors crwl's
  reliability bar: a testing tool that lies is worse than none. Second: hangs or
  corrupts on a misbehaving/slow server (mitigate with hard per-call timeouts).

## 7. Rollout plan

Open-source public repo from the start. Release v0.1.0 only once v1 success
above is verified against real servers. Distribution: GoReleaser cross-compiled
binaries (like crwl-cli), `go install`, a `curl | sh` installer, Homebrew tap
later. No staged/dark rollout needed for a CLI; the release tag is the gate.

## 8. Rollback plan

It is a client-side CLI with no server and no stored user state, so rollback is
simple and forward-only: if a release is bad, yank/retract the tag, users pin the
previous version (`go install ...@vPREV` / previous binary). No data migration,
no infra to revert. Each feature lands as its own small commit, so reverting one
slice is a single `git revert`.

## 9. Monitoring + runbook

No telemetry, no phone-home (Ryu's ethos; the only network calls are the user's
own connections to the MCP server they target). "Monitoring" = GitHub issues +
CI on the repo. Runbook = README (usage) + this DESIGN (architecture). Tests run
in CI on every push.

## 10. Bus factor

Standard Go project layout, this DESIGN + a README, and tests on the client core
and the `test` spec engine. Someone else could pick it up from the docs alone.
The MCP method set and the spec schema are small and documented here.
