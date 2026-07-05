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
		return completionFile{name: "_cod", content: zshCompletion()}, nil
	case "bash":
		return completionFile{name: "cod.bash", content: bashCompletion()}, nil
	case "fish":
		return completionFile{name: "cod.fish", content: fishCompletion()}, nil
	default:
		return completionFile{}, fmt.Errorf("--shell requires zsh, bash, or fish")
	}
}

func completionWords() string {
	return "help chat dev tui start run gauntlet doctor inbox status oauth browser computer tools production-status production-actions production-finalize resume retry rollback explain-failure review context config eval"
}

func oauthCommandWords() string {
	return "list doctor init"
}

func browserCommandWords() string {
	return "doctor manifest read"
}

func computerCommandWords() string {
	return "doctor manifest snapshot"
}

func toolsCommandWords() string {
	return "manifest"
}

func oauthProviderWords() string {
	return "kimi codex claude opencode goose"
}

func productionActionStateWords() string {
	return "ready missing_env empty_env setup_blocked waiting"
}

func productionActionProviderWords() string {
	return "openai openrouter kimi-code moonshot minimax kimi codex"
}

func zshCompletion() string {
	return `#compdef cod
local -a commands
commands=(` + completionWords() + `)
_arguments \
  '1:command:((` + completionWords() + `))' \
  '*::arg:->args'
case $words[2] in
  config)
    _arguments '2:config command:((check doctor explain completions init))' \
      '--shell[completion shell]:shell:(zsh bash fish)' \
      '--output-dir[write completion file to directory]:directory:_files -/'
    ;;
  production-actions)
    _arguments \
      '--action-state[action state]:state:(` + productionActionStateWords() + `)' \
      '--action-kind[action kind]:kind:(release_proof provider_proof competitor_setup comparison final_readiness)' \
      '--action-provider[provider]:provider:(` + productionActionProviderWords() + `)' \
      '--format[output format]:format:(json text events)'
    ;;
  oauth)
    _arguments '2:oauth command:((` + oauthCommandWords() + `))' \
      '3:oauth provider:((` + oauthProviderWords() + `))' \
      '--workspace[workspace directory]:directory:_files -/' \
      '--format[output format]:format:(json text)'
    ;;
  browser)
    _arguments '2:browser command:((` + browserCommandWords() + `))' \
      '--browser-policy[policy]:policy:(deny ask allow-localhost allow)' \
      '--format[output format]:format:(json text)'
    ;;
  computer)
    _arguments '2:computer command:((` + computerCommandWords() + `))' \
      '--computer-policy[policy]:policy:(deny ask allow)' \
      '--format[output format]:format:(json text)'
    ;;
  tools)
    _arguments '2:tools command:((` + toolsCommandWords() + `))' \
      '--format[output format]:format:(json text)'
    ;;
esac
`
}

func bashCompletion() string {
	return `_cod() {
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
  if [[ "${COMP_WORDS[1]}" == oauth && "${COMP_CWORD}" == 2 ]]; then
    COMPREPLY=( $(compgen -W "` + oauthCommandWords() + `" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == oauth && "${COMP_CWORD}" == 3 ]]; then
    COMPREPLY=( $(compgen -W "` + oauthProviderWords() + `" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == browser && "${COMP_CWORD}" == 2 ]]; then
    COMPREPLY=( $(compgen -W "` + browserCommandWords() + `" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == computer && "${COMP_CWORD}" == 2 ]]; then
    COMPREPLY=( $(compgen -W "` + computerCommandWords() + `" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == tools && "${COMP_CWORD}" == 2 ]]; then
    COMPREPLY=( $(compgen -W "` + toolsCommandWords() + `" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == gauntlet && "$prev" == --agents ]]; then
    COMPREPLY=( $(compgen -W "ceo_harness codex_cli claude_code aider opencode goose pi oh_my_pi" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == production-actions && "$prev" == --action-state ]]; then
    COMPREPLY=( $(compgen -W "` + productionActionStateWords() + `" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == production-actions && "$prev" == --action-kind ]]; then
    COMPREPLY=( $(compgen -W "release_proof provider_proof competitor_setup comparison final_readiness" -- "$cur") )
    return 0
  fi
  if [[ "${COMP_WORDS[1]}" == production-actions && "$prev" == --action-provider ]]; then
    COMPREPLY=( $(compgen -W "` + productionActionProviderWords() + `" -- "$cur") )
    return 0
  fi
  if [[ "$prev" == --shell ]]; then
    COMPREPLY=( $(compgen -W "zsh bash fish" -- "$cur") )
    return 0
  fi
}
complete -F _cod cod
`
}

func fishCompletion() string {
	return `complete -c cod -f -n "__fish_use_subcommand" -a "` + completionWords() + `"
complete -c cod -f -n "__fish_seen_subcommand_from config" -a "check doctor explain completions init"
complete -c cod -f -n "__fish_seen_subcommand_from oauth" -a "` + oauthCommandWords() + `"
complete -c cod -n "__fish_seen_subcommand_from oauth" -a "` + oauthProviderWords() + `"
complete -c cod -f -n "__fish_seen_subcommand_from browser" -a "` + browserCommandWords() + `"
complete -c cod -f -n "__fish_seen_subcommand_from computer" -a "` + computerCommandWords() + `"
complete -c cod -f -n "__fish_seen_subcommand_from tools" -a "` + toolsCommandWords() + `"
complete -c cod -n "__fish_seen_subcommand_from gauntlet" -l agents -a "ceo_harness codex_cli claude_code aider opencode goose pi oh_my_pi"
complete -c cod -n "__fish_seen_subcommand_from gauntlet" -l output-dir -r
complete -c cod -n "__fish_seen_subcommand_from production-actions" -l action-state -a "` + productionActionStateWords() + `"
complete -c cod -n "__fish_seen_subcommand_from production-actions" -l action-kind -a "release_proof provider_proof competitor_setup comparison final_readiness"
complete -c cod -n "__fish_seen_subcommand_from production-actions" -l action-provider -a "` + productionActionProviderWords() + `"
complete -c cod -n "__fish_seen_subcommand_from completions" -l shell -a "zsh bash fish"
complete -c cod -n "__fish_seen_subcommand_from completions" -l output-dir -r
`
}
