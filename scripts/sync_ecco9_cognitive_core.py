#!/usr/bin/env python3
"""Import an immutable, hash-manifested Go lineage snapshot from o9nn/ecco9.

The imported module is intentionally a nested compatibility boundary. All tracked
Go files are preserved byte-for-byte, while only explicitly reviewed packages are
activated by the canonical echo9llama adapter.
"""

from __future__ import annotations

import argparse
import glob
import hashlib
import json
import os
import pathlib
import shlex
import shutil
import subprocess
import sys
from typing import Iterable


ACTIVE_SOURCE_PATHS = {
    "core/ecco9/platform.go",
    "core/ecco9/types.go",
    "core/ecco9/drivers/reservoir_driver.go",
    "core/ecco9/drivers/memory_driver.go",
    "core/ecco9/drivers/emotion_driver.go",
    "core/ecco9/drivers/consciousness_driver.go",
}
DEFAULT_SOURCE_REPOSITORY = "https://github.com/o9nn/ecco9.git"
DEFAULT_SOURCE_COMMIT = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"


def run(repo: pathlib.Path, *args: str) -> str:
    return subprocess.check_output(args, cwd=repo, text=True).strip()


def sha256(path: pathlib.Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def canonical_repository(url: str) -> str:
    normalized = url.strip().removesuffix(".git")
    if normalized.startswith("git@github.com:"):
        normalized = "https://github.com/" + normalized.removeprefix("git@github.com:")
    return normalized.lower()


def tracked_go_files(source: pathlib.Path) -> list[pathlib.Path]:
    output = run(source, "git", "ls-files", "*.go")
    return [pathlib.Path(line) for line in output.splitlines() if line.strip()]


def tracked_license_files(source: pathlib.Path) -> list[pathlib.Path]:
    output = run(source, "git", "ls-files")
    names = {"license", "copying", "notice"}
    return [
        pathlib.Path(line)
        for line in output.splitlines()
        if pathlib.Path(line).name.lower().split(".", 1)[0] in names
    ]


def embedded_assets(source: pathlib.Path, go_files: Iterable[pathlib.Path]) -> set[pathlib.Path]:
    assets: set[pathlib.Path] = set()
    tracked = {pathlib.Path(line) for line in run(source, "git", "ls-files").splitlines() if line.strip()}
    for relative in go_files:
        absolute = source / relative
        for line in absolute.read_text(encoding="utf-8", errors="replace").splitlines():
            stripped = line.strip()
            if not stripped.startswith("//go:embed"):
                continue
            expression = stripped.removeprefix("//go:embed").strip()
            for pattern in shlex.split(expression):
                if pattern.startswith("all:"):
                    pattern = pattern[4:]
                full_pattern = absolute.parent / pattern
                for match in glob.glob(str(full_pattern), recursive=True):
                    candidate = pathlib.Path(match)
                    relative_candidate = candidate.relative_to(source)
                    if candidate.is_file() and relative_candidate in tracked:
                        assets.add(relative_candidate)
    return assets


def copy_files(source: pathlib.Path, destination: pathlib.Path, files: Iterable[pathlib.Path]) -> None:
    for relative in sorted(set(files), key=lambda path: path.as_posix()):
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source / relative, target)


def write_integration_module(destination: pathlib.Path, upstream_go_mod: pathlib.Path) -> None:
    shutil.copy2(upstream_go_mod, destination / "go.mod.upstream")
    upstream_sum = upstream_go_mod.with_name("go.sum")
    if upstream_sum.exists():
        shutil.copy2(upstream_sum, destination / "go.sum.upstream")
    (destination / "go.mod").write_text(
        "module github.com/EchoCog/echollama\n\ngo 1.25.0\n\ntoolchain go1.25.13\n",
        encoding="utf-8",
    )


