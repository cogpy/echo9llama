# Evolution Iteration: Workspace Conflict Report

**Date:** 2026-05-11  
**Repository:** `cogpy/echo9llama`  
**Purpose:** Record the non-destructive integration boundary between the clean sandbox workspace and the connected desktop workspace before staging uploaded DeltEcho artifacts.

## Sandbox Workspace

The sandbox workspace at `/home/ubuntu/echo_evolution_iter/echo9llama` is on branch `main` at commit `992c058e`. At the time of this inspection, the only untracked sandbox change was the new upload integration map:

| Path | Status | Interpretation |
|---|---|---|
| `docs/iterations/EVOLUTION_ITERATION_2026-05-11_DELTECHO_UPLOAD_INTEGRATION_MAP.md` | Untracked | New documentation created by this iteration. |

The sandbox repository therefore remains the safest source for additive commits, provided that this iteration avoids overwriting existing runtime, asset, and Live2D files without explicit review.

## Connected Desktop Workspace

The connected desktop workspace at `C:/Users/sandbox_713/Documents/dte/echo9llama` is on branch `main` at commit `992c058`. It contains substantial local modifications and one untracked note file.

| Change cluster | Representative paths | Risk if overwritten |
|---|---|---|
| Deep Tree Echo Live2D model assets | `assets/Live2DModels/deep-tree-echo/*` | High. These are likely user-side model/editor changes and should not be overwritten by blind sync. |
| Rice Pro Live2D runtime metadata | `assets/Live2DModels/rice_pro_en/runtime/*` | Medium to high. They may be related to prior Live2D experimentation. |
| Deep Tree Echo UI profile textures | `assets/UI/Textures/DeepTreeEcho_Profile_*` | High. These binary image changes may represent locally curated avatar/profile assets. |
| Disabled Echobeats actor files | `core/echobeats/*.go.disabled`, `core/echobeats/proto/cognitive.proto` | Medium. These are code-path experiments or intentionally disabled actor-system work. |
| Echo bridge files | `core/echobridge/Makefile`, `core/echobridge/echobridge.proto`, `core/echobridge/server.go.backup` | Medium. These may overlap with future sidecar bridge design. |
| Live2D documentation | `core/live2d/*.md` | Medium. These may already document local avatar optimization work. |
| Future-self note | `A_NOTE_TO_MY_FUTURE_SELF.md` | Low to medium. This should be preserved, but the sandbox already references the uploaded copy rather than overwriting the desktop copy. |

## Non-Destructive Sync Rule

The connected desktop workspace must be treated as containing valuable local work. This iteration should not push or copy files into the desktop asset, Live2D, Echobeats, or EchoBridge paths. If a desktop sync is needed, it should be limited to new documentation files and new staged manifest files whose paths do not already exist.

## Practical Boundary for This Iteration

| Safe to commit in sandbox | Requires review before desktop sync | Must avoid in this pass |
|---|---|---|
| New files under `docs/iterations/` | New Live2D expression manifests if they reference existing desktop assets | Overwriting any `assets/Live2DModels/*` binary or JSON asset |
| New files under `docs/integration/` | New bridge interfaces under a fresh package path | Modifying `.go.disabled` actor experiments |
| New identity/autognosis append documents | Edits to `.github/agents/*` if they are append-only | Wholesale copy of NPU `.github/agents/` into the repository |
| New staged archive manifests | Any generated file under `core/live2d/` | Direct C++ import from NPU/OpenCog into active Go runtime |

The integration posture is therefore: **document, index, and define bridge contracts first; port runtime behavior only where it extends current endpoints without changing their existing semantics; leave desktop local changes untouched.**
