# o9nn/ecco9 Cognitive-Core Lineage Snapshot

This directory preserves **all 553 tracked Go files** from
`https://github.com/o9nn/ecco9.git` at commit `1b22401ee8842fd1aa2769b688d588cd9f1f9ce9` plus 45 files required by Go embed
directives and 3 applicable license files. Every imported byte is
inventoried in `LINEAGE_MANIFEST.json` (manifest SHA-256 `275a0460d4c13f94f4909f9f8ac65d95d48b7511280ca0131bbe08b9bf2cc498`).

The snapshot is a **nested Go module and evidence boundary**, not a second
canonical runtime. Files retain their upstream package paths and are not copied
into root packages where duplicate symbols and stale APIs could silently replace
verified behavior. The root `core/cognitivecore` adapter activates only reviewed
reservoir, memory, emotion, and consciousness driver contracts. All other source
packages remain preserved for staged repair and promotion.

`INTEGRATION_CATALOGUE.json` assigns every preserved Go file an explicit
disposition. Current counts are `{"active_adapter_source": 6, "canonical_same_path_exact": 146, "preserved_divergent": 330, "preserved_unreviewed": 71}`;
catalogue SHA-256 is `c50f920be992ed73057ded062ce783e4c196eb5794831a5b9d5433ea127e9826`.

`go.mod.upstream` and `go.sum.upstream` preserve the upstream module graph. The
minimal local `go.mod` prevents unreviewed transitive dependencies from entering
the canonical security boundary merely because their source was preserved.

Regenerate from a clean source clone with:

```bash
git clone https://github.com/o9nn/ecco9.git /tmp/ecco9-source
git -C /tmp/ecco9-source checkout --detach 1b22401ee8842fd1aa2769b688d588cd9f1f9ce9
python3 scripts/sync_ecco9_cognitive_core.py /tmp/ecco9-source
```

## Local patches

78 preserved files carry local patches and 11 Go files were added locally (model-only character text: canned, templated, random-choice and fallback output replaced by model queries; no text when the model is absent). Each patched record in `LINEAGE_MANIFEST.json` keeps its upstream bytes and SHA-256 in `upstream_bytes`/`upstream_sha256`. Re-pin after editing with:

```bash
python3 scripts/sync_ecco9_cognitive_core.py --repin-local-patches
```
