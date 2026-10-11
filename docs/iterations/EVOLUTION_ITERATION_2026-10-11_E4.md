# E4 — evidence-grounded metacognition and honest wisdom cultivation

**Date:** 2026-10-11

**Target baseline:** `cogpy/echo9llama@c970f497c2e39ceea30405206ec3abe6284ceba3`

**Branch:** `feat/evolution-e4-evidence-metacognition`

**Source core:** `o9nn/ecco9@1b22401ee8842fd1aa2769b688d588cd9f1f9ce9`

## Result

Deep Tree Echo now has an evidence-cited reflective pivot inside its existing persistent Echobeats loop. It derives measured artifact-outcome/calibration statistics from validated E1 event chains, asks a routed real model for a public provisional lesson, uncertainty and bounded practice proposal, persists that proposal, and feeds the lesson to EchoDream without awarding wisdom or skill competence for prose.

This is a concrete increment toward autonomous wisdom cultivation, **not proof of AGI, subjective awareness, semantic truth, moral wisdom, or general learning competence**. No persistent service was deployed during this iteration; the repository contains the implementation, tests and operating contract.

## Upstream integration and repository discovery

All **553 upstream Go files** are already integrated byte-for-byte beneath `cognitive-core/ecco9`. This iteration verified the complete source tree with SHA-256 comparisons: **zero missing or altered Go files**. The root module uses the existing local replacement and `core/cognitivecore.Bridge` as its canonical runtime boundary. There is no new submodule or duplicate cognitive scheduler.

The production bridge retains its four reviewed devices and `adapted_observed` provenance. The upstream catalogue is a source lineage, not a promise that all historical experimental components are executable or safe to run together. Upstream metacognitive methods contain lock upgrades/recursive locking and invented effectiveness/awareness values; bulk activation would regress the working runtime. E4 adapts the useful process/decision/outcome monitoring idea into a bounded tested projection instead.

Authenticated user-owned organization discovery included `cogpy`, `o9nn` and related Echo/persona implementations, including `o9nn/neuro-sama`, `o9nn/neuro-sdk`, `o9nn/echo.kern` and `cogpy/bolt.echo`. Existing target Echobeats, EchoDream, E1 enaction, core-self and cognitive bridge were the best fit; no unrelated framework or heavyweight runtime was imported.

## Changes

| Centre | Implemented change | Evidence boundary |
|---|---|---|
| Metacognition | New reusable `core/metacognition` analyzer; maximum 64 records, immutable sorted snapshots, counts/rates, missing predictions, Brier score, deterministic source digest | Measures private artifact verification only; empty evidence produces no invented rate |
| Echobeats | Evidence reflection replaces generic salience at step 8 when installed; at most one reflective inference there | Rest/Stop serialize behind the step barrier; inference never holds the orchestrator status mutex |
| Reflection | Strict real-model route and JSON/citation validation; durable `reflection.proposed` event | Explicit `epistemic_status: proposal`, `experiment_type: private_note_plan`; no fake fallback responses |
| Practice | Trusted code constructs a private-note planning goal; existing policy/tool gate remains mandatory | Does not execute model experiment prose, grant shell/network/messaging authority, or award progress |
| Goal admission | New evidence is retained while a prior reflection goal is open; generation is deferred before producing a new proposal | No later valid proposal is silently discarded; durable successes always project idempotently |
| EchoDream | Source-linked, uncertainty-bearing proposed lessons become hypothesis experiences | Dream recurrence and generated text do not increase canonical wisdom totals |
| Scheduler reliability | 100-task per-engine history, locked performance snapshots, defensive goal snapshots, callbacks outside scheduler state locks, idempotent explicit-ID goals, fresh restart context, Stop quiescence | Provider failures are failed tasks, not success-like cognitive prose |
| Wisdom synthesis | Removed prose-length depth, generated confidence/profundity and fabricated automatic effectiveness/application records; calls outside locks; nil/empty provider errors; defensive snapshots | Explicit `RecordWisdomOutcome` accepts caller-supplied observed effectiveness; this legacy API is not a signed/durable verifier and is not wired to canonical wisdom promotion |
| Observability/CI | Versioned `/status` reflection counts, scoped artifact metrics, legacy wisdom qualification; new packages included in CI race suites | Public status excludes source payloads, private paths, lessons and citation IDs |

