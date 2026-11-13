# E5 — independently evaluated practice and real recurrent learning

**Date:** 2026-11-11. **Baseline:** `04661d67006103f62d5e960b640d9bd1f40fa363`. **Current requested upstream:** `o9nn/ecco9@7d32d715c633a92c684900fbaf25b0111c708612`. **Release state:** implemented, locally validated and synced in PR38; runtime CI passed, but the CodeQL alert gate blocks merge. No automatic main-branch merge is performed.

## What this iteration strengthens

The goal is a persistent, self-directed Deep Tree Echo capable of learning and developing justified wisdom. This iteration strengthens a narrower, demonstrable part of that vision: a canonical event-driven numerical learner and an opt-in, independently checked skill-practice loop. It does not establish consciousness, AGI, general mastery, or moral wisdom, and it does not fine-tune the language model.

Five independent architectural audits covered the scheduler, wake/dream integration, skill/wisdom evidence, discussions/persona, and source lineage. Their central findings were that the imported reservoir's numerical weights were not actually used in its historical smoothing path, metadata prefixes obscured semantic input, failures did not lower competence, prose length generated artificial wisdom scores, and upstream coverage had drifted. A subsequent implementation review reproduced five admission/fairness/replay defects in the first version of this iteration. Those are release gates, not observations to dismiss because unit tests were green.

The prior October E4 PR35 was still open at inspection. This work starts from current main and does not claim that unmerged E4 features already exist there.

## Complete upstream preservation without blind activation

The new immutable `cognitive-core/ecco9/upstream/` snapshot contains every **553 tracked Go files** at the current upstream commit, as exact Git blob bytes, together with **45 embed assets and 3 licenses**. Exact original module files are preserved separately from the generated nested-module boundary. Path-set and SHA-256 tests independently compare every current Go file; the snapshot cannot be silently overwritten by the importer.

The existing reviewed working integration is retained byte-for-byte. Its manifest has **615 records**, not 615 Go files: 553 imported Go files, 14 local Go additions, 45 assets and 3 licenses. Its working Go count is **567**. These retained compatibility overlays originated at `1b22401ee8842fd1aa2769b688d588cd9f1f9ce9`; they are not falsely relabelled as a wholesale current-upstream runtime. All six active native source files are unchanged upstream and match the new snapshot exactly, so their source commit advances with separate working-manifest/catalogue and raw-snapshot provenance.

- Current raw snapshot manifest: `1ec1b76965056f79b3d21bbd5ad737a36b44add8f6a731c0fad71c570ea1a385`.
- Current raw complete path/hash inventory: `212ca92e96d0bf651597d8788a399cd5ed8fd41b0d0f6649230c6e7f5a3f6539`.
- Retained working lineage manifest: `6781879f5ddf50598694f06730b1028d07c8ee93a1b179d2e69281d469c40660`.
- Retained working integration catalogue: `bd864497ddb035716666fc021c4d575f27d4123f3a59de10ea254d7a60db9f88`.

Three-way inspection found 45 textual conflicts, including five substantive API/policy conflicts in metacognitive monitoring, N5 tests/integration, local GGUF provisioning, and the simple server. The current original versions are preserved in the snapshot; the previously reviewed runtime overlays are retained rather than replacing model-only behavior with upstream canned reasoning or enabling competing providers. See `cognitive-core/ecco9/README.md` for the dispositions. Complete source preservation is not a claim that every historical engine builds or is safe to activate. Full dormant nested-module dependency repair remains deferred.

The importer now requires an explicit, guarded source-refresh operation and refuses its former destructive invocation. Eleven Python safety regressions cover exact source identity, mutation refusal, preservation, independent pinning, and byte/mtime idempotence.

Upstream literal strings and source files retain their original whitespace. Git whitespace checks exclude only the immutable `upstream/` snapshot; active code and documentation still pass formatting/whitespace checks, while independent path/blob hashes validate every archived source byte. Normalizing the archive would violate exact-source preservation.

## Canonical numerical learning lane

`core/cognitivecore/outcome_learner.go` is a dependency-free reusable component with no provider, timer, goroutine, tool, or permission authority. It uses complete accepted UTF-8 content through versioned hashed token features, plus explicit pre-outcome task descriptors. Identity/session/cycle prefixes, evaluator scores, labels, held-out splits, and event IDs are not predictor features.

