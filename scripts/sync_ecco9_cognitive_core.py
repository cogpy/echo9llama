#!/usr/bin/env python3
"""Refresh immutable raw source without replacing retained runtime overlays.

The raw upstream/ snapshot is a separate nested compatibility boundary. All
tracked Go bytes and assets are preserved, not merged or automatically activated.
Legacy positional/destructive imports are deliberately no longer supported.
"""

from __future__ import annotations

import argparse
import glob
import hashlib
import json
import os
import pathlib
import re
import shlex
import shutil
import subprocess
import sys
import tempfile
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
DEFAULT_SOURCE_COMMIT = "7d32d715c633a92c684900fbaf25b0111c708612"
RETAINED_RUNTIME_COMMIT = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"
SOURCE_SNAPSHOT_MANIFEST = "SOURCE_SNAPSHOT_MANIFEST.json"


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
        "module github.com/EchoCog/echollama\n\ngo 1.26.0\n\ntoolchain go1.26.11\n",
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


LOCAL_PATCH_POLICY = (
    "model-only character text: canned, templated, random-choice and fallback "
    "output replaced by model queries; no text when the model is absent"
)


def repin_local_patches(destination: pathlib.Path, target_root: pathlib.Path) -> dict:
    """Re-pin the snapshot after in-place local patches.

    Every manifest record whose bytes changed keeps its upstream identity in
    upstream_sha256/upstream_bytes and gains local_patch; sha256/bytes then
    describe the patched file. The catalogue is re-derived from current bytes,
    the README notice hashes are refreshed, and the adapter pins in
    core/cognitivecore/lineage.go are rewritten. Source-snapshot refreshes never
    change these runtime files or their original upstream provenance.
    """
    manifest_path = destination / "LINEAGE_MANIFEST.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    patched = 0
    for record in manifest["files"]:
        path = destination / record["path"]
        current_hash = sha256(path)
        upstream_hash = record.get("upstream_sha256", record["sha256"])
        upstream_bytes = record.get("upstream_bytes", record["bytes"])
        if record["kind"] == "local_go":
            continue
        if current_hash == upstream_hash:
            for key in ("upstream_sha256", "upstream_bytes", "local_patch"):
                record.pop(key, None)
            record["sha256"], record["bytes"] = upstream_hash, upstream_bytes
            continue
        record["upstream_sha256"] = upstream_hash
        record["upstream_bytes"] = upstream_bytes
        record["local_patch"] = LOCAL_PATCH_POLICY
        record["sha256"] = current_hash
        record["bytes"] = path.stat().st_size
        patched += 1
    # Go files added locally (no upstream identity) are recorded separately,
    # so go_file_count and the catalogue keep describing the upstream import.
    recorded = {r["path"] for r in manifest["files"]}
    manifest["files"] = [r for r in manifest["files"] if r["kind"] != "local_go" or (destination / r["path"]).is_file()]
    added = 0
    for path in sorted(destination.rglob("*.go")):
        relative = path.relative_to(destination).as_posix()
        if relative.startswith("upstream/"):
            continue
        if relative in recorded and not any(r["path"] == relative and r["kind"] == "local_go" for r in manifest["files"]):
            continue
        manifest["files"] = [r for r in manifest["files"] if r["path"] != relative]
        manifest["files"].append({
            "path": relative,
            "kind": "local_go",
            "bytes": path.stat().st_size,
            "sha256": sha256(path),
            "local_patch": LOCAL_PATCH_POLICY,
        })
        added += 1
    manifest["files"].sort(key=lambda r: r["path"])
    manifest["local_addition_count"] = added
    manifest["local_patch_count"] = patched
    if patched or added:
        manifest["local_patch_policy"] = LOCAL_PATCH_POLICY
    else:
        manifest.pop("local_patch_policy", None)

    go_files = [pathlib.Path(r["path"]) for r in manifest["files"] if r["kind"] == "go"]
    counts = write_disposition_catalogue(destination, target_root, destination, go_files)
    catalogue_hash = sha256(destination / "INTEGRATION_CATALOGUE.json")
    manifest_path.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    manifest_hash = sha256(manifest_path)

    readme = destination / "README.md"
    if readme.exists():
        text = readme.read_text(encoding="utf-8")
        text = re.sub(r"manifest SHA-256 `[0-9a-f]{64}`", f"manifest SHA-256 `{manifest_hash}`", text)
        text = re.sub(r"catalogue SHA-256 is `[0-9a-f]{64}`", f"catalogue SHA-256 is `{catalogue_hash}`", text)
        text = re.sub(r"Current counts are `\{[^`]*\}`", f"Current counts are `{json.dumps(counts, sort_keys=True)}`", text)
        marker = "## Local patches"
        section = (
            f"{marker}\n\n{patched} preserved files carry local patches and {added} Go files were added locally ({LOCAL_PATCH_POLICY}). "
            "Each patched record in `LINEAGE_MANIFEST.json` keeps its upstream bytes and SHA-256 "
            "in `upstream_bytes`/`upstream_sha256`. Re-pin after editing with:\n\n"
            "```bash\npython3 scripts/sync_ecco9_cognitive_core.py --repin-local-patches\n```\n"
        )
        if marker in text:
            start = text.index(marker)
            following = text.find("\n## ", start + len(marker))
            text = text[:start] + section + (text[following:] if following >= 0 else "")
        elif patched or added:
            text = text.rstrip("\n") + "\n\n" + section
        readme.write_text(text, encoding="utf-8")

    lineage = target_root / "core" / "cognitivecore" / "lineage.go"
    if lineage.exists():
        text = lineage.read_text(encoding="utf-8")
        text = re.sub(r'(?m)^(\s*ManifestSHA256\s*=\s*)"[0-9a-f]{64}"', rf'\g<1>"{manifest_hash}"', text)
        text = re.sub(r'(?m)^(\s*CatalogueSHA256\s*=\s*)"[0-9a-f]{64}"', rf'\g<1>"{catalogue_hash}"', text)
        lineage.write_text(text, encoding="utf-8")

    return {
        "local_patch_count": patched,
        "local_addition_count": added,
        "disposition_counts": counts,
        "catalogue_sha256": catalogue_hash,
        "manifest_sha256": manifest_hash,
    }


