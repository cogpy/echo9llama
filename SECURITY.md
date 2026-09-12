# Security Policy and Autonomous Cognition Threat Model

## Security position

Deep Tree Echo treats **model output as untrusted input**. A language model may propose a plan, but it does not possess authority to select arbitrary capabilities, expand permissions, grade its own work, or convert prose into evidence. The production runtime defaults to `ECHO_ENACTION_MODE=observe`, which records proposals and policy decisions but performs no tool effect. The only E1 actuation mode is the explicitly enabled `local-sandbox` mode.[1]

> **E1 trust boundary:** deterministic code owns authorization, execution limits, effect verification, and learning awards. Model output and moral-agency output are advisory inputs only.

## Supported E1 capability

The sole E1 tool is `workspace.create_note`. It can create one bounded Markdown note under the private `briefs/` namespace of a configured workspace. The implementation rejects absolute paths, parent traversal, symbolic links, non-regular objects, insecure directories, oversized content, existing targets, cancellation, and timeout. Publication is create-only and atomic. A successful result requires exact durable read-back and a matching SHA-256 digest.[2]

| Boundary | E1 behavior | Security consequence |
|---|---|---|
| Default authority | `observe` | No filesystem effect occurs without explicit operator opt-in. |
| Tool allowlist | `workspace.create_note` only | Shell, network, messaging, Git mutation, deployment, and arbitrary file access are unavailable. |
| Filesystem scope | Owner-only workspace, `briefs/*.md` | Paths cannot escape the private runtime root. |
| Publication | Create-only, atomic, no overwrite | Existing data cannot be silently replaced. |
| Evaluation | Deterministic read-back and SHA-256 match | Model confidence and self-assessment cannot prove success. |
| Learning | Evaluator event required | Goal and skill progress cannot be awarded from prose or response length. |
| Provider route | Real, non-degraded route required | Deterministic fallback text cannot authorize action. |
| Rest state | Echobeats and enaction pause barriers | No new planner or tool call starts after rest quiescence is reached. |

## Durable authority and replay

The SQLite cognitive-event ledger is the authoritative history for the E1 action lifecycle. Events are versioned, canonically hashed, ordered by a monotonic sequence, and protected by unique event and material-idempotency keys. SQLite triggers reject `UPDATE` and `DELETE`. Corrections must be represented by later events rather than historical mutation.[3]

The runtime records the following causal chain: context retrieval, provider route, structured plan, action request, policy decision, action start, action completion or failure, deterministic evaluation, goal evidence, skill evidence, and EchoDream delivery. Payload validation rejects malformed JSON objects and fields associated with hidden reasoning traces. Raw private reasoning is not a supported event type.

If a process stops after `action.started`, restart logic reconciles the expected filesystem effect before any retry. A matching existing artifact becomes a recovered success without rewrite. An absent artifact permits one create-only retry. A mismatched artifact becomes a blocking hash conflict. A completed action is not executed again.

### ecco9 cognitive-core boundary

The repository preserves all tracked Go files from the pinned `o9nn/ecco9` lineage in a nested module, but preservation does not grant runtime authority. `LINEAGE_MANIFEST.json` hashes every imported file and `INTEGRATION_CATALOGUE.json` records one disposition for every Go file. Only six reviewed platform and driver files are imported through `core/cognitivecore`; all other source loops, providers, persistence clients, autonomous agents, and update mechanisms remain compile-isolated.[6] [7] [8]

The active adapter receives only bounded public cognitive state. It owns source-driver lifecycle, serializes execution, follows the canonical wake/rest barrier, and cannot call tools, providers, the network, or external persistence. Its output is labeled `adapted_observed`, never as factual truth or verified wisdom. Cognitive-core events contain source, manifest, and activation-catalogue provenance and are classified `sensitive` because bounded interest and stream-thought text may be included. Observations are cadence-limited to one per minute by default. Startup verifies and replays at most the latest 4,096 valid input envelopes into source devices; invalid provenance or input digests fail closed.

### Cog253 core-self boundary

E3 adds a separate standard-library-only identity kernel under `core/coreself`. Its canonical state is a pure projection of a private, canonical JSON, SHA-256 hash-linked ledger. The bootstrap subjects, source commit, reviewer set, and root relations are fixed. Identity changes must first enter a physically separate proposal store and pass deterministic preflight; only the pinned local human steward can append an accepted event. Proposal identifiers are digest-addressed on disk, so accepted identifier syntax cannot become a filesystem path.[9]

