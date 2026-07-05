# Quick First Task

Use this when trying Cod Code in a real repo for the first time.

```sh
PREFIX="$PWD/.local" sh scripts/install-local.sh
export PATH="$PWD/.local/bin:$PATH"
ceo-packet --quickstart /path/to/repo --format text
ceo-packet --workspace /path/to/repo --plan-only --format text -- "Fix one failing test"
```

When the preview looks right, run with a real check:

```sh
ceo-packet --workspace /path/to/repo --check "go test ./... -count=1" --format text -- "Fix one failing test"
```

For a safe no-write pass:

```sh
ceo-packet --workspace /path/to/repo --dry-run --format text -- "Inspect the failing test"
```

Provider setup is separate from install:

```sh
ceo-packet --workspace /path/to/repo --provider-wizard openai --http-model gpt-5 --format text
ceo-packet --workspace /path/to/repo --doctor-provider main --format text
```

