# Echo Runtime Recovery Mapping

This note records the recovery target for the maintained Echo console/server path.

The root Cobra command currently disables the Deep Tree Echo command group: `cmd/cmd.go` comments out `AddEchoCommands(rootCmd)` because the legacy implementation in `cmd/echo.go` is excluded with the `ignore` build tag. The active `serve` command still calls `server.Serve(ln)`, but `server/stub.go` returns immediately, so the maintained runtime does not host either Ollama-compatible HTTP routes or `/api/echo/*` routes.

The disabled legacy CLI surface is still useful. It provides `echo assess`, `echo status`, and `echo think PROMPT`. The status and think commands target `/api/echo/status` and `/api/echo/think`, which should remain the stable boundary for operators.

The maintained DTE backend should not be rebuilt from ignored standalone servers. The strongest active foundation is `core/integration.IntegratedDeepTreeEcho`, which starts the integration hub, exposes `GetStatus`, `GetGestalt`, and supports `InjectThought`. Beneath it, `DeepTreeEchoHub` starts the autonomous agent, unified orchestrator, telemetry shell, and state manager. The active local provider `llm.SimpleFallbackProvider` gives deterministic bootstrapping without API keys.

The recovery should therefore introduce a small active adapter, not resurrect the ignored files. The adapter should wire `server.Serve` to a maintained `net/http` server that starts `IntegratedDeepTreeEcho` with `SimpleFallbackProvider`, exposes `/`, `/api/version`, `/api/tags`, `/api/generate`, `/api/chat`, `/api/echo/status`, `/api/echo/think`, `/api/echo/gestalt`, `/api/echo/remember`, and `/api/echo/recall`, and shuts DTE down on HTTP server return. The root CLI should regain an active `echo` command group whose status and think subcommands call those endpoints, while `assess` can provide a local architecture/self-restraint summary that does not depend on legacy excluded types.

The endogenous self-restraint invariant should be expressed in the restored API response shape. `think` should return not just a text response, but also fields such as `process`, `self_restraint`, `somatic_markers`, and `boundary_request`, making the console/server boundary a developmental interface rather than a command-only safety gate.