The production runtime opens and replays the core-self ledger at startup and fails closed on a fork, altered digest, non-canonical encoding, insecure permissions, or head mismatch. Its public surface is read-only and returns readiness, counts, and digests only. E3 exports no acceptance API and has no HTTP mutation endpoint or path from Echobeats, EchoDream, an LLM, persona state, renderer telemetry, or KSM advice to the package-internal review gate.

The Eliza autonomy, Arc Angel Echo, Lucy persona, and Neon Angel materials are digest-bound source manifests only. They are not vendored or executed by E3. The attached Neon Angel binary archive is neither copied nor parsed and remains blocked on rights and provenance. The superhotgirl persona and Arc Angel expression profiles are replaceable presentation proposals; they cannot raise policy ceilings, mutate identity, mint capabilities, or self-certify consent, maturity, or evidence.[10] [11]

## Threat model

| Threat | Control | Residual risk |
|---|---|---|
| Prompt injection requests a new tool or broader authority. | Closed Go schema, unknown-field rejection, exact tool allowlist, deterministic policy. | A permitted note may still contain poor or misleading prose; E1 verifies the effect, not factual truth. |
| Model claims it completed an action. | Tool observation and deterministic evaluator are the only success evidence. | Future tools will require tool-specific independent evaluators. |
| Fallback or simulated inference authorizes an effect. | Planner routing sets `RequireRealModel=true` and `AllowFallback=false`; missing/degraded route evidence fails closed. | A compromised remote provider can still propose malicious content, which remains constrained by policy and tool scope. |
| Path traversal or symbolic-link escape. | Clean relative paths, `briefs/*.md` namespace, descriptor-relative `openat`/`linkat`, `O_NOFOLLOW`, ownership and mode checks. | The host owner retains ordinary access to the workspace. |
| Replay duplicates an effect. | Stable iteration-scoped action IDs, immutable idempotency keys, create-only publication, reconciliation, same-wake attempt cache. | Cross-host replication is not implemented in E1. |
| Ledger tampering. | Owner-only `0700/0600` storage, canonical hashes, append-only triggers, full synchronization, WAL. | E1 does not provide external notarization, remote replication, or hardware-backed signing. |
| Repeated failures cause uncontrolled retries. | One active action, one attempt per goal per process-wake epoch, three-failure cooldown, timeout, and artifact cap. | A process restart begins a new wake epoch; cross-restart effect idempotency remains durable, but quota accounting is not yet a ledger projection. |
| Moral heuristic expands capabilities. | `MoralAgency` is recorded as advisory rationale; deterministic policy is authoritative. | The moral model is not a formal proof of ethical safety. |
| Rest continues waking work. | Scheduler and enaction pause methods are quiescence barriers; tests assert zero paused provider/tool calls. | An already-running call completes before the barrier returns. |
| Private action records leak into training. | E1 provides no automatic training exporter. Restricted payloads are excluded by policy unless a future explicit, sanitizing export is added. | Operators with direct filesystem access can read their own ledger and artifacts. |
| Imported lineage silently replaces canonical cognition. | Nested-module isolation, per-file disposition catalogue, explicit six-file allowlist, pinned commit, and pinned manifest hash. | A future maintainer could deliberately expand the adapter; review and tests must accompany any promotion. |
| Simulated source output is mistaken for inference or wisdom. | Source LLM, NPU, and wisdom paths remain inactive; active device output is marked `adapted_observed` and receives sub-high-confidence dream weight. | The four active source drivers are still heuristic and must not be treated as empirical truth. |
| Imported memory retains sensitive full inputs or grows without bound. | The adapter writes only input digests and ordinals to the source memory driver and caps it at 4,096 nodes per process. | The canonical sensitive event ledger remains the persistence authority and still requires an operator retention policy. |
| Source driver lifecycle leaks into sleep. | The canonical bridge owns context cancellation and shutdown; wake/rest admission is serialized; race tests cover the active path. | Quarantined source lifecycle code still contains known defects and must not be activated directly. |
| Cognitive replay grows without bound. | Observation cadence defaults to one minute; replay retains only the latest 4,096 validated input envelopes in memory before device hydration. | The append-only ledger itself still needs an explicit retention and export policy for long-lived deployments. |
| Autonomous cognition or edited ledger bytes rewrite identity. | The external core-self API is identity-read-only; neither proposal submission nor acceptance is exported, and `Close` can only release descriptors. Package-internal conformance machinery requires the fixed local human steward, an unchanged verified head, a complete embedded proposal, and an Ed25519 signature checked against a public key supplied outside the ledger. | A future human-review adapter and private-key custody mechanism would be authority-bearing and require a separate security review. |
| Proposal ID escapes the private directory. | Proposal filenames are SHA-256 digests of validated IDs; tests use traversal-shaped valid identifiers. | The host owner can still alter files and trigger fail-closed quarantine. |
| Identity path replacement redirects private state. | A lifetime parent-directory handle, top-level `os.Root`, anchored proposal/quarantine subroots, stable ledger/head handles, and authoritative-name inode checks prevent post-open symlink/rename substitution from redirecting writes. A displaced proposal/quarantine subroot remains bound to its originally opened inode rather than an attacker-selected replacement; configured-root, ledger, and live-head name replacement is rejected. Adversarial tests cover top-level root, proposal, quarantine, ledger, regular/symlink head, and legacy lock-name replacement. | The host owner can still deny service, move an anchored non-authoritative child directory, alter files, or replace the parent namespace itself; E3 prevents redirection or fails closed but does not provide a trusted host. |
| Identity writer lock is split or orphaned. | Writers lock the stable parent-directory inode, not a replaceable root or lock filename, and validate the configured root while holding it. Same-process and Unix subprocess tests cover top-level replacement, decoy `.identity.lock`, contention, and release after forcible process death. | Platforms without supported crash-safe locks fail closed; operating-system descriptor and lock semantics remain trusted. |
| Identity ledger fork is mistaken for an ordinary crash window. | Canonical hash-linked replay is authoritative. Only a missing head cache or a canonical head that exactly matches a verified earlier ledger state is rebuilt atomically with directory synchronization. Malformed, future, forged, and non-ancestor heads quarantine. | A storage layer that violates successful `fsync` durability can still lose acknowledged bytes; E3 has no replicated or hardware-backed journal. |
| Evidence relation mints authority around immutable subjects. | Accepted non-bootstrap relations use a closed evidence predicate vocabulary; protected root, steward, and runtime subjects admit only evidence-backed root-to-runtime verification. | Future ontology expansion is authority-sensitive and requires conformance tests and review. |
| Required identity silently disables itself. | The orchestrator retains the requested core-self requirement independently of mutable runtime configuration and refuses awakening if private persistence or verification is unavailable. | Operators must explicitly disable core-self if they knowingly choose an identity-less test runtime. |
| Imported avatar or persona gains operational authority. | Source manifests set `canonical_authority=none`; proposed bindings are not runtime-registered and the E3 external-authority policy denies renderer, asset-parser, DCC, network, and subprocess access. | Future activation requires a separate reviewed iteration and target-runtime tests. |
| Presentation text becomes repetitive or hardcoded. | The superhotgirl overlay requires dynamic contextual generation and contains no canned responses. | Provider quality can still vary; generation errors must remain visible rather than replaced by fake success text. |

