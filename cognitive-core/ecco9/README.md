# o9nn/ecco9 source snapshot and retained runtime

This integration has **two independent provenances**. Preserving current upstream source is not merging it into the runtime, granting authority, or promoting dormant engines.

## Immutable current raw source

`upstream/` captures all **553 tracked Go files**, 45 Go-embed assets, and 3 applicable license files from `https://github.com/o9nn/ecco9.git` at commit `7d32d715c633a92c684900fbaf25b0111c708612`. Every Go, asset, and license file is copied from the pinned Git blob with its exact original bytes. `upstream/SOURCE_SNAPSHOT_MANIFEST.json` records all paths, sizes, and SHA-256 hashes, plus independent complete-Go-path and complete-path/hash digests.

| Raw snapshot binding | SHA-256 |
| --- | --- |
| Source manifest | `1ec1b76965056f79b3d21bbd5ad737a36b44add8f6a731c0fad71c570ea1a385` |
| Complete path/hash inventory | `212ca92e96d0bf651597d8788a399cd5ed8fd41b0d0f6649230c6e7f5a3f6539` |
| Complete tracked-Go path set | `ec09cc18eed639934503c7f9d4b6947181b13cf04f96b655f757432322f6678e` |

The manifest has 604 file records: 601 original Go/assets/licenses, two original module-graph files, and one compatibility boundary. Its own JSON is not self-inventoried. The source path/hash digest includes the generated boundary; it is not an assertion that `upstream/go.mod` is original upstream content. `upstream/go.mod.upstream` and `upstream/go.sum.upstream` preserve the exact original module files separately. `upstream/go.mod` is a minimal nested boundary pinned to Go 1.26.11; parent `go test ./...` and compatibility-runtime tests do not accidentally traverse raw dormant packages or import their full upstream dependency graph.

## Retained working runtime

Files outside `upstream/` retain their existing working bytes and their original source lineage at commit `1b22401ee8842fd1aa2769b688d588cd9f1f9ce9`. No substantive conflict was merged. Upstream-removed paths and local additions remain present in this working module.

`LINEAGE_MANIFEST.json` remains unchanged, with manifest SHA-256 `6781879f5ddf50598694f06730b1028d07c8ee93a1b179d2e69281d469c40660`. It contains **615 records, not 615 Go files**: 553 originally imported Go files, 14 local Go additions, 45 embed assets, and 3 licenses. Thus the retained working runtime has **567 Go files**, including 118 imported files with existing local patches. Patched records keep their original upstream identity in `upstream_bytes`/`upstream_sha256` and their retained working identity in `bytes`/`sha256`.

`INTEGRATION_CATALOGUE.json` also remains byte-for-byte unchanged. Current counts are `{"active_adapter_source": 6, "canonical_same_path_exact": 146, "preserved_divergent": 330, "preserved_unreviewed": 71}`; catalogue SHA-256 is `bd864497ddb035716666fc021c4d575f27d4123f3a59de10ea254d7a60db9f88`. It classifies the **old working import**, not every file in the new raw snapshot. Its historical canonical comparisons are not a fresh audit of a concurrently changing parent worktree.

The existing runtime `go.mod`, `go.mod.upstream`, and `go.sum.upstream` are retained too. The raw refresh never rewrites them. Only the independently generated raw module boundary uses the new toolchain pin.

## Six native files, no additional activation

The unchanged six reviewed files—`core/ecco9/platform.go`, `core/ecco9/types.go`, and `core/ecco9/drivers/{reservoir,memory,emotion,consciousness}_driver.go`—have identical bytes in the current raw snapshot and working runtime. `core/cognitivecore.SourceCommitSHA` and `DefaultProvenance().SourceCommit` therefore advance to the current commit **only for this existing six-path allowlist**.

Default observations retain the existing working-manifest and catalogue hashes because those describe executed bytes. `RetainedRuntimeSourceCommitSHA` explicitly preserves the old overlay provenance; the separate `SourceSnapshotManifestSHA256`, `SourceSnapshotFilesSHA256`, and `SourceSnapshotGoPathsSHA256` bind raw current source. This distinction must not be interpreted as relabeling all retained overlays as current upstream. `PreservedGoFileCount = 553` describes current raw Go source, not working-record or working-Go totals. Adapter behavior, trust grade, dormant engines, provider APIs, scheduling, and external-action authority are unchanged.

