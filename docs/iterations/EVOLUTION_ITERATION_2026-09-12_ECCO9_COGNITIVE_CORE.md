# Deep Tree Echo Evolution Iteration: ecco9 Cognitive Core

**Date:** 2026-09-12
**Target repository:** `cogpy/echo9llama`
**Integrated lineage:** `o9nn/ecco9@1b22401ee8842fd1aa2769b688d588cd9f1f9ce9`
**Iteration:** E2 — provenance-preserving cognitive-core integration

## Executive summary

This iteration integrates the complete tracked Go lineage of `o9nn/ecco9` into `cogpy/echo9llama` without replacing the canonical runtime or allowing stale and simulated source subsystems to compete for authority. All **553 tracked Go files**, **45 Go embed assets**, and **3 applicable license files** are preserved byte-for-byte in a nested module at `cognitive-core/ecco9`. Every preserved file has a SHA-256 record and one explicit disposition in an integration catalogue. The lineage manifest itself is pinned by SHA-256 in the active adapter.

A six-file reviewed source subset now operates as a typed cognitive-core adapter: the ecco9 platform contracts plus reservoir, hypergraph-memory, affect, and layered-consciousness drivers. At a one-minute default awake cadence, bounded public cognitive state passes through those drivers. The resulting observation is marked `adapted_observed`, carries pinned source and activation provenance, enters the append-only SQLite cognitive event spine, and becomes source-linked EchoDream material. On restart, up to the latest 4,096 durable inputs rebuild the source devices' temporal state before waking cognition resumes. Canonical Echobeats, EchoDream, provider routing, identity, persistence, policy, and tool authority remain authoritative.

The iteration also repairs a canonical consciousness-hub panic affecting histories of 10–19 messages, a read-lock mutation race, nil-message handling, misleading failed-delivery history, and mutable history exposure. The focused production surface, active source packages, race detector, vet, module integrity, both CGO build modes, and deterministic lineage regeneration pass. The repository-wide historical test suite still exposes unrelated pre-existing failures in `sample`, `examples`, `orchestration`, and legacy `server` tests; these are recorded rather than concealed.

## 1. Intent and constraints

The requested direction was to integrate all Go files from `o9nn/ecco9` as a cognitive core while advancing Deep Tree Echo toward persistent, self-orchestrated, wisdom-cultivating autonomy. A direct overlay was rejected because the source and target share an ancestor but have diverged substantially: the source contributes **53 source-only commits**, while the canonical target has **167 target-only commits** after their shared base. Direct copying would have introduced duplicate symbols, stale module imports, competing schedulers, simulated inference paths, and known concurrency failures.

The integration therefore follows three explicit rules:

1. **Preserve everything.** Every source Go file remains available with immutable provenance.
2. **Activate narrowly.** Only reviewed capabilities cross a typed adapter into the canonical runtime.
3. **Keep one authority.** The canonical event ledger, Echobeats scheduler, EchoDream cycle, provider router, policy layer, and wake/rest manager continue to own decisions and lifecycle transitions.

## 2. Source-lineage inventory

The reproducible importer is [`scripts/sync_ecco9_cognitive_core.py`](../../scripts/sync_ecco9_cognitive_core.py). It requires a clean checkout whose origin and `HEAD` exactly match the pinned source repository and commit, copies only tracked Go files and tracked required assets, writes a minimal nested `go.mod`, preserves the upstream module files, and emits two machine-readable records:

| Artifact | Purpose |
|---|---|
| [`LINEAGE_MANIFEST.json`](../../cognitive-core/ecco9/LINEAGE_MANIFEST.json) | Records source repository, exact commit, source commit time, every preserved byte count, and every SHA-256 digest. |
| [`INTEGRATION_CATALOGUE.json`](../../cognitive-core/ecco9/INTEGRATION_CATALOGUE.json) | Assigns every Go file an activation disposition and rationale. |