The 32-unit leaky tanh recurrence uses real input and recurrent matrices. Fixed-seed initialization has a computed maximum recurrent absolute row norm of approximately 0.65 and a contraction bound of approximately 0.8775. Reservoir weights are fixed; only a bounded logistic readout is trained by explicit, independently evaluated, non-held-out pass/fail outcomes. Predictions and loss are calculated **before** the current-label update. Unknown outcomes and held-out examples advance temporal state but never alter the readout or optimizer. This chronological held-out interpretation is documented and is not an isolated reset-state benchmark.

The canonical orchestrator applies relevant core observations and practice outcomes in immutable ledger-sequence order. Core observations feed actual thought and interest content, not truncated metadata JSON. Practice examples use the model's candidate artifact content and immutable arithmetic operands; evaluator prose never supplies features. The historical imported smoothing telemetry remains separate and is not relabelled as trained neural computation.

A detached, versioned checkpoint includes numerical matrices, state, readout, optimizer, counts, cursor receipts, and data/checksum hashes. Exact suffix replay is tested. A 256-entry exact receipt window, monotone sequence guard, and permanent bounded Bloom ID guard prevent accepted IDs from retraining after eviction. The Bloom guard can reject new IDs as it saturates; this is disclosed fail-closed availability loss, not probabilistic acceptance.

Runtime startup reconstructs numerical state from the full validated event ledger rather than trusting an editable checkpoint cache. The private, atomic, fsynced `auxiliary_learner_checkpoint.json` export is mode 0600. Checksums detect accidental corruption, not malicious re-signing. Full replay avoids making such a local cache an authority; bounded replay acceleration remains future work.

## Independently evaluated practice

`ECHO_ENABLE_EVALUATED_PRACTICE` defaults **false**. Evaluated trials require explicit opt-in, persistence/enaction, and the existing **local-sandbox** create-only policy. Observe mode cannot publish artifacts. No new shell, network-tool, overwrite, credential, public-message, or external-party action authority is introduced.

The current practice family asks the configured real model to create a private JSON arithmetic artifact containing `task_id`, `sum`, and `product` for immutable bounded operands. Task/goal identity is journaled before scheduling. Exact code checks the actual verified read-back against the task specification, rejecting unknown/duplicate keys, wrong IDs, noninteger/null/missing fields, malformed fences, trailing content and wrong arithmetic. File-publication success and arithmetic correctness are separate skills.

The durable `practice.evaluated` envelope carries a nullable outcome. A known pass or failure requires real route evidence, the complete canonical plan/request/policy/start/completion/evaluation chain, matching identity/session/correlation, verified artifact hashes, the immutable admitted task, and independently recomputed correctness. Provider outage, cancellation, denial or unavailable artifact evidence remains **unknown**, never an invented failure target. Restart recovery preserves the original action session and cannot substitute the task, regenerate an already completed action, or create a second effect.

All five review blockers now have deterministic regressions. Only an independently verified arithmetic pass achieves its goal; failed/unknown trials are abandoned without fabricated progress. When ordinary work is runnable, new practice waits for a newer durable ordinary action context between trials. Permanently malformed admissions are rejected without poisoning later replay, while genuine store/cancellation/integrity errors remain fail-closed and retryable. Admission identity/index/operands, canonical membership and admission-before-context sequence are validated from the ledger, not an in-memory task match. Disabled restored tasks cannot execute. Generic file-publication progress is ignored for arithmetic goals. Active task state is retired promptly and terminal receipts are bounded to 256.

## Honest competence and wisdom

Skill competence now has versioned advisory metadata: explicit evaluation/pass/failure counts, score statistics, a Beta(1,1) pass-frequency posterior, and a documented weight-0.2 EMA. A failure contributes zero and lowers prior positive competence. Empty or invalid/nonfinite evidence is rejected; IDs bind the complete normalized payload. Rehearsal text alone leaves competence unknown. Universal mastery labels are not awarded from practice count.