## Causal validation, replay and liveness

E4 validates the **actual producer order**, not a hypothetical DAG:

```text
context → route → plan → requested → policy → started → completed → evaluation
                                               └─────────────→ failed
context [→ optional route] → error-planning
```

Checks cover deterministic stage/action IDs, identity/goal/correlation/policy/session bindings, causal references, increasing sequence, strict payloads, plan/request/start/content agreement, route metadata and allowed local-sandbox policy. Restart-recovered terminal sessions are accepted only with explicit recovery evidence.

The ledger reloads/hash-validates records. **Hashes and causal shape do not authenticate a hostile authorized writer.** Signing/attestation is a future boundary, not an implied guarantee.

Hash-valid malformed source records are excluded through a deterministic, sensitive quarantine event while their original append-only material is retained. They cannot become calibration or dream evidence. The source cursor advances without duplicate records; hash-integrity failures and unavailable/cancelled stores remain fail-closed. Replay reconstructs measured state and durable counters with bounded memory; reflection failures do not feed themselves back as new evidence. Successful or failed generation consumes the evidence fingerprint, so unchanged evidence cannot cause an endless reflection loop.

Identifiers use full SHA-256 namespaces. JSON parsing also rejects duplicate keys, trailing values, excessive depth and oversized input. The live path reads at most one 256-event source page per reflective cycle plus bounded causal lookups and projection reads. Startup replay is proportional to ledger history in fixed-size pages; it is not an O(1) startup checkpoint.

## Review findings and disposition

| Finding | Resolution |
|---|---|
| Valid experiment lost behind an existing reflection goal | Gate generation before proposal creation; preserve new outcome evidence; project every persisted success |
| Partial/forged chains counted as measured evidence | Full producer-DAG validator and negative regression fixtures |
| Malformed source poisons cursor and repeats evidence | Idempotent source quarantine and ordered source cursor; valid-invalid-valid regression |
| Two provider calls at reflective pivot | Installed reflection replaces step-8 salience, including no-evidence/error cases |
| Reflection counters reset on restart | Replay-derived counters, digest restoration and independent watermarks |
| Model experiment becomes unconstrained planner instruction | Persist explicit proposal/type, reject prohibited requests, trusted private-note-only goal wrapper; E1 policy remains final authority |
| Truncated IDs / duplicate JSON keys | Full SHA-256 IDs and bounded duplicate-key validator |

## Validation

All final local checks below passed:

```bash
go test ./...
go test -race -count=1 ./core/metacognition ./core/cognitivecore \
  ./core/persistence ./core/deeptreeecho ./core/coreself ./cmd/autonomous
go vet ./core/... ./cmd/...
CGO_ENABLED=0 go build ./...
CGO_ENABLED=1 go build ./...
go mod tidy -diff
go mod verify
golangci-lint run --new-from-rev=c970f497 --timeout 10m ./core/... ./cmd/...
```

The CI-matched linter reported **0 new issues**. Go dependencies were verified with no `go.mod`/`go.sum` changes. The nested upstream `core/ecco9/...` packages compiled (they have no upstream tests); full lineage comparison verified all 553 Go files.

Deterministic tests cover Brier mathematics, missing/empty evidence, bounded snapshots, invalid confidence, strict citations, duplicate JSON keys, durable replay, no wisdom awards, invalid/degraded model output, source quarantine continuity, deferred admission, restricted proposal goals, one-call step budget, rest/pause, context-aware Stop/restart and concurrency.

### Live inference evidence

