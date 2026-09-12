# o9nn/ecco9 Cognitive-Core Lineage Snapshot

This directory preserves **all 553 tracked Go files** from
`https://github.com/o9nn/ecco9.git` at commit `1b22401ee8842fd1aa2769b688d588cd9f1f9ce9` plus 45 files required by Go embed
directives and 3 applicable license files. Every imported byte is
inventoried in `LINEAGE_MANIFEST.json` (manifest SHA-256 `b382e99916a2364eb3945770835d3f907d42bbf2f548a7879a2e57c32590a5b7`).

The snapshot is a **nested Go module and evidence boundary**, not a second
canonical runtime. Files retain their upstream package paths and are not copied
into root packages where duplicate symbols and stale APIs could silently replace
verified behavior. The root `core/cognitivecore` adapter activates only reviewed
reservoir, memory, emotion, and consciousness driver contracts. All other source
packages remain preserved for staged repair and promotion.

`INTEGRATION_CATALOGUE.json` assigns every preserved Go file an explicit
disposition. Current counts are `{"active_adapter_source": 6, "canonical_same_path_exact": 149, "preserved_divergent": 327, "preserved_unreviewed": 71}`;
catalogue SHA-256 is `3456bd87a074478b0e60485ec891a1d8d91a443f1aca504840f678dfc8b69b63`.

`go.mod.upstream` and `go.sum.upstream` preserve the upstream module graph. The
minimal local `go.mod` prevents unreviewed transitive dependencies from entering
the canonical security boundary merely because their source was preserved.

Regenerate from a clean source clone with:

```bash
git clone https://github.com/o9nn/ecco9.git /tmp/ecco9-source
git -C /tmp/ecco9-source checkout --detach 1b22401ee8842fd1aa2769b688d588cd9f1f9ce9
python3 scripts/sync_ecco9_cognitive_core.py /tmp/ecco9-source
```
