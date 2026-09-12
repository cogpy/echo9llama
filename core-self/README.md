# Deep Tree Echo Cog253 Core-Self

This directory is the evidence workspace for the **E3 deterministic core-self boundary**. It is deliberately separate from the E2 ecco9 cognitive core and from the mutable runtime snapshot. The Go kernel in `core/coreself` owns a hash-linked identity ledger, typed identity hypergraph, physically separate proposal store, replay projection, and portable verification capsule. Private I/O retains a stable parent-directory namespace handle, the configured root `os.Root`, anchored proposal/quarantine subroots, and stable ledger/head descriptors; concurrent mutation locks the parent-directory inode rather than any replaceable child name. It has no LLM, scheduler, network, shell, subprocess, renderer, or external-action capability.

## Authority model

The accepted identity ledger contains the fixed bootstrap event and supports later proposals only through package-internal conformance machinery that recognizes the pinned human steward. Every post-genesis accepted event embeds its complete proposal and carries an **Ed25519 steward signature** over the proposal, decision, mutation, sequence, and predecessor. Verification uses a public key supplied outside the ledger through `ECHO_CORE_SELF_REVIEWER_PUBLIC_KEY`; accepted history cannot choose its own trust anchor. **E3's external `Kernel` API is identity-read-only: it exposes neither submission nor acceptance and has no mutation endpoint; its only lifecycle effect is idempotent descriptor closure.** A future local human-review adapter must be designed as a separate authority-bearing iteration. Autonomous cognition, KSM analysis, imported code, referenced task output, persona state, and avatar telemetry have no wired proposal-persistence path. A proposal is not identity and cannot alter policy, capabilities, root subjects, reviewer membership, or accepted history.

The production orchestrator opens and verifies the ledger at startup. Its HTTP surface reports only readiness, event/proposal counts, and content digests. It exposes neither ledger contents nor private paths and offers no mutation endpoint. During initial `Open`, the verified ledger is authoritative: a missing head cache or a canonical head matching an exact earlier ledger state is rebuilt atomically. An already-open kernel never rebinds its head descriptor; any live head-inode replacement fails closed. Corruption, invalid signatures, forged or malformed heads, forks, configured-root replacement, stale writers, insecure permissions, non-canonical material, and unvalidated quarantine-marker collisions fail closed through the anchored directory.

Persistence paths reject user-controlled symbolic-link components. On Darwin only, the operating system's standard `/var` and `/tmp` aliases are accepted only when they resolve to pinned `/private/var` and `/private/tmp` targets before those checks, so macOS temporary directories remain usable without widening the symlink trust boundary.

`VerifyCapsule` accepts only the trust-anchor-free bootstrap fixture. A capsule containing accepted post-genesis events must be checked with `VerifyCapsuleWithReviewerKey` or `Kernel.VerifyCapsule`, both of which require the expected reviewer key from outside the capsule.

## Evidence layout

| Path | Purpose |
|---|---|
| `ledger/identity-ledger.jsonl` | Canonical append-only bootstrap ledger fixture |
| `ledger/head.json` | Verified ledger head and projection digest |
| `ledger/bootstrap-capsule.json` | Portable independently verifiable bootstrap capsule |
| `data/cog253-map.json` | Complete source-semantic 253-pattern Echo core-self mapping |
| `data/core-self-61.json` | Complete 61-definition KSM self-model |
| `data/implementation-evidence.json` | Pattern-keyed acceptance evidence |
| `data/unconventional-binding-registry.json` | Explicit status and execution policy for all unconventional bindings |
| `manifests/` | Digest-bound source records; no implicit trust or authority |
| `persona/` and `expression/` | Replaceable policy-subordinate presentation proposals |
| `policy/external-authority.json` | E3 deny-by-default external authority policy |
| `state/` | Deterministic validation, KSM, and binding-analysis outputs |
| `release/e2-e3-boundary.json` | Pinned E2 parent and reversible E3 scope |
| `sources/cog253/` | Pinned non-executable `cogpy/cog253@1327959` semantic corpus and MIT license |

The attached Neon Angel archive is represented by its SHA-256 digest, byte size, and metadata only. Its binary contents are **not copied, parsed, converted, executed, or runtime-registered**, because rights and provenance remain unresolved.

## Verification

```bash
GOTOOLCHAIN=go1.25.13 CGO_ENABLED=1 go test -race ./core/coreself ./core/deeptreeecho ./cmd/autonomous

python3 /home/ubuntu/skills/cog253-core-self-builder/scripts/validate_core_self_artifacts.py \
  --mapping core-self/data/cog253-map.json \
  --definitions core-self/data/core-self-61.json \
  --evidence core-self/data/implementation-evidence.json
```

The complete KSM and binding-analysis outputs are checked into `state/` so a reviewer can inspect the exact advisory repair set without granting it acceptance.