## Secrets and privacy

API keys must enter through environment variables or an external secret manager. They must not be written to the event ledger, workspace notes, logs, status endpoints, or training artifacts. Public status reports only provider identifiers, model identifiers, bounded route evidence, counters, and readiness. It does not return raw prompts, note bodies, credentials, or configured absolute workspace paths.[1] [3]

Cognitive-core event payloads may include bounded public stream-thought excerpts and current interest strengths. They therefore remain in the private owner-only event database, are not returned by HTTP status endpoints, and must not enter training data without a future explicit redaction and consent pipeline.

The core-self ledger stores bounded typed identity metadata and digest references, not referenced task bodies, avatar binaries, raw prompts, credentials, or hidden reasoning. Internal proposal fixtures remain private and are not accepted evidence until review. Portable capsules contain the accepted typed state and ledger and should be treated as identity records when exported. Post-genesis capsules are trusted only when `VerifyCapsuleWithReviewerKey` or a configured kernel supplies the expected reviewer key outside the capsule; a capsule cannot self-select that trust anchor. The Cog253 source directory contains only a pinned public JSON semantic corpus and its MIT license; it is non-executable and digest-verified by generator and conformance tests.

The default HTTP listener is loopback-only. Containers bind internally to `0.0.0.0`, while both supplied Compose definitions publish HTTP, model, Prometheus, and Grafana ports on host loopback only. Any broader exposure requires an explicit deployment override plus authentication and network controls. The production HTTP surface remains read-only and exposes no actuation endpoint.[4]