The deterministic manifest SHA-256 is `b382e99916a2364eb3945770835d3f907d42bbf2f548a7879a2e57c32590a5b7`; the integration-catalogue SHA-256 is `3456bd87a074478b0e60485ec891a1d8d91a443f1aca504840f678dfc8b69b63`. Both are pinned by the active adapter.

| Disposition | Files | Meaning |
|---|---:|---|
| `canonical_same_path_exact` | 149 | Canonical root already contains identical bytes; the snapshot preserves source provenance. |
| `preserved_divergent` | 327 | A canonical same-path file differs; no overwrite was allowed. |
| `preserved_unreviewed` | 71 | The source file has no canonical same-path implementation and remains quarantined. |
| `active_adapter_source` | 6 | The source file is executed only through `core/cognitivecore`. |
| **Total** | **553** | Every tracked Go file has exactly one disposition. |

The six active source files are `core/ecco9/platform.go`, `core/ecco9/types.go`, and the reservoir, memory, emotion, and consciousness drivers under `core/ecco9/drivers`.

## 3. Audit findings that shaped the design

Parallel package audits and direct compile probes identified useful concepts alongside substantial integration hazards.

| Finding | Consequence in this iteration |
|---|---|
| Source and target use different module paths and contain many same-name, semantically divergent types. | The source is a nested module; no bulk overwrite or duplicate root package is permitted. |
| Source provider packages reference missing contracts such as `GenerateOptions`, `ChatMessage`, `ChatOptions`, and `ProviderInfo`. | Source provider and LLM paths are preserved but not activated. Canonical capability-aware routing remains authoritative. |
| Several source cognition paths generate heuristic or simulated outputs, including NPU inference, embeddings, feedback values, and some wisdom/application paths. | Such outputs cannot be labeled observed or used to award wisdom, skill, or goal progress. |
| Source metacognition, interest consolidation, memory reset, unified-agent, stream, phase-manager, and entelechy paths contain deadlocks, lock recursion, lifecycle leaks, or races. | These files remain quarantined in the catalogue. The adapter owns lifecycle and does not call unsafe reset or source orchestration paths. |
| Source persistence includes shallow-copy, collision, migration, and external-service risks. | Canonical SQLite append-only events remain the sole durable authority; no Supabase or source persistence path was activated. |
| The source reservoir, memory, affect, and consciousness devices compile independently and expose small typed device contracts. | These drivers form the first active, constrained cognitive-core subset. |

## 4. Implemented architecture

### 4.1 Nested lineage boundary

The root module imports `github.com/EchoCog/echollama` through a local `replace` directive pointing to `./cognitive-core/ecco9`. The nested module has a minimal dependency surface and therefore does not pull every unreviewed upstream dependency into the canonical build. `go.mod.upstream` and `go.sum.upstream` preserve the original dependency graph for future staged work.

Root `go test ./...` and `go build ./...` do not recursively adopt the nested module as another canonical runtime. Only explicit imports from [`core/cognitivecore`](../../core/cognitivecore) activate source code.

### 4.2 Typed active adapter

[`core/cognitivecore/bridge.go`](../../core/cognitivecore/bridge.go) owns four source devices:

| Device | Bounded role | Current epistemic status |
|---|---|---|
| Echo-state reservoir | Retains a deterministic, leaky temporal trace of bounded canonical input bytes. | `adapted_observed`; not a prediction or proof. |
| Hypergraph memory device | Retains only bounded input digests and reports memory occupancy. | `adapted_observed`; not canonical semantic memory. |
| Emotion processing unit | Produces decaying affective activation from bounded inputs. | `adapted_observed`; heuristic affect, not ground truth. |
| Consciousness layer processor | Projects input activation across basic, reflective, and metacognitive layers. | `adapted_observed`; activation telemetry, not a consciousness claim. |

The adapter validates identity, session, cycle, time, cognitive-load range, interest range, thought count, and total serialized size. It serializes device execution, rejects stale or conflicting same-session cycles, caches same-cycle replay idempotently, scopes cycle monotonicity to a session, and exposes no LLM, network, filesystem, scheduler, or action authority.

