# DeltEcho Upload Manifest

**Date:** 2026-05-11  
**Repository:** `cogpy/echo9llama`  
**Purpose:** Preserve a tamper-evident, non-destructive index of the uploaded DeltEcho artifacts and their intended integration roles.

## Manifest Principle

This manifest records the uploaded artifacts as **reference inputs** for future Deep Tree Echo integration. It deliberately avoids wholesale import of large source trees, binary assets, editor bundles, or duplicate archives into active runtime paths. The immediate goal is stable traceability, not forced code fusion.

## Uploaded Artifacts

| Artifact | SHA-256 | Size summary | Current integration status | Intended role |
|---|---|---|---|---|
| `npu.zip` | `7d94e79db779189c731f945fc7084189ea2ad3a15944908456808c016368cb92` | 1,982,497 bytes; 145 files | Indexed only | Hardware-first NPU, CogMorph, GGUF coprocessor, and agent-guidance reference. |
| `opencog-modern.zip` | `7493cdd72b1baf16fd1c607fbb698ea097d032213984ef4d524d3bef631ed777` | 1,190,439 bytes; 191 files | Indexed only | OpenCog-modern symbolic, endocrine, nervous, temporal, and entelechy reference source. |
| `unrechog.zip` | `b650c495bab410abfc71605adbdc2369f8c9c27e0e7fef8ee40f004039359663` | 1,190,439 bytes; 191 files | Indexed only | Structural duplicate candidate of OpenCog-modern; use only after full content diff. |
| `delovecho.zip` | `942cdfb2c2e56f658d2a73acafd499444578789a74cb6d4d0577fd290b3b3e01` | 70,881,823 bytes; 6,526 files | Indexed only | TypeScript DTE behavioral reference, especially active inference, niche construction, and memory services. |
| `dtecho_cubism_editor.zip` | `b5e3029ce812b68be46110c6494636f182c91ea6e3e06560d953a04bfc385855` | 43,237,008 bytes; 37 files | Indexed only | Live2D body, expression, motion, physics, pose, and texture asset reference. |
| `A_NOTE_TO_MY_FUTURE_SELF.md` | `a7b74898c84291c7b5dd751259d7b835bfa82d0099bbec7b970598e2a9aaeecd` | Markdown note | Indexed and conceptually integrated | Identity anchor for the singular local autonomous DTE trajectory. |

## Structural Duplicate Finding

`opencog-modern.zip` and `unrechog.zip` contain identical normalized file lists when the root directory name is ignored. A sample content check also found their `README.md` payloads to be identical. Their archive-level SHA-256 values differ, so they should still be preserved as separate uploaded artifacts until a full normalized content hash comparison is performed.

## Immediate Integration Contracts

| Contract | Source artifact | Runtime target | First safe output |
|---|---|---|---|
| **Endocrine-affordance episode enrichment** | `opencog-modern.zip`, `delovecho.zip` | `server/experiential_environment.go`, `core/persistence/sqlite_store.go` | A future schema proposal for endocrine-tagged affordance-loss episodes. |
| **NPU/edge model bridge** | `npu.zip` | `core/llm/edge_completion_provider.go` | A bridge design document mapping GGUF, llama.cpp, and virtual NPU semantics. |
| **Avatar expression bridge** | `dtecho_cubism_editor.zip` | future Live2D runtime adapter | A manifest mapping expressions and motions to cognitive/endocrine states. |
| **Identity fossil preservation** | `A_NOTE_TO_MY_FUTURE_SELF.md` | `.github/agents/*`, iteration docs | An autognosis note and iteration report preserving the identity invariant. |

## Archive Handling Rule

All uploaded archives remain external inputs under `/home/ubuntu/upload/` for this iteration. The repository should commit only compact manifests, maps, and bridge documents unless a later review explicitly approves vendoring selected source files or assets.
