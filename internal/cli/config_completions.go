package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type completionFile struct {
	name    string
	content string
}

func runConfigCompletions(out io.Writer, opts options) error {
	file, err := completionForShell(opts.completionShell)
	if err != nil {
		return err
	}
	outputDir := strings.TrimSpace(opts.completionOutputDir)
	if outputDir == "" {
		return fmt.Errorf("--output-dir requires a directory")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create completion output dir: %w", err)
	}
	path := filepath.Join(outputDir, file.name)
	if err := os.WriteFile(path, []byte(file.content), 0o644); err != nil {
		return fmt.Errorf("write completion file: %w", err)
	}
	if _, err := fmt.Fprintf(out, "Wrote %s\n", path); err != nil {
		return fmt.Errorf("write completion result: %w", err)
	}
	return nil
}

func completionForShell(shell string) (completionFile, error) {
	switch strings.TrimSpace(shell) {
	case "zsh":
		return completionFile{name: "_ceo-packet", content: zshCompletion()}, nil
	case "bash":
		return completionFile{name: "ceo-packet.bash", content: bashCompletion()}, nil
	case "fish":
		return completionFile{name: "ceo-packet.fish", content: fishCompletion()}, nil
	default:
		return completionFile{}, fmt.Errorf("--shell requires zsh, bash, or fish")
	}
}

func completionWords() string {
	return "start run gauntlet doctor inbox status production-status production-actions production-finalize resume retry rollback explain-failure review context config eval"
}

func zshCompletion() string {
	return `#compdef ceo-packet
local -a commands
commands=(` + completionWords() + `)
_arguments \
  '1:command:((start run gauntlet doctor inbox status production-status production-actions production-finalize resume retry rollback explain-failure review context config eval))' \
  '*::arg:->args'
case $words[2] in
  config)
    _arguments '2:config command:((check doctor explain completions init))' \
      '--shell[completion shell]:shell:(zsh bash fish)' \
      '--output-dir[write completion file to directory]:directory:_files -/'
    ;;
esac
`
}

func bashCompletion() string {
	return `_ceo_packet() {
  local cur prev
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  if [[ "${COMP_CWORD}" == 1 ]]; then
    COMPREPLY=( $(compgen -W "` + completionWords() + `" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == config && "${COMP_CWORD}" == 2 ]]; then
    COMPREPLY=( $(compgen -W "check doctor explain completions init" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == gauntlet && "$prev" == --agents ]]; then
    COMPREPLY=( $(compgen -W "ceo_harness codex_cli opencode pi" -- "$cur") )
    return 0
  fi
  if [[ "$prev" == --shell ]]; then
    COMPREPLY=( $(compgen -W "zsh bash fish" -- "$cur") )
    return 0
  fi
}
complete -F _ceo_packet ceo-packet
`
}

func fishCompletion() string {
	return `complete -c ceo-packet -f -n "__fish_use_subcommand" -a "` + completionWords() + `"
complete -c ceo-packet -f -n "__fish_seen_subcommand_from config" -a "check doctor explain completions init"
complete -c ceo-packet -n "__fish_seen_subcommand_from gauntlet" -l agents -a "ceo_harness codex_cli opencode pi"
complete -c ceo-packet -n "__fish_seen_subcommand_from gauntlet" -l output-dir -r
complete -c ceo-packet -n "__fish_seen_subcommand_from completions" -l shell -a "zsh bash fish"
complete -c ceo-packet -n "__fish_seen_subcommand_from completions" -l output-dir -r
`
}