The imported memory driver does not enforce its own configured node capacity. The adapter therefore never gives it full thought or interest envelopes: it writes only an input SHA-256 and ordinal, and stops memory-driver writes after 4,096 nodes per process. Reservoir, affect, and consciousness observations continue after that cap. A regression test lowers the cap to two and proves node growth stops exactly there.

### 4.3 Canonical event and dream integration

At a bounded awake cadence (one minute by default), orchestration now performs this path:

```text
canonical bounded state
    -> reviewed ecco9 devices
    -> typed observation + source provenance
    -> cognitive_core.observed append-only event
    -> source-linked EchoDream experience
    -> later dream consolidation and waking integration
```

The durable payload contains both the bounded input and the resulting observation. Replay verifies input schema, source repository, source commit, manifest hash, integration-catalogue hash, exact active source paths, adapter version, trust grade, preserved-file count, cycle agreement, and input SHA-256 before any projection. The event is classified `sensitive` because it may contain bounded public stream-thought excerpts and interests. It is not eligible to self-award skill or goal progress.

On process restart, the adapter scans canonical events in ledger order while retaining a bounded window of the latest 4,096 valid inputs. It replays these inputs through the source devices before the orchestrator publishes the core as ready. This reconstructs reservoir, memory, affect, and layer state without introducing a second persistence authority. Canonical event projection is a serialized one-time reconstruction per process, so bounded source-ID cache eviction cannot cause a second full replay to duplicate EchoDream inputs.

### 4.4 Wake/rest ownership

The canonical wake/rest manager remains authoritative. Rest closes enaction, scheduler, stream, and cognitive-core admission before `isAwake=false` is published. Wake reopens cognitive-core admission before `isAwake=true` is published. Final `Sleep` cancels source-driver background loops and shuts devices down before the event store closes; it is explicitly terminal, and unsafe same-instance `Awaken` calls fail closed. Process restart constructs a fresh orchestrator and rehydrates from the ledger.

### 4.5 Production observability

`/health`, `/status`, and `/metrics` now report cognitive-core readiness, active status, cycle counts, replay counts, active device count, source commit, manifest digest, adapter version, and trust grade. They do not expose raw event payloads or configured absolute state paths.

The feature is enabled by default and can be explicitly disabled with:

```shell
export ECHO_ENABLE_COGNITIVE_CORE=false
```

Set `ECHO_COGNITIVE_CORE_INTERVAL` to change the one-minute default. Both supplied Compose definitions explicitly default the feature to enabled and the cadence to one minute. The Docker build copies the nested module's `go.mod` before dependency download so the local replacement resolves in cached build stages.

## 5. Canonical defect repairs

The source audit exposed a separate production hazard in [`core/consciousness/layer_communication.go`](../../core/consciousness/layer_communication.go). The emergence detector sliced from `len(history)-20` whenever history contained at least ten entries, so histories of 10–19 messages panicked with a negative lower bound. It also incremented emergence metrics while holding only a read lock.

This iteration:

- clamps the emergence window to a valid lower bound;
- snapshots history under a read lock and mutates metrics under a write lock;
- rejects nil messages;
- records history only after successful channel admission;
- deep-clones JSON-compatible context for accepted messages and returned history to prevent nested caller mutation;
- handles nil handler responses safely; and
- returns an empty result for non-positive recent-message requests.

Regression tests exercise the former 10-message panic boundary, nil admission, and immutable history behavior.

The production Compose file also used legacy `ECHO_CYCLE_INTERVAL`, `ECHO_DREAM_INTERVAL`, and `ECHO_DREAM_DURATION` names that the production parser ignored. They now use the supported main-loop, thought, goal, wisdom, and three-phase EchoDream duration variables.

## 6. Verification evidence

The following gates pass on Go 1.25.13:

