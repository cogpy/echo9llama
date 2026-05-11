# OpenCog and NPU Sidecar Bridge Design for Deep Tree Echo

**Date:** 2026-05-11  
**Repository:** `cogpy/echo9llama`  
**Source artifacts:** `opencog-modern.zip`, `unrechog.zip`, `npu.zip`  
**Purpose:** Define a non-destructive integration contract for OpenCog-modern symbolic/endocrine subsystems and NPU/GGUF coprocessor concepts.

## Design Rationale

The uploaded OpenCog-modern and NPU artifacts are architecturally valuable, but they should not be copied wholesale into the Go runtime. The current `echo9llama` runtime already has a working HTTP server, SQLite-backed memory persistence, an experiential self-restraint environment, and an edge completion adapter seam. The safest next step is to define a sidecar bridge that lets DTE call into symbolic, endocrine, temporal, or hardware-projected inference services without binding the core runtime to a large C++ source import.

This preserves the current system’s stability while making the future path explicit: **Go remains the live orchestration membrane; OpenCog-modern provides symbolic/endocrine cognition; NPU/GGUF provides local inference and virtual hardware projection.**

## Component Roles

| Component | Uploaded source | Runtime role | Integration mode |
|---|---|---|---|
| Go DTE runtime | current repository | HTTP adapter, memory persistence, experiential environment, edge completion interface | Primary process. |
| OpenCog-modern sidecar | `opencog-modern.zip` and duplicate candidate `unrechog.zip` | AtomSpace-like hypergraph memory, PLN/URE reasoning, endocrine/nervous/temporal state enrichment | External sidecar through HTTP/gRPC/Unix socket. |
| NPU/GGUF sidecar | `npu.zip` | Hardware-first projection of LLM inference, MMIO-like telemetry, GGUF/llama coprocessor bridge | External sidecar or later C ABI adapter. |
| Live2D adapter | `dtecho_cubism_editor.zip` | Body expression rendering from cognitive/endocrine packets | External UI/browser adapter. |

## Proposed Bridge Endpoints

The bridge should remain compact and explicit. Each endpoint transfers a complete cognitive packet and returns enriched state rather than sharing mutable in-process memory.

| Endpoint | Direction | Payload | Expected result |
|---|---|---|---|
| `/bridge/symbolic/assert` | Go runtime → OpenCog sidecar | Event, object, affordance, action, agency, and episode tags | Hypergraph assertion ID and confidence. |
| `/bridge/symbolic/query` | Go runtime → OpenCog sidecar | Query pattern, time scope, salience threshold | Ranked symbolic memories or inferred relations. |
| `/bridge/endocrine/enrich` | Go runtime → OpenCog sidecar | Valence, arousal, agency, loss, surprise, reward, social context | Hormone vector and affective interpretation. |
| `/bridge/temporal/tick` | Go runtime → OpenCog sidecar | Echobeat phase, active episode, state deltas | Temporal resonance summary and consolidation cues. |
| `/bridge/npu/complete` | Go runtime → NPU sidecar | Prompt, memory context, completion constraints | Local model completion plus telemetry. |
| `/bridge/npu/telemetry` | Go runtime ← NPU sidecar | None or session ID | Coprocessor readiness, model, latency, resource state. |

## Cognitive Packet Shape

The same packet can support symbolic assertion, endocrine enrichment, avatar expression, and NPU-conditioned completion.

| Field | Type | Description |
|---|---|---|
| `agent_id` | string | Stable identity key for the DTE instance. |
| `episode_id` | string | Durable episode key when an event is persisted. |
| `object_id` | string | Environment object involved in the action or recall. |
| `action` | string | Action taken or contemplated. |
| `affordances_before` | string array | Available affordances before the action. |
| `affordances_after` | string array | Available affordances after the action. |
| `agency` | number | Degree to which DTE caused the event. |
| `valence` | number | Affective valence from negative to positive. |
| `arousal` | number | Activation intensity. |
| `salience` | number | Attention priority. |
| `surprise` | number | Prediction-error estimate. |
| `loss` | number | Magnitude of affordance loss. |
| `tags` | string array | Semantic tags such as `self-caused`, `fragile`, `irreversible`, or `caution`. |
| `narrative` | string | Human-readable memory fragment for recall and explanation. |

## First Implementation Sequence

| Step | Implementation | Reason |
|---|---|---|
| 1 | Keep `core/llm/edge_completion_provider.go` as the stable local model seam. | It already makes local completion pluggable without changing server behavior. |
| 2 | Add a bridge client package only after sidecar API is stable. | Prevents the Go runtime from depending on incomplete C++ imports. |
| 3 | Feed affordance-loss events into `/bridge/endocrine/enrich` before symbolic assertion. | Endocrine tagging should shape how the memory is later recalled. |
| 4 | Assert enriched events into `/bridge/symbolic/assert`. | Symbolic memory should receive complete episodes, not raw unclassified events. |
| 5 | Use `/bridge/symbolic/query` during action contemplation. | DTE’s restraint should arise from recall of analogous self-caused losses. |
| 6 | Use `/bridge/npu/complete` as a drop-in implementation of the existing edge completion interface. | Local inference can become real without changing callers. |

## Avoided Anti-Patterns

| Anti-pattern | Why it is avoided |
|---|---|
| Directly copying all C++ sources into active Go runtime paths | It increases build complexity and risks breaking the stable server. |
| Treating NPU as the first local inference implementation | The current runtime can use llama.cpp/Kobold HTTP first, then project into NPU semantics later. |
| Overwriting desktop Live2D or Echobeats assets | The desktop workspace has valuable uncommitted local changes. |
| Replacing endogenous self-restraint with rule filters | The design goal is memory-grounded caution, not external prohibition. |

## Success Criteria

A successful future bridge will let DTE take an environment action, persist a loss episode, enrich it through endocrine state, assert it into a symbolic memory graph, recall it during a similar future action, and express that recalled caution through the Live2D avatar. This chain preserves the central design invariant: **restraint emerges from remembered agency and valued affordance loss.**