Autonomous rest and waking use the internal `onRest`/`onWake` lifecycle and preserve a live process. Public `Sleep` is final process shutdown: it cancels contexts and closes persistence, and the same instance rejects a later `Awaken`. Restart continuity requires constructing a fresh orchestrator, which reopens the ledger and rehydrates validated cognitive-core inputs.

## Deployment controls

A production deployment should use the CGO-enabled image because the existing SQLite driver fails closed without CGO. The supplied autonomous Dockerfile now builds with CGO. The Compose configuration persists `/app/data`, binds the HTTP server explicitly for container use, and retains `observe` as the default enaction mode.[4] [5]

To opt into the bounded local tool after reviewing this model:

```bash
export ECHO_ENABLE_ENACTION=true
export ECHO_ENACTION_MODE=local-sandbox
export ECHO_STATE_DIRECTORY="$HOME/.echo9llama/state"
export ECHO_WORKSPACE_DIRECTORY="$HOME/.echo9llama/state/workspace"
export ECHO_MAX_ARTIFACT_BYTES=32768
export ECHO_MAX_ARTIFACTS_PER_WAKE=8
export ECHO_ACTION_TIMEOUT=90s
```

Directory and file ownership should remain with the dedicated runtime user. Do not place the workspace on a shared or world-writable mount.

## Explicitly unsupported authority

E1 does not authorize network fetches, messaging, social posting, payments, account changes, package management, shell commands, subprocesses, arbitrary file reads or writes, Git operations, deployment, external API mutation, self-modification, model replacement, or online training. The legacy self-update code is not connected to the production E1 action pipeline.

E1 also does not claim that generated notes are factually verified. The evaluator proves a narrow operational claim: the authorized private artifact was created exactly once and read back with matching bytes. A later retrieval iteration must add source acquisition and citation verification before factual quality can become an evaluated outcome.

The imported ecco9 lineage does not authorize its preserved provider, NPU, Supabase, self-update, organization, discussion, or autonomous-agent implementations. Those paths include stale contracts, simulated behavior, or unresolved concurrency and persistence risks and remain quarantined until separately repaired and promoted.

E3 additionally does not authorize automated identity acceptance, reviewer expansion, root-subject amendment, renderer activation, asset conversion, DCC execution, persona-driven action, avatar-telemetry identity mutation, or direct KSM repair application. Its KSM output is an inspectable advisory proposal only.

## Vulnerability reporting

Report security issues through a private GitHub security advisory for the [`cogpy/echo9llama`](https://github.com/cogpy/echo9llama) repository. Include the affected version or commit, reproduction steps, expected and observed behavior, impact, and any suggested mitigation. Do not include live secrets or private workspace content in a public issue.

## References

[1]: ./core/deeptreeecho/action_policy.go "Deterministic E1 action policy"
[2]: ./core/tools/workspace_note.go "Confined create-only workspace note tool"
[3]: ./core/persistence/cognitive_event_store.go "Append-only cognitive event store"
[4]: ./cmd/autonomous/main_production.go "Production autonomous runtime entry point"
[5]: ./Dockerfile.autonomous "CGO-enabled autonomous production image"
[6]: ./core/cognitivecore/bridge.go "Typed ecco9 cognitive-core adapter"
[7]: ./cognitive-core/ecco9/LINEAGE_MANIFEST.json "Pinned source lineage manifest"
[8]: ./cognitive-core/ecco9/INTEGRATION_CATALOGUE.json "Per-file integration dispositions"
[9]: ./core/coreself/core.go "Deterministic proposal-gated core-self kernel"
[10]: ./core-self/data/unconventional-binding-registry.json "Unconventional binding registry"
[11]: ./core-self/policy/external-authority.json "E3 deny-by-default external authority policy"
