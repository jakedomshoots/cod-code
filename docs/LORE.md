# Cod Code Lore

Cod Code is the product name and operator-facing story for this CLI harness. The existing `ceo-packet` binary, JSON field names, config keys, and evidence paths stay stable for v0.1 compatibility; the lore gives the harness a clearer voice without breaking scripts.

## The Concept

Cod Code is a sub-agent native CLI harness. A CEO agent, the **Alpha Cod**, delegates work to sub-agents, the **swimmers**, reviews their output, and decides what gets merged. The vocabulary below reskins the standard agent-harness concepts into a consistent fish/shoal theme.

## Core Lore Blurb

> Deep beneath the terminal, the Alpha Cod leads its shoal. When work comes in, the Cod casts its swimmers into the current — fry for the simple stuff, seasoned swimmers for the hard problems. Each comes back to surface with its catch. The Cod inspects every catch before it's netted into your code. Nothing gets merged that hasn't passed the Cod's gaze.

## Terminology Reference

| Existing concept | Cod Code term | Why it fits |
|---|---|---|
| CEO/orchestrator agent | **The Alpha Cod** or **Head Cod** | Leads the shoal and makes the final call. |
| Sub-agent, individual | **Swimmer** | Clean in CLI output: `Swimmer #3 surfaced`. |
| Sub-agents, group | **The Shoal** | A bounded group swimming together for one purpose. |
| Cheaper/smaller model | **Fry** | Small, numerous swimmers for simple work. |
| Delegating a task | **Casting** | The Alpha Cod casts swimmers into the current. |
| Delegated work item/result | **Catch** | Each swimmer brings back a completed catch. |
| Worker returns work | **Surfacing** | The swimmer returns with evidence. |
| CEO review/approval | **Inspecting the Catch** | The Alpha Cod reviews before merge. |
| Rejected/retry work | **Tossed Back** | The catch needs another pass. |
| Accepted/merged work | **Netted** | The catch is kept and merged. |
| Context window | **The Tank** | Each swimmer works in an isolated tank. |
| Shared long-term memory | **The Reef** | Persistent knowledge base. |
| Agent tools | **Tackle** | The swimmer's available tools. |
| Failed task | **Belly-up** or **Caught in the Net** | Stuck, timed out, or failed run. |
| Task queue | **The Current** | Work flows through it to swimmers. |
| System prompt/config | **The Chart** | The guide each swimmer follows. |
| Logs/history | **The Ledger** or **Ship's Log** | Saved, inspectable run history. |
| Full run/session | **Voyage** or **Run of the Tide** | A complete Cod Code session. |

## State Machine

Keep swimmer states fixed so terminal renderers can color them consistently:

- `swimming` — actively working
- `surfaced` — returned with a catch, awaiting review
- `inspecting` — Alpha Cod is reviewing the catch
- `reworking` — tossed back, retrying
- `netted` — approved and merged
- `belly-up` — hard failure

## CLI-Facing Strings

Use these in human-facing terminal output when the mode is interactive or verbose. Keep CI and JSON output plain and stable.

### Startup / session init

```text
🐟 Cod Code v0.1 — the shoal awakens
🎣 Alpha Cod is charting the work...
```

### Delegation

```text
🌊 Alpha Cod casts 3 swimmers into the current
🐟 Swimmer #1 -> tackle: [file-edit, bash]  |  task: refactor auth module
🐠 Fry #2 -> tackle: [file-edit]            |  task: update README
🐟 Swimmer #3 -> tackle: [bash, test-runner]|  task: fix failing tests
```

### In progress

```text
Swimmer #1 is swimming... (14s)
Fry #2 is swimming... (6s)
Swimmer #3 caught in a current (retrying...)
```

### Surfacing and review

```text
Swimmer #1 surfaced with its catch.
Alpha Cod is inspecting Swimmer #1's catch.
Catch approved — netted into main.
Catch tossed back — Swimmer #1 sent to rework (reason: tests failing).
```

### Failure states

```text
Swimmer #3 has gone belly-up (error: timeout).
Fry #2 caught in the net — task aborted.
3 swimmers surfaced, 1 gone belly-up, 2 netted.
```

### Session summary

```text
Voyage complete.
  Swimmers cast: 3
  Catches netted: 2
  Tossed back: 1
  Belly-up: 0
  Time in the current: 42s
```

### Memory / persistent context

```text
Consulting the Reef for prior context...
3 entries pulled from the Reef.
```

## Live Status Table Mockup

```text
COD CODE — Voyage in progress

ID   AGENT      STATE        TASK                         TIME
--   -----      -----        ----                         ----
#1   Swimmer    swimming     refactor auth module          18s
#2   Fry        surfaced     update README                 6s
#3   Swimmer    inspecting   fix failing tests             22s
#4   Fry        belly-up     lint config                   4s

Alpha Cod: reviewing catch from Swimmer #3...
Cast: 4   Netted: 0   Tossed back: 0   Belly-up: 1
```

## Implementation Notes

- Prefer plain-text fallbacks in non-verbose output: `[SWIMMER 1] surfaced`, `[ALPHA] netted`.
- Keep emoji minimal and opt-in; dense logs should stay readable in terminals and CI.
- Keep stable schema names like `ceo_review`, `subagent_results`, and `ceo-packet` until a versioned breaking-change plan exists.
- Use lore terms in documentation, help text, and interactive status lines first; do not rename persisted JSON fields casually.
- The Alpha Cod's verdict line should be short and factual, not flavor-only.