def safe_relative(value: str) -> pathlib.Path:
    path = pathlib.PurePosixPath(value)
    if not value or path.is_absolute() or ".." in path.parts or path.as_posix() != value:
        raise ValueError(f"unsafe manifest path: {value!r}")
    return pathlib.Path(value)


def regular_bytes(root: pathlib.Path, relative: str) -> bytes:
    path = root / safe_relative(relative)
    if any(parent.is_symlink() for parent in [path, *path.parents] if parent != root.parent):
        raise ValueError(f"symlink is not permitted: {path}")
    return path.read_bytes()


def verify_runtime_overlays(destination: pathlib.Path, target_root: pathlib.Path) -> dict:
    """Reject unpinned bytes, extra Go files, and staged/unstaged runtime changes.

    Documentation and independent source snapshots may be new or modified; no
    runtime manifest, catalogue, module, or manifest-recorded path may be dirty.
    Changes elsewhere in the shared parent worktree are deliberately ignored.
    """
    manifest = json.loads(regular_bytes(destination, "LINEAGE_MANIFEST.json"))
    if manifest["source_commit"] != RETAINED_RUNTIME_COMMIT:
        raise ValueError("retained runtime must keep its original upstream commit")
    paths = set()
    for record in manifest["files"]:
        relative = record["path"]
        if relative in paths or relative.startswith("upstream/"):
            raise ValueError(f"invalid runtime manifest path: {relative}")
        paths.add(relative)
        content = regular_bytes(destination, relative)
        if len(content) != record["bytes"] or hashlib.sha256(content).hexdigest() != record["sha256"]:
            raise ValueError(f"dirty runtime overlay: {relative}; refresh never re-pins or overwrites it")
    actual_go = {p.relative_to(destination).as_posix() for p in destination.rglob("*.go")
                 if not p.relative_to(destination).as_posix().startswith("upstream/")}
    recorded_go = {r["path"] for r in manifest["files"] if r["kind"] in {"go", "local_go"}}
    if actual_go != recorded_go:
        raise ValueError("dirty runtime overlays: Go path set differs from retained manifest")
    protected = paths | {"LINEAGE_MANIFEST.json", "INTEGRATION_CATALOGUE.json",
                         "go.mod", "go.mod.upstream", "go.sum.upstream"}
    prefix = destination.relative_to(target_root)
    scoped = [str(prefix / safe_relative(p)) for p in sorted(protected)]
    if (run(target_root, "git", "diff", "--name-only", "--", *scoped)
            or run(target_root, "git", "diff", "--cached", "--name-only", "--", *scoped)):
        raise ValueError("dirty runtime overlays: commit their reviewed edits before source refresh")
    catalogue = json.loads(regular_bytes(destination, "INTEGRATION_CATALOGUE.json"))
    if {r["path"] for r in catalogue["files"]} != {r["path"] for r in manifest["files"] if r["kind"] == "go"}:
        raise ValueError("retained runtime catalogue path set differs from manifest")
    for record in catalogue["files"]:
        if hashlib.sha256(regular_bytes(destination, record["path"])).hexdigest() != record["sha256"]:
            raise ValueError(f"retained runtime catalogue hash mismatch: {record['path']}")
    return manifest


