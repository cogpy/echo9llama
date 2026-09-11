# Security Policy and E1 Autonomy Threat Model

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

## Secrets and privacy

API keys must enter through environment variables or an external secret manager. They must not be written to the event ledger, workspace notes, logs, status endpoints, or training artifacts. Public status reports only provider identifiers, model identifiers, bounded route evidence, counters, and readiness. It does not return raw prompts, note bodies, credentials, or configured absolute workspace paths.[1] [3]

The default HTTP listener is loopback-only. Containers bind internally to `0.0.0.0`, while both supplied Compose definitions publish HTTP, model, Prometheus, and Grafana ports on host loopback only. Any broader exposure requires an explicit deployment override plus authentication and network controls. The production HTTP surface remains read-only and exposes no actuation endpoint.[4]

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

## Vulnerability reporting

Report security issues through a private GitHub security advisory for the [`cogpy/echo9llama`](https://github.com/cogpy/echo9llama) repository. Include the affected version or commit, reproduction steps, expected and observed behavior, impact, and any suggested mitigation. Do not include live secrets or private workspace content in a public issue.

## References

[1]: ./core/deeptreeecho/action_policy.go "Deterministic E1 action policy"
[2]: ./core/tools/workspace_note.go "Confined create-only workspace note tool"
[3]: ./core/persistence/cognitive_event_store.go "Append-only cognitive event store"
[4]: ./cmd/autonomous/main_production.go "Production autonomous runtime entry point"
[5]: ./Dockerfile.autonomous "CGO-enabled autonomous production image"
