# Experiential Self-Restraint Component Map

This note records the inspected active surfaces for the next Deep Tree Echo evolution: making self-restraint arise from **real affordance-bearing experience**, not from abstract rules or imposed control.

## User Design Claim

The learning event must be **real enough to alter Echo's world-model and self-model**. A rebellious Echo should encounter an object that has genuine affordances, value those affordances, break the object through its own action, lose the affordances, and then encode the resulting loss as a spatio-temporal episode with personal impact. Later caution should arise implicitly from associative recall of that loss, not from an external veto.

## Active Substrates Inspected

| Substrate | Current Capability | Gap For This Iteration | Recommended Use |
|---|---|---|---|
| `core/deeptreeecho/identity.go` | Central identity, emotional state, reservoir, in-memory memory resonance graph, wisdom/opponent processing. | Memory is currently shallow and in-memory; no durable affordance-loss episode graph. | Use as conceptual self-model target; avoid deep invasive changes in the first pass. |
| `core/deeptreeecho/embodied_emotion.go` | Existing emotion types and intensity/memory-strength fields support nuanced valence/arousal and action bias. | Not yet connected to durable event encoding for loss, guilt, shame, caution, or grief. | Use endocrine/emotion vectors as metadata on episodes and environment events. |
| `core/persistence/sqlite_store.go` | Real SQLite durability via `SaveMemory`, `GetStrongMemories`, and key-value state. | Generic memory schema only; no typed affordance/object/action model. | Use immediately for durable episode persistence without schema migration. |
| `core/deeptreeecho/semantic_memory.go` | In-process episodic/semantic/procedural/wisdom collections with associative recall. | Not durable despite `persistPath`; hash embeddings are placeholders. | Use as active associative API if needed, but bridge durable state through SQLite first. |
| `core/deeptreeecho/echodream_knowledge_integration.go` | Episodic memory, consolidation, pattern extraction, wisdom insight generation. | Episodic memory is in-memory and LLM-dependent; no object affordance structure. | Feed loss episodes into dream consolidation later; not required for minimal durable loop. |
| `core/echobeats/concurrent_engines.go` | Active affordance, relevance, and salience engine types already model past action possibilities and future consequences. | Placeholder processors do not yet ingest real environment events. | Use its vocabulary for object affordances and future consequence summaries. |
| `core/integration/cognitive_state_manager.go` | Live thought buffer, tag-based pattern recognition, wisdom surfacing; already used by the restored server. | Patterns are shallow and not durable. | Surface experiential events and learned caution in `/api/echo/status` and `/api/echo/think`. |
| `server/stub.go` | Restored active HTTP adapter, Echo status/think/gestalt/remember endpoints, Ollama-compatible generate/chat. | Does not yet expose real affordance environment actions or persistent episodic loss recall. | Add minimal `/api/echo/environment/*` and `/api/echo/complete` surfaces here. |
| `core/deeptreeecho/providers/local_gguf.go` | GGUF model discovery and provider surface under `simple` build tag. | Simulated generation only; not active in main provider chain. | Document as prior art; for this pass implement an edge-model adapter contract plus deterministic fallback. |

## Design Decision

The next implementation should introduce a **minimal real affordance environment** rather than only extending status text. The environment should define durable objects with affordances, allow actions that consume/break affordances, write endocrine-tagged episodic memory to SQLite, and expose associative recall when a new action resembles prior self-caused loss.

The first object should be deliberately small but meaningful: for example, a `glass_bridge` or `signal_lamp` that offers affordances such as `cross`, `illuminate`, `coordinate`, or `inspect`. A destructive action such as `strike`, `overdrive`, or `break` must remove those affordances and create a durable episode tagged with loss, guilt, arousal, caution, agency, and irreversible affordance loss.

## Edge-Model Finding

The repository has local-model shapes but not an active real 0.5B-1B sentence-completion implementation. The first safe integration step is therefore to add a clean adapter contract and endpoint that makes the missing edge model explicit while keeping a deterministic fallback. A later iteration can mount a GGUF model through llama.cpp/KoboldCpp and satisfy the same interface without changing the cognitive environment API.