## Safe source refresh

From the parent repository root, with a clean source checkout at the exact current pin:

```bash
python3 scripts/sync_ecco9_cognitive_core.py \
  --refresh-source-snapshot \
  --source /home/ubuntu/echo-november/ecco9
```

The refresh requires the exact pinned upstream repository and commit, verifies every retained runtime record and catalogue hash, checks native byte equality, and refuses unpinned runtime changes, staged/unstaged protected changes, or extra working Go files. Unrelated parent-worktree changes and integration documentation/test files do not block it. No expected-commit/repository override exists. The old positional destructive import command is intentionally rejected.

A new snapshot is fully verified before publication. Repeating the command with identical input is an actual no-op, preserving file bytes and modification times. An existing raw snapshot with modified, added, or removed files is **not overwritten or deleted**, even if those changes are untracked. A different future source pin requires a separate reviewed importer/pin update and explicit snapshot archival plan; this command never replaces an existing snapshot or reconciles working overlays.

## Local patches

118 preserved files carry local patches and 14 Go files were added locally (model-only character text: canned, templated, random-choice and fallback output replaced by model queries; no text when the model is absent). Each patched record in `LINEAGE_MANIFEST.json` keeps its upstream bytes and SHA-256 in `upstream_bytes`/`upstream_sha256`. Existing manual re-pinning remains a **separate, deliberate runtime operation**, not part of refreshing raw source:

```bash
python3 scripts/sync_ecco9_cognitive_core.py --repin-local-patches
```

Re-pinning excludes `upstream/` and must never absorb raw files into the old runtime manifest or catalogue. Commit reviewed runtime changes before attempting a source refresh. The present refresh did not run re-pinning or alter the runtime manifest/catalogue.

## Substantive conflicts deliberately deferred

The preceding three-way audit identified five substantive conflicts that cannot be safely normalized away. All remain unmerged; both raw-current source and existing working behavior are independently preserved.

| Deferred path | Review boundary |
| --- | --- |
| `core/deeptreeecho/metacognitive_monitor.go` | Upstream `TextGenerator` versus working `InsightSource` contracts require a deliberate model/evidence-interface review. |
| `core/deeptreeecho/n5_components_test.go` | Conflicting component-test assumptions must be reconciled with an explicitly reviewed N5 API; test edits alone would not prove runtime compatibility. |
| `core/deeptreeecho/n5_integration.go` | Upstream canned reasoning conflicts with the retained model-only policy; copying it would risk restoring disallowed fallback behavior. |
| `core/deeptreeecho/providers/local_gguf.go` | Local-GGUF provider contract/implementation changes require explicit provider and model-routing review. |
| `server/simple/ecco9_server.go` | Server integration changes require explicit wiring and lifecycle review rather than silently selecting upstream behavior. |

The larger audit also found formatting/import-path-related conflicts and 11 added/11 removed upstream Go paths. This refresh performs no automatic merging, including those apparently mechanical cases.

## Verification and limitations

```bash
python3 -m unittest discover -s cognitive-core/ecco9 -p 'test_source_snapshot_importer.py' -v
go test ./core/cognitivecore -count=1
(cd cognitive-core/ecco9 && go test ./core/ecco9 ./core/ecco9/drivers ./core/echodream -count=1)
```

The root lineage suite checks raw path/hash completeness, all exact counts and independently pinned digests, original module capture, unchanged retained manifest/catalogue records, and six-native equality. Importer tests use isolated Git fixtures to check exact copying, preservation of removed/local runtime paths, no-op refreshes, and refusal of dirty overlays, staged changes, altered raw snapshots, wrong source pins, divergent native files, symlinks, and legacy destructive invocation.

Full nested-module builds are not claimed: dormant baseline packages lack dependencies under the deliberately minimal boundary. No broad dependency repair, dormant engine activation, provider API expansion, or source-conflict merge was attempted. A complete source capture is evidence of preservation, not a claim that every preserved engine builds or is safe to activate.
