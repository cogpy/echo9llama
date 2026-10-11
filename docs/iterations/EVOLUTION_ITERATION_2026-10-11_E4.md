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
go test -race -count=1 ./core/... ./cmd/... ./progress
go vet ./progress/... ./core/... ./cmd/...
CGO_ENABLED=0 go build ./...
CGO_ENABLED=1 go build ./...
go mod tidy -diff
go mod verify
golangci-lint run --new-from-rev=c970f497 --timeout 10m ./progress/... ./core/... ./cmd/...
```

The CI-matched linter reported **0 new issues** after the lifecycle/security follow-up. The final whole-core/command/progress race suite and repository-wide tests both pass with a 120-second per-package timeout. The final compatibility baseline is Go **1.26.0** with patched toolchain **1.26.9**. The affected `golang.org/x/net` dependency was upgraded from `v0.58.0` to `v0.60.0`, with its required Go/x-module version alignment and corresponding `go.mod`/`go.sum` updates. The nested integration module, autonomous Docker builder and three active CI Go pins are aligned; original `go.mod.upstream` and all 553 source Go files remain untouched. The nested upstream `core/ecco9/...` packages compiled (they have no upstream tests).

Deterministic tests cover Brier mathematics, missing/empty evidence, bounded snapshots, invalid confidence, strict citations, duplicate JSON keys, durable replay, no wisdom awards, invalid/degraded model output, source quarantine continuity, deferred admission, restricted proposal goals, one-call step budget, rest/pause, context-aware Stop/restart and concurrency.

### Live inference evidence

The optional live test used the available `gemini-3-flash-preview` endpoint through an explicit test-only OpenAI-compatible HTTP client. The input included an **actually created and read-back-verified local private note**. The output passed schema and source-citation checks and was durably recorded with no competence award. One supplied prediction `0.9` against passed outcome `1` gave artifact Brier **0.010000**; that arithmetic is an operational calibration example, not a wisdom score.

Initial live probes with a 900-token total Gemini budget were truncated and correctly rejected. The optional client now allocates 8192 total tokens because that endpoint includes reasoning in the budget. No schema gate was weakened. Production providers still use their own token semantics and fail closed if their output is truncated; validate each selected production model before deployment.

A final public reflection explicitly distinguished a single success from broader competence. Citation validity proves the cited source exists, **not that every sentence inferred from it is correct**.

### Transient finding retained honestly

One concurrent broad-test run observed a cross-process core-self root-lock assertion. No core-self code differed from baseline. The untouched baseline full suite passed, the targeted regression passed 100 times on baseline and 1000 times on target (plus 50 race runs), and final target broad/race suites passed. The security assertion was not weakened. If this environment-specific failure recurs, capture filesystem/mount and helper-process lock diagnostics rather than accepting a looser result.

Earlier build failures during concurrent file creation are not final validation failures; final checks ran against complete materialized files.

### Repository security backlog

The GitHub push response reported **1547 default-branch dependency alerts**: 456 critical, 445 high, 567 moderate and 79 low. This is GitHub's repository-wide aggregate (including legacy/multi-language dependencies), not a verified count of exploitable vulnerabilities in the canonical E4 runtime. E4 does not remediate that entire backlog; dependency reachability and upgrade triage remain a separate, explicit release concern. The existing hosted security scan should be reviewed before deployment, and no deployment or merge is implied by the passing local checks.

### Post-push lifecycle and security follow-up

A broader `go test -race ./core/... ./cmd/... ./progress` probe identified a real inherited hub send-versus-close race and CLI progress ticker state races. The untouched baseline CLI reproduced the progress races. Repairs keep hub channels open for its terminal lifecycle, gate publishes/commands by running state and cancellation, reject unsupported restarts, serialize progress rendering/shutdown, join the progress worker, eliminate spinner ticker goroutines and synchronize bar rendering/updates. Concurrent lifecycle tests and expanded CI race coverage were added.

The initial hub publication fix exposed a synchronous callback deadlock: integration held `hub.mu` while `AddThought` called back into `publishEvent`. A captured goroutine dump identified this exact cycle. Integration/status methods now snapshot lifecycle state and call subsystems without holding the hub lock; a dedicated callback-reentry test passes. No regression assertion was relaxed.

The expanded race run also exposed Sys6 complementary MP2 goroutines returning outside their cycle's wait barrier and unsynchronized worker counters. All six MP1/MP2 goroutines now remain simultaneous but join the parent cycle; worker counters use their metric-reader mutex, and Stop snapshots/deactivation serialize with cycles. A regression proves exactly 2 dyadic plus 6 triadic worker completions and no late updates after return/Stop; Sys6 race tests pass ten repetitions. The full repaired core/command/progress race suite then passes. These changes repair concurrency, not replace learning or cognitive computation with deterministic output.

Hosted `govulncheck` also found **13 reachable standard-library vulnerability reports** at Go 1.25.13, with fixes listed at Go 1.26.9. The precise toolchain/network-module upgrade above addresses this failure rather than disabling the security gate. The patched local `govulncheck v1.7.0 ./...` passes with **0 reachable vulnerabilities**, while retaining **1 imported-package** and **1 module-only** finding not called by this code. These residuals and the default-branch aggregate do not justify an “all dependencies secure” claim. No unrelated automatic major-version migration was attempted.

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