The optional live test used the available `gemini-3-flash-preview` endpoint through an explicit test-only OpenAI-compatible HTTP client. The input included an **actually created and read-back-verified local private note**. The output passed schema and source-citation checks and was durably recorded with no competence award. One supplied prediction `0.9` against passed outcome `1` gave artifact Brier **0.010000**; that arithmetic is an operational calibration example, not a wisdom score.

Initial live probes with a 900-token total Gemini budget were truncated and correctly rejected. The optional client now allocates 8192 total tokens because that endpoint includes reasoning in the budget. No schema gate was weakened. Production providers still use their own token semantics and fail closed if their output is truncated; validate each selected production model before deployment.

A final public reflection explicitly distinguished a single success from broader competence. Citation validity proves the cited source exists, **not that every sentence inferred from it is correct**.

### Transient finding retained honestly

One concurrent broad-test run observed a cross-process core-self root-lock assertion. No core-self code differed from baseline. The untouched baseline full suite passed, the targeted regression passed 100 times on baseline and 1000 times on target (plus 50 race runs), and final target broad/race suites passed. The security assertion was not weakened. If this environment-specific failure recurs, capture filesystem/mount and helper-process lock diagnostics rather than accepting a looser result.

Earlier build failures during concurrent file creation are not final validation failures; final checks ran against complete materialized files.

## Operation

Build the single canonical entry point and use an appropriate persistent host:

```bash
go build -o echoself ./cmd/autonomous
# Configure a supported real provider via environment; never put secrets in Git.
ECHO_STATE_DIRECTORY=/var/lib/echoself \
ECHO_ENACTION_MODE=observe \
./echoself
```

Default enaction mode remains `observe`; this generates proposals but no artifact effects. The evidence-driven reflective path needs qualifying durable E1 outcomes. For authorized private workspace practice, explicitly select `ECHO_ENACTION_MODE=local-sandbox` and retain existing action timeout/artifact/wake budgets. There is no new public communication connector.

`/status.evidence_reflection` exposes bounded scoped measurements. `wisdom` and `wisdom_depth` remain historical compatibility fields; new generation does not increment them and `/status.wisdom_epistemic_status` marks their unvalidated legacy semantics. Reflection counters reconstruct from the ledger; windowed evaluation statistics are gauges rather than lifelong totals.

Optional local live probe:

```bash
ECHO_REFLECTION_LIVE_MODEL=gemini-3-flash-preview \
  go test -v ./core/deeptreeecho -run '^TestEvidenceReflectionLiveModel$' -count=1
```

This probe requires explicitly configured `OPENAI_API_BASE` and `OPENAI_API_KEY`. The test endpoint/client is not a deployed production inference integration.

## Next bounded frontier

1. Replace artifact-only success with **task-specific, independent knowledge and skill evaluators** (falsification, transfer tests, calibration), retaining signed provenance and failure evidence.
2. Promote or revise hypotheses only after independently evaluated outcomes; add durable hypothesis/version/outcome ledgers rather than a scalar “wisdom” reward.
3. Add a bounded, policy-aware discussion adapter with durable incoming/outgoing acknowledgments, interest selection, consent/permission controls and replay-safe start/respond/end behavior.
4. Unify remaining legacy timer loops and queue/backpressure under Echobeats; bound older wisdom graphs/pattern histories and remove nonessential detached goroutines.
5. Introduce a durable measured projection checkpoint for faster large-ledger startup and explicit operator-controlled retry of consumed failed reflection windows.
6. Validate selected GGUF/remote models, run soak tests across repeated wake/rest/restart cycles, and deploy only on an authorized persistent host.

The charismatic, playful, witty “superhotgirl” persona remains configurable and present in reflective prompting. It is not used as a truth, capability, or safety score. The architectural direction remains persistent internally initiated activity, evidence-grounded learning, restorative integration and expressive identity—without confusing generated narrative with achieved wisdom.
