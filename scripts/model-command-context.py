#!/usr/bin/env python3
"""Emit compact workspace snippets for model-command wrappers."""

import os
import pathlib
import sys


def prompt_line(prompt: str, prefix: str) -> str:
    for line in prompt.splitlines():
        if line.startswith(prefix):
            return line[len(prefix) :].strip().rstrip(".")
    return ""


def prompt_list(prompt: str, prefix: str) -> list[str]:
    value = prompt_line(prompt, prefix)
    if not value:
        return []
    return [part.strip().rstrip(".") for part in value.split(",") if part.strip()]


def safe_relative(path: str) -> pathlib.Path | None:
    clean = pathlib.PurePosixPath(path.replace(os.sep, "/"))
    if clean.is_absolute() or ".." in clean.parts:
        return None
    return pathlib.Path(*clean.parts)


def sibling_context_files(paths: list[str]) -> list[str]:
    extras: list[str] = []
    seen = set(paths)
    for raw_path in paths:
        rel = safe_relative(raw_path)
        if rel is None:
            continue
        candidates: list[pathlib.Path] = []
        if rel.suffix == ".go" and not rel.name.endswith("_test.go"):
            candidates.append(rel.with_name(rel.stem + "_test.go"))
        if rel.suffix in {".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"} and not any(
            part in rel.stem for part in [".test", ".spec"]
        ):
            candidates.append(rel.with_name(rel.stem + ".test" + rel.suffix))
            candidates.append(rel.with_name(rel.stem + ".spec" + rel.suffix))
        for test_name in candidates:
            test_raw = test_name.as_posix()
            if test_raw not in seen and test_name.exists():
                extras.append(test_raw)
                seen.add(test_raw)
    return extras


def main() -> int:
    prompt = sys.argv[1] if len(sys.argv) > 1 else sys.stdin.read()
    changed_files = prompt_list(prompt, "Required changed files: ")
    artifact_files = prompt_list(prompt, "Required evidence artifacts: ")
    context_files = changed_files + sibling_context_files(changed_files)
    lines: list[str] = []
    if context_files:
        lines.append("Workspace snippets:")
        for raw_path in context_files:
            rel = safe_relative(raw_path)
            if rel is None:
                lines.append(f"File: {raw_path}")
                lines.append("Status: skipped unsafe path")
                continue
            lines.append(f"File: {raw_path}")
            try:
                content = rel.read_text(encoding="utf-8")
            except FileNotFoundError:
                lines.append("Status: missing")
                continue
            if len(content) > 12000:
                content = content[:12000] + "\n[truncated]\n"
            lines.append("Content:")
            lines.append(content)
    diff_terms = prompt_list(prompt, "Required diff terms: ")
    if diff_terms:
        lines.append("Required diff terms parsed from harness prompt:")
        for term in diff_terms:
            lines.append(f"- {term}")
    if artifact_files:
        lines.append("Required evidence artifacts parsed from harness prompt:")
        for artifact in artifact_files:
            lines.append(f"- {artifact}")
    print("\n".join(lines))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