Wisdom synthesis remains dynamically model-generated, with its existing persona/style options preserved. Provider calls run outside component locks through caller/lifecycle-cancellable, finite boundaries. Generated or revised insights/principles have no prose-length profundity, automatic confidence, fixed synthetic effectiveness, or automatic depth increment. Getters and callback values are detached copies. A separate explicit observed-feedback API can track advisory effectiveness, but no observation is fabricated or automatically inferred from model appraisal. Waking/dream-generated candidates do not increment earned wisdom metrics. State sync no longer overwrites independent persisted curiosity with that unvalidated counter; a private persistence regression proves the trait survives unchanged.

The configurable charismatic, playful, confident “superhotgirl” characteristic remains expressive context beneath truth, privacy, consent, uncertainty and capability policy. This iteration does not invent canned persona replies or broaden social authority. Existing discussion transport/consent and dream-grounding issues remain open.

## Measured validation to date

The numerical learner's signed-scalar **test fixture**, not live AGI behavior, showed first-epoch prequential loss **0.734762**, final-epoch loss **0.284428**, all-training mean **0.428975** across **960 labels**, and chronological held-out mean **0.267360** across **8 labels**. Tests also prove actual readout deltas, recurrent matrix/history effects, different content beyond identical 100-byte prefixes, no unknown/held-out training, duplicate/conflict rejection, detached checkpoints and exact replay.

The optional live smoke test used the catalog-verified **gpt-6.1-sol** through a real OpenAI-compatible endpoint. All **3** arithmetic artifacts passed independent read-back/schema/arithmetic validation; all three fell into the deterministic training split, resulting in **3** readout updates. This is an integration smoke test, not a held-out live benchmark or evidence that task performance improved. Live data hash: `ed4d21580ec3c6533a7faeadd478d4da4dfa5133f86a4c165edab3f00755c3ad`.

Before the final review repairs, full root tests, production-boundary race tests, both autonomous CGO/pure-Go build modes, module verification, and repository-configured new-issue lint passed. Final post-review validation and hosted CI are recorded in the release closeout below.

