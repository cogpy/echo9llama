# E4 evolution plan — evidence-grounded metacognition

## Verified starting point

Target: `cogpy/echo9llama@c970f497`. Source: `o9nn/ecco9@1b22401ee8842fd1aa2769b688d588cd9f1f9ce9`.
All 553 upstream Go files already exist byte-for-byte under `cognitive-core/ecco9`; retain the immutable lineage and the single canonical runtime. Do not bulk-overlay divergent schedulers/providers or activate upstream synthetic-feedback/deadlocking paths.

## Intended change

1. Add a reusable metacognition analyzer that consumes canonical, hash-validated action outcomes, confidence predictions, and failures. Expose measured outcome/calibration statistics and explicit missing-evidence counts, never a consciousness/wisdom score.
2. Bind real-model, evidence-cited public reflection to Echobeats step 8, with bounded context, timeout, replay-safe event IDs, and no added external-action authority. EchoDream receives source-linked proposed lessons. A generated experiment becomes a bounded goal, not a completed action or competence award.
3. Repair scheduler races, unbounded history, misleading provider-error text, snapshot aliasing, and lifecycle quiescence. Preserve existing 12-step/provider architecture.
4. Verify complete upstream lineage, focused production regression/race tests, vet, builds and module integrity; inspect broad inherited suite failures without hiding them.
5. Document exact progress, remaining limits, operational reproduction, and a next bounded frontier. Commit and push to a feature branch and open a PR; no force push or implicit merge.

## Acceptance criteria

- Generated reflections must come from a routed real model and cite only supplied durable source IDs; invalid/degraded output fails closed and is logged.
- Empty/repeated evidence cannot award progress, repeat goals, or generate endless self-reflection loops.
- Reflection state is reconstructible from the append-only SQLite ledger across process sessions; input work/memory is bounded.
- Pause/Stop are barriers for the reflective scheduler hook; no provider call under the orchestrator state lock.
- Measured confidence calibration is scoped to the exact artifact-verification outcome, not semantic correctness or moral wisdom.
- Persona remains charismatic, witty, playful and truth/uncertainty-aware.
- Tests distinguish deterministic fixtures from live inference; no claim of full AGI or subjective consciousness.

## Explicit scope

One current evolution iteration and repository synchronization. No training in unrelated repositories, new paid infrastructure, live public discussion connector, or unattended deployment is implied. Production deployment remains a separate action on an appropriate persistent host.