| Gate | Result |
|---|---|
| Active nested source packages: `go test ./core/ecco9/...` | Pass |
| Focused canonical packages: cognitive core, consciousness, persistence, unified orchestrator, production command | Pass |
| Focused race detector | Pass |
| `go vet ./core/... ./cmd/...` | Pass |
| New-diff `golangci-lint` | Pass, 0 issues |
| Root build with `CGO_ENABLED=1` | Pass |
| Root build with `CGO_ENABLED=0` | Pass |
| `go mod tidy -diff` and `go mod verify` | Pass |
| Importer Python bytecode compilation | Pass |
| Deterministic second import compared byte-for-byte | Pass |
| Manifest verification | Pass: all 553 Go files, 45 embed assets, and 3 licenses |
| Pinned importer contract and negative tests | Pass: repository/commit overrides are unavailable; non-pinned inputs fail hard-coded guards before destination mutation |
| Default and autonomous Docker module-cache stages | Pass; nested replacement exists before module download |
| Redacted gitleaks scan of imported lineage | Pass: 0 findings after two verified public filename/tensor-name false positives were narrowly allowlisted |
| Redacted gitleaks scan of all other iteration files | Pass: 0 findings |
| Two-process production smoke | Pass: first process emitted 5 cadence-bounded observations; second process rehydrated all 5 before a new live cycle |

The complete root test suite still fails only on inherited, unrelated historical surfaces:

- `sample/samplers_test.go` contains a syntax error;
- `examples/api_test.go` has stale imports and a duplicate `main`;
- `orchestration.TestEvolutionTimelineProgression` expects `Integration` but observes `Transcendence`;
- legacy `server` tests reference removed symbols including `LlmRequest`, `errFilePath`, `convertFromSafetensors`, and `quantizeState`.

These failures existed before this integration and do not affect the focused production runtime. They remain explicit backlog rather than being hidden with build tags or deleted tests.

## 7. Security and epistemic boundaries

The imported source is data until an adapter explicitly activates it. The nested module cannot silently register loops, providers, tools, or persistence paths in the canonical runtime. Active observations carry the exact source repository, commit, manifest digest, activation-catalogue digest, source file list, adapter version, and trust grade. Device health demonstrates execution health only; it does not establish factual truth, moral correctness, wisdom, or consciousness.

The source LLM, NPU, discussion, autonomous-agent, self-update, Supabase, and organization paths remain inactive. The adapter accepts no credentials and performs no external effects. Cognitive-core event payloads are stored in the existing owner-only, append-only SQLite ledger and must remain excluded from training exports unless a future explicit redaction and consent policy is added.

## 8. Remaining limitations and next bounded frontier

This iteration provides **continuity and provenance**, not complete source promotion. The source driver algorithms are deliberately simple: reservoir dynamics are byte-oriented, memory edges are not yet populated, affect is length-biased, and consciousness activation is heuristic. Their outputs are therefore low-weight dream material, never direct wisdom or competence awards.

The next bounded iteration should create an evidence-grounded metacognitive adapter rather than activating the source monitor directly. It should first repair the source lock-upgrade deadlock, replace synthetic performance feedback with canonical evaluation events, derive deterministic uncertainty and contradiction signals from the event ledger, and let Echobeats schedule explicit reflection experiments. Promotion should require replay determinism, race-free lifecycle tests, calibrated metrics, and no expansion of action authority.

## 9. Reproduction

```shell
git clone https://github.com/o9nn/ecco9.git /tmp/ecco9
git -C /tmp/ecco9 checkout --detach 1b22401ee8842fd1aa2769b688d588cd9f1f9ce9
cd /path/to/echo9llama
python3 scripts/sync_ecco9_cognitive_core.py /tmp/ecco9 \
  --destination cognitive-core/ecco9 \
  --target-root .

GOTOOLCHAIN=go1.25.13 go test ./core/cognitivecore ./core/consciousness ./core/persistence ./core/deeptreeecho ./cmd/autonomous
GOTOOLCHAIN=go1.25.13 CGO_ENABLED=1 go test -race ./core/cognitivecore ./core/consciousness ./core/persistence ./core/deeptreeecho ./cmd/autonomous
```
