#!/usr/bin/env python3
"""Extract and normalize Cod Code model-command JSON output."""

import json
import os
import pathlib
import re
import sys
from typing import Any


raw_path = sys.argv[1]
prompt = sys.argv[2] if len(sys.argv) > 2 else ""
raw = pathlib.Path(raw_path).read_text(encoding="utf-8")


def prompt_line(prefix: str) -> str:
    for line in prompt.splitlines():
        if line.startswith(prefix):
            return line[len(prefix) :].strip().rstrip(".")
    return ""


def prompt_list(prefix: str) -> list[str]:
    value = prompt_line(prefix)
    if not value:
        return []
    return [part.strip().rstrip(".") for part in value.split(",") if part.strip()]


def candidate_names() -> list[str]:
    names: list[str] = []
    in_block = False
    for line in prompt.splitlines():
        if line.strip() == "candidate_subagents:":
            in_block = True
            continue
        if not in_block:
            continue
        if not line.startswith("- "):
            if names:
                break
            continue
        name = line[2:].split(None, 1)[0].strip()
        if name and name not in names:
            names.append(name)
    return names


def safe_relative(path: str) -> pathlib.Path | None:
    clean = pathlib.PurePosixPath(str(path).replace(os.sep, "/"))
    if clean.is_absolute() or ".." in clean.parts:
        return None
    return pathlib.Path(*clean.parts)


def normalize_patch(patch: Any) -> None:
    if not isinstance(patch, dict):
        return
    raw_patch_path = patch.get("path")
    if not isinstance(raw_patch_path, str) or not raw_patch_path.strip():
        return
    rel = safe_relative(raw_patch_path)
    if rel is None:
        return
    content = patch.get("content")
    old = patch.get("old")
    new = patch.get("new")
    has_content = isinstance(content, str) and content != ""
    has_old = isinstance(old, str) and old != ""
    has_new = isinstance(new, str) and new != ""
    if has_content:
        try:
            current = rel.read_text(encoding="utf-8")
        except FileNotFoundError:
            return
        patch["old"] = current
        patch["new"] = content
        patch.pop("content", None)
        return
    if has_new and not has_old:
        try:
            current = rel.read_text(encoding="utf-8")
        except FileNotFoundError:
            patch["content"] = new
            patch.pop("old", None)
            patch.pop("new", None)
            return
        patch["old"] = current


def validate_required_diff_terms(payload: dict[str, Any]) -> None:
    required_files = set(prompt_list("Required changed files: "))
    terms = prompt_list("Required diff terms: ")
    if not required_files or not terms:
        return
    patches = payload.get("patches")
    if not isinstance(patches, list):
        return
    text_by_path: dict[str, str] = {}
    for patch in patches:
        if not isinstance(patch, dict):
            continue
        path = patch.get("path")
        if path not in required_files:
            continue
        text = patch.get("new")
        if not isinstance(text, str) or text == "":
            text = patch.get("content")
        if isinstance(text, str) and text != "":
            text_by_path[path] = text_by_path.get(path, "") + "\n" + text
    for path, text in text_by_path.items():
        missing = [term for term in terms if term not in text]
        if missing:
            raise SystemExit(f"required diff terms missing from patch for {path}: {', '.join(missing)}")


def normalize_delegation(payload: dict[str, Any]) -> None:
    if "selected_subagents" not in payload:
        return
    candidates = candidate_names()
    if not candidates:
        return
    selected = payload.get("selected_subagents")
    valid = [name for name in selected if isinstance(name, str) and name in candidates] if isinstance(selected, list) else []
    if valid:
        payload["selected_subagents"] = valid[: len(candidates)]
        return
    fallback = "coder" if "coder" in candidates else candidates[0]
    payload["selected_subagents"] = [fallback]
    payload.setdefault("summary", f"Select {fallback} for the narrow coding task.")


def normalize_payload(payload: Any) -> Any:
    if not isinstance(payload, dict):
        return payload
    patches = payload.get("patches")
    if isinstance(patches, list):
        for patch in patches:
            normalize_patch(patch)
    normalize_delegation(payload)
    validate_required_diff_terms(payload)
    return payload


def emit(payload: Any) -> None:
    print(json.dumps(normalize_payload(payload), separators=(",", ":")))
    raise SystemExit(0)


def try_payload(text: str) -> None:
    text = text.strip()
    if not text:
        return
    if text.startswith("{"):
        try:
            parsed = json.loads(text)
            if isinstance(parsed, dict):
                emit(parsed)
        except json.JSONDecodeError:
            pass
    fenced = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL)
    if fenced:
        try:
            emit(json.loads(fenced.group(1).strip()))
        except json.JSONDecodeError:
            pass
    decoder = json.JSONDecoder()
    for index, char in enumerate(text):
        if char != "{":
            continue
        try:
            payload, _ = decoder.raw_decode(text[index:])
        except json.JSONDecodeError:
            continue
        if isinstance(payload, dict):
            emit(payload)


def collect_text(value: Any) -> str:
    if isinstance(value, str):
        return value
    if isinstance(value, list):
        parts: list[str] = []
        for item in value:
            if isinstance(item, str):
                parts.append(item)
            elif isinstance(item, dict):
                for key in ("text", "content", "message"):
                    if isinstance(item.get(key), str):
                        parts.append(item[key])
        return "\n".join(parts)
    if isinstance(value, dict):
        for key in ("text", "content", "message", "result", "response"):
            if key in value:
                return collect_text(value[key])
    return ""


def inspect_event(event: Any) -> None:
    if not isinstance(event, dict):
        return
    for key in ("status", "recommended_verdict", "selected_subagents", "patches"):
        if key in event:
            emit(event)
    text = collect_text(event)
    if text:
        try_payload(text)


try:
    outer = json.loads(raw.strip())
except json.JSONDecodeError:
    outer = None

if outer is not None:
    inspect_event(outer)
    if isinstance(outer, list):
        for item in outer:
            inspect_event(item)

for line in raw.splitlines():
    try:
        event = json.loads(line)
    except json.JSONDecodeError:
        continue
    inspect_event(event)

try_payload(raw)
raise SystemExit("model command returned no JSON object")
