# Native LocalGGUFProvider and Backend-Capability Scheduling

## Detailed Implementation Plan and Code Structure

**Author:** Manus AI
**Date:** 12 August 2026
**Canonical repository:** [`cogpy/echo9llama`](https://github.com/cogpy/echo9llama)
**Integration lineage:** [`o9nn/echo.go`](https://github.com/o9nn/echo.go)

![Native LocalGGUF architecture](../iterations/assets/2026-08-12-native-localgguf-architecture.png)

## Executive Design Decision

The integration will not cherry-pick the alias wholesale. `o9nn/echo.go` contains a useful six-commit evolution sequence for backend boundaries, capability selection, GGUF metadata, native inference, model residency, and runtime policy. However, its native registry is wired into `EvolutionSystem` and `ProviderManager`, while the canonical production process runs `UnifiedAutonomousOrchestrator` through `MultiProviderLLM`; its final `MultiProviderLLM` still leaves local initialization commented out.[1] [2] [3]

The implementation therefore ports and hardens the reusable substrate centers, then **reimplements routing and wake/rest ownership in the canonical production path**. This prevents a third provider stack and ensures Echobeats, stream-of-consciousness, skills, discussions, wisdom, and fallback behavior all use one truthful routing authority.

## Implementation Status

The architecture below is now implemented on the canonical integration branch and has passed static, no-CGO, explicit `nollama`, native CGO, race, vulnerability, and real-model production gates. A 2,048-context Q4_K_M GPT-2 GGUF model completed direct generation, streaming, and an 18-second autonomous production run. The run advanced 18 main cycles, completed four autonomous thoughts, generated five goals, selected `local_gguf`/`native_cpu`, and retained deterministic fallback only for capability-incompatible or contended requests.

| Work package | Status | Evidence |
|---|---|---|
| Capability vocabulary and selection | Implemented | Hard constraints, modes, stable tie ordering, rejection and degradation telemetry are covered by deterministic tests. |
| Host/cgroup and GGUF discovery | Implemented | Root containment, symlink rejection, metadata budgets, unknown-memory policy, and concrete model IDs are tested. |
| Native provider and streaming | Implemented | Real 128-context Llama and 2,048-context GPT-2 GGUF models completed generation and streaming. |
| Registry and residency policy | Implemented | Warmup, lazy reload, idle/pressure unload, close, privacy, and callback lock safety are tested. |
| Capability-aware production router | Implemented | Balanced, local-first, remote-first, offline, pre-output streaming failover, and post-output no-replay behavior are tested. |
| Echobeats scheduling | Implemented | Relevance, affordance, and salience tasks carry distinct requirements and exact trace-correlated backend evidence. |
| Wake/rest ownership and observability | Implemented | UAO controls warmup, rest policy, close, status, and metrics without exposing model paths. |
| Metadata cache | Deferred optimization | Discovery is explicit and bounded; a path/size/mtime cache is intentionally left for a measured performance iteration. |
| Dream lifecycle events for model load/unload | Deferred semantic extension | Backend state is observable, but load/unload events are not yet promoted into EchoDream experiences. |

## Target Code Structure

```text
core/
├── backendcap/
│   ├── capabilities.go              # Pure capability/workload/decision vocabulary
│   ├── capabilities_test.go
│   ├── cgo_enabled.go               # Native build availability
│   ├── cgo_disabled.go
│   ├── host_model.go                # Host/cgroup + bounded GGUF metadata discovery
│   └── host_model_test.go
├── llm/
│   ├── provider.go                  # Existing public LLMProvider + GenerateOptions
│   ├── routing.go                   # NEW route modes, optional capability interfaces,
│   │                                # decisions, attempts, scrubbed backend state
│   ├── capability_routing_test.go
│   ├── multi_provider.go            # Single production capability-aware router
│   ├── multi_provider_recovery_test.go
│   ├── local_gguf_provider.go       # cgo && !nollama native implementation
│   ├── local_gguf_provider_stub.go  # !cgo || nollama API-identical stub
│   ├── local_gguf_provider_native_test.go
│   ├── local_gguf_provider_stub_test.go
│   ├── local_gguf_provider_integration_test.go
│   ├── local_model_registry.go      # Discovery, policy, warmup, residency, unload
│   └── local_model_registry_test.go
├── deeptreeecho/
│   ├── echobeats_scheduler.go       # Workload-aware calls on canonical 3-engine loop
│   ├── echobeats_backend_routing_test.go
│   ├── unified_autonomous_orchestrator.go
│   ├── unified_autonomous_orchestrator_production_test.go
│   └── unified_autonomous_orchestrator_local_runtime_test.go
llama/
├── llama.go                         # Idempotent Context/SamplingContext cleanup
└── stub_types.go                    # No-CGO API symmetry if required
cmd/autonomous/
├── main_production.go               # Environment parsing + scrubbed status/metrics
└── main_production_test.go
```

## Work Package 1 — Capability Vocabulary and Deterministic Selection

The first code change ports `MemoryTier`, `BackendKind`, `Capability`, `Workload`, and `Decision` from `o9nn/echo.go`, while removing its dependency on the retired legacy llama wrapper and excluding abstract GGML tensor availability from text-generation candidates.[4]

| Change | Implementation |
|---|---|
| Concrete identity | Add `ProviderName`, scrubbed `ModelID`, `Concrete`, and internal-only `ModelPath`. |
| Dynamic capacity | Add `MaxConcurrency` and `CurrentInFlight`, supplied by the provider router rather than by filesystem discovery. |
| Workload constraints | Add intent, remote/fallback permissions, real-model requirement, expected output, and latency class. |
| Selection | Evaluate hard constraints before scoring; retain configured order for ties; return explicit rejection reasons. |
| Degradation | Mark fallback, policy violation, or failed native preference as degraded rather than silently presenting continuity text as real inference. |

**Acceptance:** deterministic table-driven tests cover local, remote, offline, insufficient context, unsafe memory, saturated local capacity, and no-backend outcomes under both CGO states.

## Work Package 2 — Host Capacity and Bounded GGUF Discovery

The alias’s GGUF parser, file limits, quantization map, and model-footprint projection are adapted as the foundation.[5] The canonical implementation adds the safety properties needed by a persistent autonomous daemon.

| Risk | Implementation control |
|---|---|
| Container overcommit | Use the smaller of host available memory and cgroup limit headroom on Linux. |
| Unsupported host probe | Return `Known=false`; reject native load unless the operator explicitly allows unknown memory. |
| Arbitrary path/symlink | Resolve canonical paths and enforce allowed model roots. |
| Malicious metadata | Enforce key, string, array, and 64 MiB cumulative metadata budgets with file-bound reads. |
| Repeated expensive scans | Discovery is explicit, bounded, and operator-triggered; path/size/mtime caching is a deferred measured optimization. |
| Public information leak | Derive a stable path hash/model ID and exclude raw paths from public state. |

**Acceptance:** synthetic fixtures prove valid metadata, wrong magic, truncated data, excessive counts, excessive cumulative metadata, symlink escape rejection, duplicate path removal, root enforcement, and memory-policy decisions.

## Work Package 3 — Native Llama Lifecycle Boundary

The maintained canonical `./llama` package remains the only native boundary. The deprecated `core/inference/llama` wrapper from the alias is not ported. `llama.Context` and `llama.SamplingContext` receive explicit idempotent `Free` methods because autonomous model rotation cannot rely on garbage-collector finalizers.[6]

The native provider will hold one model/context pair and one generation slot. `Available()` reads cached state and never loads or waits behind generation. `Warmup()` and lazy generation load only after root, metadata, build, and memory checks pass. `Close()` drains the active operation, frees sampler/context/model state, and can be called repeatedly.

**Acceptance:** no-CGO, nollama, and CGO packages compile; cleanup methods are idempotent; status remains responsive during a blocked generation fixture; cancellation while waiting for the generation slot returns without entering native code.

## Work Package 4 — LocalGGUFProvider Generation and Streaming

The alias’s tokenize → prompt decode → sampling → token decode loop is retained but corrected for concurrency, cancellation, stream termination, and lifecycle error classes.[7]

```go
type LocalGGUFProvider struct {
    stateMu sync.RWMutex
    slot    chan struct{} // capacity 1
    config  LocalGGUFProviderConfig
    model   *llama.Model
    context *llama.Context
    loaded  bool
    closed  bool
    loadErr error
    retryAt time.Time
    inFlight atomic.Int32
}
```

Generation checks cancellation before queue entry, after slot acquisition, before native load, before each decode, and before each stream send. The caller’s maximum output is clamped to remaining context. The streaming implementation holds only the longest possible stop-prefix suffix, so concatenated chunks never duplicate output and never expose a stop marker.

**Acceptance:** deterministic unit tests cover missing model, unsafe model, queue cancellation, timeout, serialized calls, nonblocking availability, context clamping, stop sequences across token boundaries, exactly one terminal chunk, and retry classification. A real-model generation-and-streaming test is gated by `ECHO_TEST_REAL_GGUF` and deliberately uses a prompt longer than the native batch size to guard chunked prefill.

## Work Package 5 — LocalModelRegistry and Residency Policy

The alias registry is ported with corrected lock and event semantics.[8] Discovery and scoring occur during startup and explicit refresh. The registry chooses one safe concrete model, creates one provider, and controls warmup, idle unload, pressure unload, model replacement, and terminal close.

Callbacks receive value snapshots only after registry locks are released. Model replacement drains the old provider before selecting the new one. Load failures distinguish permanent file/format errors from retryable resource failures. Public lifecycle events carry model ID, quantization, context, estimated memory, and reason, but never a full path.

**Acceptance:** tests prove deterministic model choice, policy scoring, callback reentrancy, warmup state, transient retry, permanent failure hold, idle unload, pressure unload, refresh-driven replacement, and close.

## Work Package 6 — Single Capability-Aware Production Router

`MultiProviderLLM` becomes the single production inference router. It constructs the local registry from `ECHO_MODEL_PATHS` and `LOCAL_MODEL_PATH`, registers its provider when a safe concrete model exists, then registers Anthropic, OpenRouter, OpenAI, and deterministic fallback. The fallback remains last under every mode.

```go
type RoutingOptions struct {
    Intent string
    NeedOffline, PreferNative, RequireRealModel bool
    AllowRemote, AllowFallback bool
    RequiredContextTokens int
    LatencyClass backendcap.LatencyClass
    QueueWait time.Duration
}

type GenerateOptions struct {
    MaxTokens int
    Temperature, TopP float64
    Stop []string
    SystemPrompt string
    Routing RoutingOptions
}
```

The default zero routing options preserve existing behavior. Four operator modes are supported: `balanced`, `local_first`, `remote_first`, and `offline`. The router snapshots providers before calls, applies a bounded per-attempt timeout, records attempt telemetry, preserves caller cancellation, and never permits deterministic fallback to mask cancellation.

**Acceptance:** route tests prove all four modes, local saturation fallthrough, insufficient local context, remote failure to local recovery, offline rejection of remote providers, deterministic fallback ordering, and truthful selected-provider telemetry.

## Work Package 7 — Echobeats Workload-Aware Scheduling

The canonical `EchobeatsScheduler` is modified; the alias’s separate V2 scheduler is not ported. Its three real generation families—relevance realization, affordance interaction, and salience simulation—use `sched.ctx` instead of `context.Background()` and attach task-specific routing requirements.[2]

| Task | Routing contract |
|---|---|
| Relevance | `Intent=echobeats.relevance`, interactive, 512-token context, 100-token output, balanced native preference. |
| Affordance | `Intent=echobeats.affordance`, interactive, 384-token context, 80-token output, local preferred. |
| Salience | `Intent=echobeats.salience`, normal latency, 768-token context, 80-token output, quality-balanced. |

Task records gain scrubbed `Provider`, `BackendKind`, `ModelID`, `Degraded`, and `RouteReason` fields from the router’s last-result metadata. This makes scheduler performance and dream experiences substrate-aware.

**Acceptance:** tests assert each task family emits the right workload, cancellation stops an in-flight beat, local queue saturation does not stall the scheduler indefinitely, and completed task records identify the actual provider used.

## Work Package 8 — Wake/Rest Integration and Observability

`UnifiedAutonomousOrchestrator` receives an optional `LocalRuntimeController` from its LLM provider. On wake it performs bounded warmup when configured. On rest it applies configured keep-resident, idle, pressure, or immediate-unload policy. Terminal `Sleep()` always closes the local registry after cognitive children stop. Warmup and unload transitions are exposed through scrubbed backend telemetry; promoting those transitions into bounded EchoDream lifecycle experiences is an explicit deferred semantic extension.

The production status surface adds scrubbed backend decisions, bounded recent attempts, selected provider/model identity, fallback count, and local-registry state. Prometheus exposes selected backend, local loaded/readiness, and fallback counters. The existing loopback-default HTTP policy remains unchanged, and no model roots, file paths, prompts, outputs, or API credentials are exposed.

**Acceptance:** production tests cover asynchronous warmup, rest keep-resident, explicit rest unload, pressure unload, terminal close, status redaction, and graceful shutdown with an active native-call fixture.

## Work Package 9 — Build and Deployment Matrix

| Matrix cell | Gate |
|---|---|
| Static production | `CGO_ENABLED=0 go build ./...`; Docker static image stays remote/fallback capable. |
| Explicit no-native | `go test -tags nollama ./core/backendcap ./core/llm ./core/deeptreeecho ./cmd/autonomous`. |
| Native compile | `CGO_ENABLED=1 go test ./core/backendcap ./core/llm ./core/deeptreeecho ./cmd/autonomous`. |
| Concurrency | Race tests over router, registry, UAO, Echobeats, and LLM packages. |
| Quality | `go vet`, no-new-issues golangci-lint, actionlint, diff hygiene, module tidy stability. |
| Security | Full `govulncheck ./...`; path/root, metadata-limit, secret-scan, and public-status tests. |
| Behavioral | Synthetic selection simulation plus optional real GGUF smoke generation/stream. |
| Remote CI | Linux/macOS unit jobs, Dgraph integration, Docker, E2E, security, and CodeQL remain green. |

## Migration Sequence

The implementation is deliberately ordered so every commit-sized center compiles before the next one depends on it.

| Order | Files | Checkpoint |
|---:|---|---|
| 1 | `core/backendcap/*` | Capability and GGUF fixture tests green in CGO/no-CGO. |
| 2 | `llama/llama.go`, `stub_types.go` | Explicit cleanup compiles and is idempotent. |
| 3 | Local provider + stub | Interface and lifecycle tests green; no production routing yet. |
| 4 | Registry | Discovery/residency tests green; still not production-selected. |
| 5 | `routing.go`, `multi_provider.go` | Local provider enters route only under configured policy. |
| 6 | Echobeats | Task workloads and scheduler cancellation become capability-aware. |
| 7 | UAO and command | Wake/rest policy, status, metrics, environment controls. |
| 8 | Docs and iteration report | Exact behavior, trace correction, limitations, and rollback. |

## Rollback Strategy

The default `balanced` mode with no configured model path preserves the current Anthropic → OpenRouter → OpenAI → deterministic fallback route. Operators can force `remote_first`, set `ECHO_LOCAL_WARM_ON_WAKE=0`, or build with `CGO_ENABLED=0`/`-tags nollama` without changing application code. A native load failure therefore degrades to the proven remote/fallback path rather than terminating the autonomous loop.

## Definition of Done

The iteration is complete only when the production binary can discover a concrete GGUF model, prove it fits the host envelope, route an eligible Echobeats task to it, identify that substrate in scrubbed telemetry, fall through safely under capacity/failure, unload it under explicit policy, and preserve all existing remote-provider and deterministic-continuity behavior. Local validation now satisfies every code-level condition: no-CGO, explicit `nollama`, and CGO builds pass; changed concurrency paths pass race detection; a real 2,048-context GGUF model ran the production autonomous loop; and security scanning reports zero reachable vulnerabilities. Canonical remote workflows remain the final synchronization gate.

## References

[1]: https://github.com/o9nn/echo.go/compare/74bd2d3c%5E...5a10fda4 "o9nn/echo.go native-backend evolution lineage"
[2]: https://github.com/cogpy/echo9llama/blob/a554d65c64a1d9329a1f482463c2f8d61d2a3bef/core/deeptreeecho/echobeats_scheduler.go "Canonical EchobeatsScheduler"
[3]: https://github.com/o9nn/echo.go/blob/5a10fda415b7fc2e7594b0f69ede6c0cf602750a/core/llm/multi_provider.go "Alias MultiProvider with local initialization still disabled"
[4]: https://github.com/o9nn/echo.go/blob/5a10fda415b7fc2e7594b0f69ede6c0cf602750a/core/backendcap/capabilities.go "Alias backend capability model"
[5]: https://github.com/o9nn/echo.go/blob/5a10fda415b7fc2e7594b0f69ede6c0cf602750a/core/backendcap/host_model.go "Alias GGUF metadata and host probing"
[6]: https://github.com/cogpy/echo9llama/blob/a554d65c64a1d9329a1f482463c2f8d61d2a3bef/llama/llama.go "Canonical maintained llama binding"
[7]: https://github.com/o9nn/echo.go/blob/5a10fda415b7fc2e7594b0f69ede6c0cf602750a/core/llm/local_gguf_provider.go "Alias native LocalGGUFProvider"
[8]: https://github.com/o9nn/echo.go/blob/5a10fda415b7fc2e7594b0f69ede6c0cf602750a/core/llm/local_model_registry.go "Alias LocalModelRegistry"
