# Roadmap

## Now: Productized CLI

- Keep the Alpha Cod/swimmer loop CLI-first.
- Keep local state compact and inspectable.
- Make install, doctor, dogfood, release, and recovery paths boring.
- Keep local release artifacts reproducible: archives, checksums, version output, and a Homebrew formula draft.
- Dogfood real coding tasks before adding a GUI.

## Next: Better Operator Experience

- Dogfood `start`, `run`, `gauntlet`, `doctor`, `inbox`, `status`, `resume`, `retry`, `rollback`, and `explain-failure` on real repos.
- Keep provider setup simple first: Codex CLI and Kimi CLI through adapter presets, OpenRouter through provider wizard plus a real `OPENROUTER_API_KEY`.
- Treat Codex/Kimi/OpenRouter missing key or login states as setup blockers, not product failures.
- Keep gauntlet reports honest when evidence is partial/incomplete.
- Tighten text output from dogfood notes instead of adding more flags.
- Keep shell completions in sync whenever command names change.

## Later: Product Layer

- Polish the current stdin-driven TUI into a richer full-screen TUI only after CLI workflows are stable.
- Add project dashboards over local job history.
- Add richer adapter packs only if the thin Codex CLI, Claude Code, OpenCode, Aider, and Goose scripts prove useful.
- Add optional remote sync only after local privacy-first mode is proven.

## Anti-Goals

- Do not become a full IDE clone.
- Do not dump whole repositories into every prompt.
- Do not run several agents on the same job without a single owner.
- Do not force one premium model to do every task.
- Do not add a GUI before the CLI is trusted on real tasks.