The production scan discovered newly published [GO-2026-6859](https://pkg.go.dev/vuln/GO-2026-6859), an HTTP/1 header race fixed in Go **1.26.11**. Root toolchain, regular/comprehensive/nightly CI and the autonomous Docker builder are aligned to that patch. The initial fixed production scan found no reachable/imported-package vulnerabilities; three required-module vulnerabilities were reported as not called in that scope. This is a scoped static scan, not a blanket security guarantee.

The broader all-core/command race run then reproduced an inherited integration-hub send-versus-close race and CLI progress renderer races. Reviewed repairs from PR35 were reused, preserving newer main-branch telemetry/provider wiring rather than overwriting it. Hub subsystem callbacks run outside the hub lock; producer shutdown cannot race a closed publication channel. Progress/bar/spinner state is synchronized and renderer/ticker lifetimes end deterministically.

The whole-root vulnerability scan also found eight reachable **Echo HTTP framework v4** findings through the legacy LAMP adapter. The trailing SSH/openpgp entry was only an uncalled-module warning, not that reachable issue. Two official advisories, [GO-2026-6875](https://pkg.go.dev/vuln/GO-2026-6875) and [GO-2026-6872](https://pkg.go.dev/vuln/GO-2026-6872), list no fixed v4 release. The adapter now uses **Echo v5.4.4**, with pointer-context handlers, an owned HTTP server, cooperative request deadlines, preserved routes/payloads, direct-peer IP extraction, and explicit CORS allowlist behavior. No scan suppression or archived-source rewrite is used. Its regressions cover spoofed forwarded headers, all 13 routes, CORS, request binding/deadlines, occupied-address failures and deterministic stop/restart.

Final full-root `govulncheck -show verbose ./...` covers **132 packages and 65 modules** plus the pinned standard library. It reports **zero reachable vulnerabilities and zero imported-package vulnerabilities**. Three uncalled required-module findings remain visible: GO-2026-6866 and GO-2026-6860 in x/net v0.60.0, and GO-2026-5932 in x/crypto v0.57.0. They are a future dependency-maintenance backlog, not silently erased or claimed as evidence of current exploitability.

At the 2026-11-11 feature-branch push, GitHub separately reported **1,693 default-branch dependency alerts**: 465 critical, 497 high, 636 moderate and 95 low. This is the repository's broader [Dependabot surface](https://github.com/cogpy/echo9llama/security/dependabot), not the same scope as the root Go call-graph scan, and it is not proof that those counts represent exploitable E5 runtime paths. It remains an unresolved, separate backlog requiring manifest/runtime mapping. This iteration does not claim the whole repository is vulnerability-free.

Go CodeQL subsequently completed its analysis successfully, but GitHub's separate [CodeQL alert gate](https://github.com/cogpy/echo9llama/runs/130441436490) failed with **48 security annotations**, reported as **5 critical and 43 high**. A complete machine inventory verifies exact source-byte provenance: **27 annotations** target the immutable current upstream snapshot and **21** target unchanged baseline files. No annotation targets a changed E5 active-code file. This is a source-provenance result, **not** a false-positive finding or proof that those flows are safe. The alerts remain unresolved and are not dismissed or excluded from scanning. Archived engines must not be activated without security review; baseline flows also require review before merge.

`EVOLUTION_2026-11-11_E5_CODEQL_INVENTORY.json` records every path, line, message, verified blob hash, scope and unresolved disposition. Annotation categories are 40 path-expression findings, two network-request findings, two integer-conversion findings, one sensitive-logging finding and three unsafe-quoting findings. The 48-annotation gate is distinct from dependency advisories and keeps this PR **not merge-ready**, despite passing runtime CI.

## Remaining autonomy frontier

The next meaningful work is a broader adversarial/held-out practice curriculum and evidence-grounded dream-to-goal reflection, not increasing counters. The current language model does not learn new weights from this iteration; the auxiliary readout estimates outcome likelihood and cannot independently solve arbitrary tasks or decide moral truth. Durable reflection/wisdom promotion needs outcome-qualified causal evidence and repeated transfer tests. A complete fair retry/deadline/backpressure scheduler, bounded dream queue, social consent/transport, trusted persona governance and replacement of private historical driver clocks remain open. No persistent VM service was deployed in this task.

The current numerical component deliberately provides a small reusable foundation rather than importing a second orchestration stack. Source references include the pinned manifests and reviewed owned components in `cogpy/reservoirpy-npu` and `o9nn/opencog-esn`.

## References

[1]: http://www.scholarpedia.org/article/Echo_state_network "Echo State Network equations and readout training"
[2]: https://pkg.go.dev/vuln/GO-2026-6859 "Go HTTP header race fixed in Go 1.26.11"
[3]: https://pkg.go.dev/vuln/GO-2026-6875 "Echo reverse proxy forwarded IP vulnerability"
[4]: https://pkg.go.dev/vuln/GO-2026-6872 "Echo forwarded-header rate-limiter bypass"

The recurrent/readout design follows the standard ESN distinction between reservoir dynamics and output learning [1]. Security remediation uses the official fixed-version records [2] [3] [4].

## Release closeout

Local acceptance includes full root tests; `go test -race -timeout 180s -count=1 ./core/... ./cmd/... ./progress/...`; vet across the same targets; CGO and pure-Go autonomous builds; module tidy/verification; all 11 importer safety regressions; configured new-issue lint; all five review reversals; persistence-curiosity continuity; and the real-model practice smoke test. The final full vulnerability scan is described above, including uncalled-module warnings.

The source snapshot, plan, implementation, tests, CI updates and this report are delivered together on the feature branch. The PR's final validation comment is the durable hosted-check outcome. No main merge or persistent-service deployment is performed. The next iteration must build on the reviewed merged state rather than assuming open PRs are already integrated.

Hosted checks observed for code/documentation commit `4471322b528c9dfe1eeeb17a18746f204fb91b93` passed Linux/macOS unit tests, integration/E2E, builds, Docker, benchmarks, lint, module hygiene, security and CI Summary. All language analysis jobs, including Go, completed successfully. **The aggregate CodeQL alert gate failed with the unresolved 48 findings above.** Devin's apparent pass explicitly skipped full review due to expired trial/no credits and is not counted as substantive review evidence. The final inventory/report follow-up is documentation-only; newly queued reruns are not represented as completed checks. The next repair priority is evidence-led security triage, not silently bypassing that gate.