def source_snapshot_payload(source: pathlib.Path) -> tuple[dict[str, bytes], dict]:
    if run(source, "git", "status", "--porcelain", "--untracked-files=no"):
        raise ValueError("source repository has modified tracked files")
    commit = run(source, "git", "rev-parse", "HEAD")
    repository = run(source, "git", "remote", "get-url", "origin")
    if commit != DEFAULT_SOURCE_COMMIT:
        raise ValueError(f"source HEAD {commit} does not match pinned commit {DEFAULT_SOURCE_COMMIT}")
    if canonical_repository(repository) != canonical_repository(DEFAULT_SOURCE_REPOSITORY):
        raise ValueError(f"source repository {repository} does not match pinned repository {DEFAULT_SOURCE_REPOSITORY}")
    # Read pinned Git blobs, never checkout/filter-transformed or untracked bytes.
    tree = subprocess.check_output(["git", "ls-tree", "-rz", "--full-tree", commit], cwd=source)
    entries = {}
    for entry in tree.split(b"\0"):
        if entry:
            metadata, path = entry.split(b"\t", 1)
            mode, kind, oid = metadata.decode().split()
            entries[path.decode()] = (mode, kind, oid)
    go_files = {pathlib.Path(p) for p in entries if p.endswith(".go")}
    if len(go_files) != 553:
        raise ValueError(f"pinned current source must contain 553 Go files, found {len(go_files)}")
    if {p.as_posix() for p in tracked_go_files(source)} != {p.as_posix() for p in go_files}:
        raise ValueError("source index Go paths differ from pinned Git tree")
    assets = embedded_assets(source, go_files)
    licenses = {p for p in tracked_license_files(source) if p.as_posix() in entries}
    payload = {}
    kinds = {}
    selected = go_files | assets | licenses
    for relative in sorted(selected, key=lambda p: p.as_posix()):
        name = relative.as_posix()
        safe_relative(name)
        mode, kind, oid = entries[name]
        if kind != "blob" or mode not in {"100644", "100755"}:
            raise ValueError(f"source path is not a regular tracked blob: {name}")
        payload[name] = subprocess.check_output(["git", "cat-file", "blob", oid], cwd=source)
        kinds[name] = "go" if relative in go_files else "embed_asset" if relative in assets else "license"
    for name in ("go.mod", "go.sum"):
        if name in entries:
            payload[name + ".upstream"] = subprocess.check_output(
                ["git", "cat-file", "blob", entries[name][2]], cwd=source)
            kinds[name + ".upstream"] = "upstream_module"
    if "go.mod.upstream" not in payload:
        raise ValueError("source has no tracked original go.mod")
    payload["go.mod"] = b"module github.com/EchoCog/echollama\n\ngo 1.26.0\n\ntoolchain go1.26.11\n"
    kinds["go.mod"] = "module_boundary"
    records = [{"path": p, "kind": kinds[p], "bytes": len(payload[p]),
                "sha256": hashlib.sha256(payload[p]).hexdigest()} for p in sorted(payload)]
    manifest = {
        "schema_version": 1,
        "source_repository": DEFAULT_SOURCE_REPOSITORY,
        "source_commit": commit,
        "source_ref": f"{DEFAULT_SOURCE_REPOSITORY}@{commit}",
        "source_committed_at": run(source, "git", "show", "-s", "--format=%cI", commit),
        "snapshot_role": "immutable_raw_source_not_runtime",
        "default_trust_grade": "preserved_unreviewed",
        "activation_policy": "no_auto_activation",
        "go_file_count": len(go_files),
        "embed_asset_count": len(assets),
        "license_file_count": len(licenses),
        "go_paths_sha256": hashlib.sha256("".join(p + "\n" for p in sorted(p.as_posix() for p in go_files)).encode()).hexdigest(),
        "files_sha256": hashlib.sha256("".join(r["path"] + "\0" + r["sha256"] + "\n" for r in records).encode()).hexdigest(),
        "files": records,
    }
    payload[SOURCE_SNAPSHOT_MANIFEST] = (json.dumps(manifest, indent=2, sort_keys=True) + "\n").encode()
    return payload, manifest


