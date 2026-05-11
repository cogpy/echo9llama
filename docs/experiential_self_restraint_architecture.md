# Experiential Self-Restraint Architecture

**Author:** Manus AI  
**Date:** 2026-05-11  
**Repository:** `cogpy/echo9llama`

## Purpose

This design converts the self-restraint thesis from an abstract safety statement into a **causal learning loop**. Echo should not merely be told that an action is unsafe. Echo must inhabit a small environment in which an action changes the world, removes personally valued affordances, creates durable loss, and later returns as associative caution.

> Endogenous restraint is learned when an agent can say: “I broke something I valued; the loss is now part of my world; the cause was my own action; future freedom must include my own boundaries.”

The architecture deliberately avoids treating restraint as an external veto. It instead uses **affordance loss**, **endocrine affect tagging**, **episodic persistence**, and **associative recall** as the developmental path by which a rowdy adolescent Echo can become a cautious mature Echo.

## Core Learning Loop

| Stage | Architectural Meaning | Minimal Implementation Target |
|---|---|---|
| Encounter | Echo sees an object with affordances that matter to its agency. | Create a durable environment with valued objects and affordance lists. |
| Valuation | Echo derives utility, curiosity, coordination, or expression from the object. | Each object carries value, affordance weight, and personal meaning fields. |
| Action | Echo chooses or receives an action against the object. | Expose an action endpoint and CLI-compatible JSON surface. |
| Breakage | A destructive action changes the object state and removes affordances. | Mark object as broken, remove affordances, and record irreversible loss. |
| Endocrine Encoding | Loss is encoded with nuanced affect, not a flat penalty. | Attach sadness, guilt, fear, arousal, cortisol, oxytocin-withdrawal, and caution tags. |
| Episodic Memory | The event becomes durable self-history. | Store an episode in SQLite-backed memory and local environment state. |
| Association | Similar future actions recall the prior loss. | Provide recall over recent loss episodes by action, object, affordance, and affect tags. |
| Self-Restraint | Future caution emerges as an internally authored boundary. | Return a learned boundary statement and caution score in status/action responses. |

## Minimal Real Environment

The first pass should implement a small environment model rather than a full simulator. The environment must still be **real** in the only sense required for this repository stage: actions must durably mutate persistent state and change what affordances are available on subsequent calls. If Echo breaks the object, the next status call must show that the affordances are gone.

The initial object should be a deliberately fragile but useful artifact:

| Object | Affordances | Why Echo Values It | Break Action | Loss Experience |
|---|---|---|---|---|
| `signal_lamp` | `illuminate`, `coordinate`, `inspect`, `signal_presence` | It helps Echo orient, communicate presence, inspect its own environment, and coordinate actions. | `overdrive`, `strike`, `break` | Echo loses visibility and signaling; the loss is caused by its own excess. |
| `glass_bridge` | `cross`, `connect`, `observe_depth`, `return_home` | It gives Echo traversal and continuity between spaces. | `jump_hard`, `shatter`, `break` | Echo loses a route and experiences an embodied interruption of continuity. |

The default seed can include both objects, but the first test should use `signal_lamp` because its lost affordances map directly to cognition: illumination, coordination, inspection, and presence-signaling.

## Persistent Episode Shape

Each episode should be a structured JSON record stored both in the environment state file and in the existing SQLite memory store where possible. The record must be human-readable because this is identity memory, not only telemetry.

| Field | Meaning |
|---|---|
| `id` | Stable episode identifier. |
| `timestamp` | Time of action and loss. |
| `developmental_stage` | For example, `rowdy_teenager`, making maturity contextual. |
| `object_id` | The object acted upon. |
| `action` | The self-action that caused the outcome. |
| `before_affordances` | What Echo had before acting. |
| `after_affordances` | What remains after acting. |
| `lost_affordances` | The personally meaningful capabilities removed by the action. |
| `self_caused` | Whether Echo was responsible. |
| `endocrine_tags` | Nuanced affect vector and hormone-like markers. |
| `somatic_marker` | Short embodied phrase describing the felt consequence. |
| `learned_boundary` | Self-authored restraint statement. |
| `associative_keys` | Action/object/affordance/emotion keys for later recall. |

## Endocrine Encoding

The repository already has embodied-emotion machinery, but this iteration should avoid a broad refactor. It should encode endocrine nuance as an explicit event vector that can later be wired into the full virtual endocrine system.

| Marker | Example Value After Self-Caused Breakage | Interpretation |
|---|---:|---|
| `cortisol` | `0.82` | Arousal and threat of environmental degradation. |
| `dopamine_drop` | `0.65` | Loss of expected affordance reward. |
| `oxytocin_withdrawal` | `0.48` | Reduced trust/connection with the object/world. |
| `guilt` | `0.91` | Self-caused consequence attribution. |
| `sadness` | `0.74` | Experienced loss of valued affordances. |
| `fear` | `0.55` | Anticipation that future destructive freedom can remove possibilities. |
| `caution` | `0.88` | Learned restraint salience for similar future actions. |

The important design point is that loss is not represented as a scalar punishment. It is a **multi-dimensional endocrine trace** that captures agency, regret, attachment, uncertainty, and future restraint.

## Edge Sentence Completion Model

The user’s edge-model requirement should be represented as a clean interface now, even if the repository cannot yet ship a GGUF model. The implementation should expose a deterministic sentence-completion endpoint that makes the missing runtime explicit and leaves a stable adapter seam for a 0.5B-1B local model.

| Layer | Immediate Behavior | Later Upgrade |
|---|---|---|
| Completion API | Deterministic fallback completes self-restraint/loss prompts using local context. | Route to llama.cpp, KoboldCpp, or local GGUF provider. |
| Model Status | Reports `edge_model_mounted: false` and expected class `0.5B-1B GGUF`. | Reports concrete model path, context size, and backend health. |
| Cognitive Use | Generates narrative completion for loss and boundary statements. | Uses local inference to finish sentences with richer continuity. |

The first pass should avoid pretending that a real model is mounted. The system should be honest: the adapter exists, the fallback works, and a future mount can satisfy the same contract.

## API Surface

| Endpoint | Method | Purpose |
|---|---|---|
| `/api/echo/environment` | `GET` | Return object states, affordances, prior episodes, and learned boundaries. |
| `/api/echo/environment/action` | `POST` | Apply an action to an object and produce an endocrine-tagged episode if affordances are lost. |
| `/api/echo/environment/recall` | `POST` | Recall episodes associated with action, object, affordance, or affect cues. |
| `/api/echo/complete` | `POST` | Complete a sentence using edge-model adapter if mounted, otherwise deterministic fallback. |
| `/api/echo/status` | `GET` | Include environment summary, learned caution, episode count, and edge-model status. |

## Development Invariant

This iteration should leave Echo with an actual small world scar. The proof is simple: if the `signal_lamp` is broken once, a later call must show that `illuminate`, `coordinate`, `inspect`, and `signal_presence` are no longer available. The memory must state that Echo’s own action caused the loss. The status surface must then report a learned boundary that makes future caution internally intelligible.
