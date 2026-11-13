"""Isolated safety tests; no edits or dependencies in either real repository.

Run from the root: python3 -m unittest discover -s cognitive-core/ecco9
"""

import hashlib
import importlib.util
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

SCRIPT = Path(__file__).resolve().parents[2] / "scripts/sync_ecco9_cognitive_core.py"
spec = importlib.util.spec_from_file_location("source_importer", SCRIPT)
importer = importlib.util.module_from_spec(spec)
spec.loader.exec_module(importer)


def git(root, *arguments):
    return subprocess.check_output(["git", *arguments], cwd=root, text=True, stderr=subprocess.DEVNULL).strip()


def initialize(root):
    git(root, "init", "-q")
    git(root, "config", "user.name", "Source snapshot safety test")
    git(root, "config", "user.email", "snapshot-tests@example.invalid")


def inventory(root):
    return {p.relative_to(root).as_posix(): p.read_bytes() for p in root.rglob("*") if p.is_file()}


class SourceSnapshotSafetyTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.shared = tempfile.TemporaryDirectory()
        cls.source = Path(cls.shared.name) / "source"
        cls.source.mkdir()
        initialize(cls.source)
        git(cls.source, "remote", "add", "origin", importer.DEFAULT_SOURCE_REPOSITORY)
        cls.native = sorted(importer.ACTIVE_SOURCE_PATHS)
        for name in cls.native + [f"file_{i:04d}.go" for i in range(547)]:
            path = cls.source / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(b"package fixture\n")
        (cls.source / "file_0000.go").write_bytes(b'package fixture\nimport _ "embed"\n//go:embed asset.txt\nvar data string\n')
        (cls.source / "asset.txt").write_bytes(b"exact embedded bytes\r\n")
        (cls.source / "LICENSE").write_bytes(b"fixture license\n")
        (cls.source / "go.mod").write_bytes(b"module fixture\n\ngo 1.25.0\n")
        (cls.source / "go.sum").write_bytes(b"fixture module graph\n")
        git(cls.source, "add", ".")
        git(cls.source, "commit", "-qm", "fixture source")
        cls.commit = git(cls.source, "rev-parse", "HEAD")

    @classmethod
    def tearDownClass(cls):
        cls.shared.cleanup()

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name) / "target"
        self.root.mkdir()
        initialize(self.root)
        self.destination = self.root / "cognitive-core/ecco9"
        self.destination.mkdir(parents=True)
        records = []
        for name in self.native + ["removed_by_current_upstream.go", "overlaid.go", "locally_added.go"]:
            content = b"package fixture\n" if name in self.native else b"package retained\n// Preserve runtime policy, never merge.\n"
            path = self.destination / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(content)
            kind = "local_go" if name == "locally_added.go" else "go"
            record = {"path": name, "kind": kind, "bytes": len(content), "sha256": hashlib.sha256(content).hexdigest()}
            if name in {"overlaid.go", "locally_added.go"}:
                record["local_patch"] = "retained policy"
            records.append(record)
        manifest = {"schema_version": 1, "source_commit": importer.RETAINED_RUNTIME_COMMIT,
                    "go_file_count": 8, "embed_asset_count": 0, "license_file_count": 0,
                    "local_addition_count": 1, "files": records}
        (self.destination / "LINEAGE_MANIFEST.json").write_text(json.dumps(manifest))
        catalogue = {"files": [dict(r, disposition="active_adapter_source" if r["path"] in self.native else "preserved_divergent")
                               for r in records if r["kind"] == "go"]}
        (self.destination / "INTEGRATION_CATALOGUE.json").write_text(json.dumps(catalogue))
        for name in ("go.mod", "go.mod.upstream", "go.sum.upstream"):
            (self.destination / name).write_bytes(b"retained runtime module bytes\n")
        git(self.root, "add", ".")
        git(self.root, "commit", "-qm", "retained runtime fixture")
        self.pin = patch.object(importer, "DEFAULT_SOURCE_COMMIT", self.commit)
        self.pin.start()
        self.addCleanup(self.pin.stop)
        self.original = inventory(self.destination)

    def refresh(self):
        return importer.refresh_source_snapshot(self.source, self.destination, self.root)

    def assert_retained_unchanged(self):
        for path, content in self.original.items():
            self.assertEqual((self.destination / path).read_bytes(), content, path)

    def test_exact_complete_source_refresh_is_idempotent_and_retains_removed_and_added_runtime(self):
        result = self.refresh()
        self.assertTrue(result["changed"])
        self.assertEqual(result["source_go_file_count"], 553)
        self.assertEqual(result["source_embed_asset_count"], 1)
        self.assertEqual(result["verified_native_file_count"], 6)
        snapshot = self.destination / "upstream"
        for name in git(self.source, "ls-files", "*.go").splitlines():
            self.assertEqual((snapshot / name).read_bytes(), (self.source / name).read_bytes())
        self.assertFalse((snapshot / "removed_by_current_upstream.go").exists())
        self.assertFalse((snapshot / "locally_added.go").exists())
        self.assertEqual((snapshot / "asset.txt").read_bytes(), b"exact embedded bytes\r\n")
        self.assertEqual((snapshot / "go.mod.upstream").read_bytes(), (self.source / "go.mod").read_bytes())
        before = inventory(self.destination)
        mtimes = {p: (self.destination / p).stat().st_mtime_ns for p in before}
        self.assertFalse(self.refresh()["changed"])
        self.assertEqual(inventory(self.destination), before)
        self.assertEqual(mtimes, {p: (self.destination / p).stat().st_mtime_ns for p in before})
        self.assert_retained_unchanged()

    def test_refuses_dirty_runtime_without_writing_snapshot(self):
        path = self.destination / "overlaid.go"
        path.write_bytes(b"unreviewed runtime change\n")
        with self.assertRaisesRegex(ValueError, "dirty runtime overlay"):
            self.refresh()
        self.assertEqual(path.read_bytes(), b"unreviewed runtime change\n")
        self.assertFalse((self.destination / "upstream").exists())

    def test_refuses_new_unmanifested_runtime_go(self):
        (self.destination / "untracked.go").write_bytes(b"package unreviewed\n")
        with self.assertRaisesRegex(ValueError, "Go path set differs"):
            self.refresh()
        self.assert_retained_unchanged()
        self.assertFalse((self.destination / "upstream").exists())

    def test_refuses_staged_change_even_when_working_bytes_restored(self):
        path = self.destination / "overlaid.go"
        path.write_bytes(b"staged dirty bytes\n")
        git(self.root, "add", str(path))
        path.write_bytes(self.original["overlaid.go"])
        with self.assertRaisesRegex(ValueError, "dirty runtime overlays"):
            self.refresh()
        self.assert_retained_unchanged()
        self.assertFalse((self.destination / "upstream").exists())

    def test_refuses_dirty_raw_bytes_and_never_overwrites_them(self):
        self.refresh()
        path = self.destination / "upstream" / self.native[0]
        path.write_bytes(b"uncommitted source work\n")
        with self.assertRaisesRegex(ValueError, "modified source snapshot"):
            self.refresh()
        self.assertEqual(path.read_bytes(), b"uncommitted source work\n")
        self.assert_retained_unchanged()

    def test_refuses_added_raw_file_and_never_deletes_it(self):
        self.refresh()
        path = self.destination / "upstream" / "valuable.txt"
        path.write_bytes(b"retain tracked or untracked work\n")
        with self.assertRaisesRegex(ValueError, "path set differs"):
            self.refresh()
        self.assertEqual(path.read_bytes(), b"retain tracked or untracked work\n")
        self.assert_retained_unchanged()

    def test_refuses_wrong_source_pin(self):
        with patch.object(importer, "DEFAULT_SOURCE_COMMIT", "0" * 40):
            with self.assertRaisesRegex(ValueError, "does not match pinned commit"):
                self.refresh()
        self.assertFalse((self.destination / "upstream").exists())
        self.assert_retained_unchanged()

    def test_manual_runtime_repin_never_absorbs_raw_source_or_rewrites_its_pin(self):
        self.refresh()
        raw_before = inventory(self.destination / "upstream")
        lineage = self.root / "core/cognitivecore/lineage.go"
        lineage.parent.mkdir(parents=True)
        source_pin = 'SourceSnapshotManifestSHA256 = "' + 'a' * 64 + '"'
        lineage.write_text('package cognitivecore\nconst (\n\t' + source_pin + '\n\tManifestSHA256 = "'
                           + 'b' * 64 + '"\n\tCatalogueSHA256 = "' + 'c' * 64 + '"\n)\n')
        importer.repin_local_patches(self.destination, self.root)
        self.assertIn(source_pin, lineage.read_text())
        self.assertEqual(inventory(self.destination / "upstream"), raw_before)
        manifest = json.loads((self.destination / "LINEAGE_MANIFEST.json").read_text())
        self.assertFalse(any(r["path"].startswith("upstream/") for r in manifest["files"]))
        self.assertEqual(manifest["source_commit"], importer.RETAINED_RUNTIME_COMMIT)
        self.assertEqual(manifest["local_addition_count"], 1)

    def test_refuses_native_divergence_even_if_runtime_repin_was_committed(self):
        path = self.destination / self.native[0]
        path.write_bytes(b"package different_native\n")
        for name in ("LINEAGE_MANIFEST.json", "INTEGRATION_CATALOGUE.json"):
            metadata = json.loads((self.destination / name).read_text())
            for record in metadata["files"]:
                if record["path"] == self.native[0]:
                    record["bytes"] = path.stat().st_size
                    record["sha256"] = hashlib.sha256(path.read_bytes()).hexdigest()
            (self.destination / name).write_text(json.dumps(metadata))
        git(self.root, "add", ".")
        git(self.root, "commit", "-qm", "different reviewed native fixture")
        with self.assertRaisesRegex(ValueError, "native driver differs"):
            self.refresh()
        self.assertFalse((self.destination / "upstream").exists())

    def test_refuses_symlink_snapshot(self):
        (self.destination / "upstream").symlink_to(self.source, target_is_directory=True)
        with self.assertRaisesRegex(ValueError, "must not be a symlink"):
            self.refresh()
        self.assert_retained_unchanged()

    def test_legacy_positional_import_is_refused_without_edits(self):
        result = subprocess.run(["python3", str(SCRIPT), str(self.source), "--target-root", str(self.root),
                                 "--destination", str(self.destination)], capture_output=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(inventory(self.destination), self.original)


if __name__ == "__main__":
    unittest.main()