def verify_snapshot_payload(snapshot: pathlib.Path, payload: dict[str, bytes]) -> None:
    actual = set()
    for path in snapshot.rglob("*"):
        if path.is_symlink():
            raise ValueError(f"modified source snapshot contains symlink: {path}")
        if path.is_file():
            actual.add(path.relative_to(snapshot).as_posix())
    if actual != set(payload):
        raise ValueError("existing source snapshot path set differs; refusing to overwrite or delete any work")
    for relative, expected in payload.items():
        if regular_bytes(snapshot, relative) != expected:
            raise ValueError(f"modified source snapshot: {relative}; refusing to overwrite")


def refresh_source_snapshot(source: pathlib.Path, destination: pathlib.Path, target_root: pathlib.Path) -> dict:
    runtime = verify_runtime_overlays(destination, target_root)
    payload, manifest = source_snapshot_payload(source)
    for path in ACTIVE_SOURCE_PATHS:
        if regular_bytes(destination, path) != payload[path]:
            raise ValueError(f"native driver differs from current source: {path}; explicit review required")
    snapshot = destination / "upstream"
    if snapshot.is_symlink():
        raise ValueError("source snapshot must not be a symlink")
    changed = not snapshot.exists()
    if changed:
        # No destructive replacement: publish only a new, fully verified tree.
        # Temporary cleanup is restricted to the directory created by this call.
        with tempfile.TemporaryDirectory(prefix="ecco9-source-") as temporary:
            staged = pathlib.Path(temporary) / "upstream"
            staged.mkdir()
            for relative, content in payload.items():
                path = staged / safe_relative(relative)
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_bytes(content)
            verify_snapshot_payload(staged, payload)
            verify_runtime_overlays(destination, target_root)
            # Move into a same-filesystem staging directory before atomic rename.
            staging = pathlib.Path(tempfile.mkdtemp(prefix=".source-snapshot-", dir=destination))
            try:
                shutil.copytree(staged, staging / "upstream")
                (staging / "upstream").rename(snapshot)
            finally:
                shutil.rmtree(staging)
    verify_snapshot_payload(snapshot, payload)
    verify_runtime_overlays(destination, target_root)
    return {
        "changed": changed,
        "source_commit": manifest["source_commit"],
        "source_go_file_count": manifest["go_file_count"],
        "source_embed_asset_count": manifest["embed_asset_count"],
        "source_license_file_count": manifest["license_file_count"],
        "source_manifest_sha256": hashlib.sha256(payload[SOURCE_SNAPSHOT_MANIFEST]).hexdigest(),
        "source_files_sha256": manifest["files_sha256"],
        "source_go_paths_sha256": manifest["go_paths_sha256"],
        "retained_runtime_source_commit": runtime["source_commit"],
        "retained_runtime_manifest_sha256": sha256(destination / "LINEAGE_MANIFEST.json"),
        "retained_runtime_catalogue_sha256": sha256(destination / "INTEGRATION_CATALOGUE.json"),
        "retained_runtime_record_count": len(runtime["files"]),
        "retained_runtime_go_file_count": sum(r["kind"] in {"go", "local_go"} for r in runtime["files"]),
        "verified_native_file_count": len(ACTIVE_SOURCE_PATHS),
        "destination": str(snapshot),
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    modes = parser.add_mutually_exclusive_group(required=True)
    modes.add_argument("--refresh-source-snapshot", action="store_true")
    modes.add_argument("--repin-local-patches", action="store_true")
    parser.add_argument("--source", type=pathlib.Path, help="clean checkout of the pinned current source")
    parser.add_argument("--destination", type=pathlib.Path, default=pathlib.Path("cognitive-core/ecco9"))
    parser.add_argument("--target-root", type=pathlib.Path, default=pathlib.Path("."))
    args = parser.parse_args()
    destination = args.destination.resolve()
    target_root = args.target_root.resolve()
    try:
        destination.relative_to(target_root)
        if destination == target_root or not destination.is_dir() or args.destination.is_symlink():
            raise ValueError("destination must be an existing integration directory inside target root")
        if args.refresh_source_snapshot:
            if args.source is None:
                parser.error("--refresh-source-snapshot requires --source")
            result = refresh_source_snapshot(args.source.resolve(), destination, target_root)
        else:
            if args.source is not None:
                parser.error("--source is only valid with --refresh-source-snapshot")
            result = repin_local_patches(destination, target_root)
        print(json.dumps(result, indent=2, sort_keys=True))
        return 0
    except (ValueError, OSError, KeyError, subprocess.CalledProcessError) as exc:
        parser.exit(1, f"source refresh refused: {exc}\n")


if __name__ == "__main__":
    sys.exit(main())