def write_disposition_catalogue(
    source: pathlib.Path,
    target_root: pathlib.Path,
    destination: pathlib.Path,
    go_files: Iterable[pathlib.Path],
) -> dict[str, int]:
    counts: dict[str, int] = {}
    rows = []
    for relative in sorted(go_files, key=lambda path: path.as_posix()):
        source_hash = sha256(source / relative)
        target_path = target_root / relative
        if relative.as_posix() in ACTIVE_SOURCE_PATHS:
            disposition = "active_adapter_source"
            reason = "Activated only through core/cognitivecore with canonical lifecycle, persistence, and trust grading."
        elif target_path.is_file():
            if sha256(target_path) == source_hash:
                disposition = "canonical_same_path_exact"
                reason = "Canonical root already contains the exact source bytes; snapshot preserves provenance."
            else:
                disposition = "preserved_divergent"
                reason = "Conflicts with a canonical implementation and requires a separate reviewed adapter or reconciliation."
        else:
            disposition = "preserved_unreviewed"
            reason = "Absent from the canonical root and quarantined pending build, concurrency, evidence, and authority review."
        counts[disposition] = counts.get(disposition, 0) + 1
        rows.append(
            {
                "path": relative.as_posix(),
                "sha256": source_hash,
                "disposition": disposition,
                "reason": reason,
            }
        )
    catalogue = {
        "schema_version": 1,
        "active_adapter": "core/cognitivecore",
        "counts": counts,
        "files": rows,
    }
    (destination / "INTEGRATION_CATALOGUE.json").write_text(
        json.dumps(catalogue, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    return counts


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("source", type=pathlib.Path)
    parser.add_argument(
        "--destination",
        type=pathlib.Path,
        default=pathlib.Path("cognitive-core/ecco9"),
    )
    parser.add_argument("--target-root", type=pathlib.Path, default=pathlib.Path("."))
    args = parser.parse_args()

    source = args.source.resolve()
    destination = args.destination.resolve()
    target_root = args.target_root.resolve()
    if not (source / ".git").exists():
        raise SystemExit(f"source is not a Git repository: {source}")
    try:
        destination.relative_to(target_root)
    except ValueError as exc:
        raise SystemExit(f"destination must be inside target root: {target_root}") from exc
    if destination == target_root:
        raise SystemExit("destination must not equal target root")
    if run(source, "git", "status", "--porcelain", "--untracked-files=no"):
        raise SystemExit("source repository has modified tracked files")

    commit = run(source, "git", "rev-parse", "HEAD")
    commit_time = run(source, "git", "show", "-s", "--format=%cI", "HEAD")
    repository = run(source, "git", "remote", "get-url", "origin")
    if commit != DEFAULT_SOURCE_COMMIT:
        raise SystemExit(f"source HEAD {commit} does not match pinned commit {DEFAULT_SOURCE_COMMIT}")
    if canonical_repository(repository) != canonical_repository(DEFAULT_SOURCE_REPOSITORY):
        raise SystemExit(f"source repository {repository} does not match pinned repository {DEFAULT_SOURCE_REPOSITORY}")
    go_files = tracked_go_files(source)
    assets = embedded_assets(source, go_files)
    licenses = tracked_license_files(source)

    if destination.exists():
        shutil.rmtree(destination)
    destination.mkdir(parents=True)
    copy_files(source, destination, [*go_files, *assets, *licenses])
    write_integration_module(destination, source / "go.mod")
    disposition_counts = write_disposition_catalogue(source, target_root, destination, go_files)
    catalogue_hash = sha256(destination / "INTEGRATION_CATALOGUE.json")

    records = []
    for relative in sorted([*go_files, *assets, *licenses], key=lambda path: path.as_posix()):
        target = destination / relative
        kind = "license"
        if relative.suffix == ".go":
            kind = "go"
        elif relative in assets:
            kind = "embed_asset"
        records.append(
            {
                "path": relative.as_posix(),
                "kind": kind,
                "bytes": target.stat().st_size,
                "sha256": sha256(target),
            }
        )

    manifest = {
        "schema_version": 1,
        "source_repository": repository,
        "source_commit": commit,
        "source_ref": f"{repository}@{commit}",
        "source_committed_at": commit_time,
        "go_file_count": len(go_files),
        "embed_asset_count": len(assets),
        "license_file_count": len(licenses),
        "default_trust_grade": "preserved_unreviewed",
        "activation_policy": "explicit_adapter_allowlist",
        "files": records,
    }
    manifest_path = destination / "LINEAGE_MANIFEST.json"
    manifest_path.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    manifest_hash = sha256(manifest_path)

    notice = f"""# o9nn/ecco9 Cognitive-Core Lineage Snapshot

This directory preserves **all {len(go_files)} tracked Go files** from
`{repository}` at commit `{commit}` plus {len(assets)} files required by Go embed
directives and {len(licenses)} applicable license files. Every imported byte is
inventoried in `LINEAGE_MANIFEST.json` (manifest SHA-256 `{manifest_hash}`).

The snapshot is a **nested Go module and evidence boundary**, not a second
canonical runtime. Files retain their upstream package paths and are not copied
into root packages where duplicate symbols and stale APIs could silently replace
verified behavior. The root `core/cognitivecore` adapter activates only reviewed
reservoir, memory, emotion, and consciousness driver contracts. All other source
packages remain preserved for staged repair and promotion.

`INTEGRATION_CATALOGUE.json` assigns every preserved Go file an explicit
disposition. Current counts are `{json.dumps(disposition_counts, sort_keys=True)}`;
catalogue SHA-256 is `{catalogue_hash}`.

`go.mod.upstream` and `go.sum.upstream` preserve the upstream module graph. The
minimal local `go.mod` prevents unreviewed transitive dependencies from entering
the canonical security boundary merely because their source was preserved.

Regenerate from a clean source clone with:

```bash
git clone {DEFAULT_SOURCE_REPOSITORY} /tmp/ecco9-source
git -C /tmp/ecco9-source checkout --detach {DEFAULT_SOURCE_COMMIT}
python3 scripts/sync_ecco9_cognitive_core.py /tmp/ecco9-source
```
"""
    (destination / "README.md").write_text(notice, encoding="utf-8")

    print(json.dumps({
        "source_commit": commit,
        "go_file_count": len(go_files),
        "embed_asset_count": len(assets),
        "license_file_count": len(licenses),
        "disposition_counts": disposition_counts,
        "catalogue_sha256": catalogue_hash,
        "manifest_sha256": manifest_hash,
        "destination": str(destination),
    }, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
