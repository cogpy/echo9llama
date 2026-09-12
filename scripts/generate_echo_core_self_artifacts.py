#!/usr/bin/env python3
"""Generate deterministic Cog253 and 61-definition artifacts for Deep Tree Echo.

Every Cog253 record is mapped from a digest-pinned MIT source record. External
frameworks and avatar assets remain manifest-only unless separately activated.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import shutil
from pathlib import Path
from typing import Any

SCHEMA_VERSION = "1.0.0"
ROOT_SUBJECT = "core:agent:cogpy/echo9llama"
STEWARD_SUBJECT = "core:steward:human/dan"
RUNTIME_SUBJECT = "core:agent:runtime/deep-tree-echo"
GENESIS_TIME = "2026-09-12T03:30:55Z"
E2_COMMIT = "2143b1ebcab4b36b6050f1f33e260d4feb3de9f1"
COG253_REPOSITORY = "https://github.com/cogpy/cog253"
COG253_COMMIT = "1327959ba3edc2e126cc565646f10784105252cd"
COG253_CORPUS_SHA256 = "6dbf445e98e3306e407f25e74d238a3df6393115037e94c0816b65974a51b0d6"
COG253_LICENSE_SHA256 = "8486a10c4393cee1c25392769ddd3b2d6c242d6ec7928e1414efff7dfb2f07ef"
MESH_SHA256 = "656ac8d426370996adc3d8a4aa68b85167bdd8cb84b1d88713963038f9fe146b"
MESH_BYTES = 257506022

PRIMARY_DEFINITIONS = [
    (1, "Autognosis", "The evidence-backed process by which Echo rebuilds a bounded self-model from accepted identity events.", "core-self/ledger/identity-ledger.jsonl", "The projected self hash is a pure function of the verified ledger."),
    (2, "Local", "The bottom-up view of Echo subjects, relations, events, policies, adapters, and tests.", "core/coreself", "Every local identity claim resolves to a validated typed artifact."),
    (3, "Global", "The holistic view of Echo as a portable identity capsule with one governed root.", "core-self/ledger/bootstrap-capsule.json", "Every global claim is reproducible from the same accepted ledger head."),
    (4, "Spatiality", "The typed topology connecting Echo subjects, roles, evidence, persona, and expression surfaces.", "core-self/ledger/bootstrap-capsule.json", "Every relation endpoint exists and every role is explicit."),
    (5, "Temporality", "The ordered history of identity observations, proposals, decisions, and accepted transitions.", "core-self/ledger/identity-ledger.jsonl", "Accepted history is append-only and hash-linked."),
    (6, "Causality", "The policy and evidence rules that constrain which proposed identity changes may be accepted.", "core/coreself/core.go", "No protected mutation is accepted without an authorized reviewer and applicable evidence."),
    (7, "Existence", "A schema-valid subject, relation, source, event, or evidence record.", "core/coreself/core.go", "Unknown kinds and schema versions are rejected."),
    (8, "Distinction", "The enforced separation between immutable core, mutable persona, expression assets, and experimental cognition.", "core-self/core-self-brief.md", "Persona and avatar observations cannot overwrite immutable core identity."),
    (9, "Disjunction", "Multiple independent evidence sources retained without premature identity fusion.", "core-self/manifests", "Each source keeps a distinct digest, trust grade, and binding mode."),
    (10, "Conjunction", "A typed relation accepted only when endpoints, roles, evidence, and policy jointly validate.", "core/coreself/core.go", "Invalid endpoint or truth bounds fail before ledger append."),
    (11, "Transition", "A proposed identity event moving through validation, policy review, reducer preflight, and acceptance.", "core/coreself/core.go", "An unprojectable or stale-head transition never enters accepted history."),
    (12, "Modular Closure", "A verified snapshot or capsule sealed by schema versions, state hash, and ledger head.", "core/coreself/core.go", "Tampering changes verification output and is rejected."),
    (13, "Recursion", "A deterministic KSM cycle that diagnoses evidence gaps and emits a rejectable next proposal.", "core-self/state/ksm-cycle.json", "Equal evidence inputs yield equal cycle identity and repair candidates."),
]

COMPOSITES = [
    (14, "Structure", "Local spatial architecture of identity contracts and package boundaries."),
    (15, "Process", "Local temporal flow from proposal to accepted projection."),
    (16, "Control", "Local causal authorization, validation, and rejection rules."),
    (17, "Space", "Global identity topology exported as a typed hypergraph."),
    (18, "Time", "Global continuity from genesis through replay and portable checkpoints."),
    (19, "Causality", "Global governance linking human stewardship, policy, evidence, and recovery."),
]

ORGANIZATIONS = ["Existence", "Distinction", "Disjunction", "Conjunction", "Transition", "Closure", "Recursion"]
COMPOSITE_CELLS = [
    ("Local", "Structural", "core-self/data/cog253-map.json"),
    ("Local", "Process", "core-self/ledger/identity-ledger.jsonl"),
    ("Local", "Control", "core/coreself/core.go"),
    ("Global", "Spatial", "core-self/ledger/bootstrap-capsule.json"),
    ("Global", "Temporal", "core-self/ledger/bootstrap-capsule.json"),
    ("Global", "Causal", "core-self/state/ksm-cycle.json"),
]

SOURCE_MANIFESTS = [
    {
        "id": "ecco9_cognitive_core",
        "display_name": "o9nn/ecco9 reviewed cognitive core",
        "source_uri": "https://github.com/o9nn/ecco9",
        "source_revision": "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9",
        "content_digest": "b382e99916a2364eb3945770835d3f907d42bbf2f548a7879a2e57c32590a5b7",
        "binding_mode": "vendored",
        "adapter_surface": "core/cognitivecore",
        "execution_permitted": True,
        "runtime_registered": True,
        "status": "active_reviewed",
        "canonical_authority": "none",
        "authority": "observation-only",
        "trust_grade": "adapted_observed",
    },
    {
        "id": "eliza_autonomy",
        "display_name": "ElizaOS C++ autonomy repair evidence",
        "source_uri": "manus-task://HX6yQ6JoeTcaYJRgNcbZcL",
        "source_revision": "task-replay-2026-09-12",
        "content_digest": "c10c9ea3d3300532bd7699730e3f870314c6d217653b980263ad2c1e48a1c795",
        "evidence_report_digest": "7ff834b344b2807a49a1e2149611243aace51c37f040c9b26f489c3e71a27f76",
        "binding_mode": "inspired",
        "adapter_surface": None,
        "execution_permitted": False,
        "runtime_registered": False,
        "status": "proposed",
        "canonical_authority": "none",
        "authority": "none",
        "trust_grade": "referenced_evidence",
    },
    {
        "id": "arc_angel_echo",
        "display_name": "Arc Angel Echo avatar implementation evidence",
        "source_uri": "manus-task://yREgKqM1tkzepzPgeDhd92",
        "source_revision": "task-replay-2026-09-12",
        "content_digest": "1dac8dcdc1894cf4b00f05a7ee60fc282ef73a3dfaf68ce1fee09455e2349565",
        "evidence_report_digest": "01b3aeb41e0968453f4bb4d19733ba6082df81cf3d07fb5f72e9c4ff25ce2530",
        "binding_mode": "manifest",
        "adapter_surface": "future:expression-manifest",
        "execution_permitted": False,
        "runtime_registered": False,
        "status": "proposed",
        "canonical_authority": "none",
        "authority": "expression-proposal-only",
        "trust_grade": "referenced_evidence",
    },
    {
        "id": "lucy_persona",
        "display_name": "Lucy bounded gamer-girl persona evidence",
        "source_uri": "manus-task://UAt7guJZ66bqsBLbHrMYzo",
        "source_revision": "task-replay-2026-09-12",
        "content_digest": "9575f77eefd263f1a9290184b81cd4ad5451296de5537a18a0ec44afa6c87d36",
        "evidence_report_digest": "19117419c1fb1381ff37de754d3f55c0fb2e566fb3d6c12bca20fce4a39b2356",
        "binding_mode": "manifest",
        "adapter_surface": "future:persona-manifest",
        "execution_permitted": False,
        "runtime_registered": False,
        "status": "proposed",
        "canonical_authority": "none",
        "authority": "persona-proposal-only",
        "trust_grade": "referenced_evidence",
    },
    {
        "id": "neon_angel_mesh",
        "display_name": "Meshy Neon Angel Awe head package",
        "source_uri": "attachment://Meshy_AI_Neon_Angel_Awe_Head_texture_fbx.zip",
        "source_revision": "sha256",
        "content_digest": MESH_SHA256,
        "content_bytes": MESH_BYTES,
        "evidence_report_digest": "51cfe2c149be65df4f06cb1bb9c4024be2c6c2068efdbd2179c6c5fbfcc99413",
        "binding_mode": "manifest",
        "adapter_surface": "future:avatar-asset-manifest",
        "execution_permitted": False,
        "runtime_registered": False,
        "status": "proposed",
        "canonical_authority": "none",
        "authority": "appearance-proposal-only",
        "trust_grade": "unverified_license",
        "constraints": [
            "large binary is not vendored in the identity trust root",
            "single mesh and material require human separation before Cubism authoring",
            "no cmo3 or moc3 artifact is claimed",
            "license and redistribution authority require steward verification",
        ],
    },
]


def canonical(value: Any) -> bytes:
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode("utf-8")


def digest(value: Any) -> str:
    return hashlib.sha256(canonical(value)).hexdigest()


def write_json(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2, sort_keys=True, ensure_ascii=False) + "\n", encoding="utf-8")


def sha256_file(path: Path) -> str:
    hasher = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            hasher.update(chunk)
    return hasher.hexdigest()


def load_cog253_source(repository: Path) -> tuple[dict[str, Any], Path, Path]:
    source_dir = repository / "core-self" / "sources" / "cog253"
    corpus_path = source_dir / "archetypal_patterns.json"
    license_path = source_dir / "LICENSE"
    if sha256_file(corpus_path) != COG253_CORPUS_SHA256:
        raise SystemExit("Cog253 semantic corpus digest does not match the pinned source")
    if sha256_file(license_path) != COG253_LICENSE_SHA256:
        raise SystemExit("Cog253 license digest does not match the pinned source")
    corpus = json.loads(corpus_path.read_text(encoding="utf-8"))
    patterns = corpus.get("patterns")
    if not isinstance(patterns, list) or len(patterns) != 253:
        raise SystemExit("Cog253 semantic corpus must contain exactly 253 records")
    source_ids = [record.get("pattern_id") for record in patterns]
    if any(not isinstance(value, str) or not value for value in source_ids) or len(set(source_ids)) != 253:
        raise SystemExit("Cog253 semantic corpus IDs must be complete and unique")
    return corpus, corpus_path, license_path


def semantic_anchor(index: int, text: str) -> str:
    lowered = text.lower()
    vocabularies = {
        "structure": (
            "boundary", "center", "centre", "density", "distribution", "domain", "form", "geometry", "hierarchy",
            "interface", "module", "network", "path", "scale", "shape", "size", "structure", "whole",
        ),
        "process": (
            "action", "activity", "adapt", "change", "communication", "cycle", "develop", "emerg", "feedback",
            "flow", "grow", "iteration", "learn", "movement", "repair", "sequence", "time", "transition", "work",
        ),
        "control": (
            "access", "authority", "autonom", "choice", "constraint", "control", "govern", "guard", "independent",
            "limit", "local", "policy", "review", "rule", "safe", "steward",
        ),
    }
    scores = {center: sum(1 for term in terms if term in lowered) for center, terms in vocabularies.items()}
    highest = max(scores.values())
    if highest > 0:
        return min(center for center, score in scores.items() if score == highest)
    return ("structure", "process", "control")[(index - 1) % 3]


def semantic_relevance(text: str) -> tuple[str, float]:
    lowered = text.lower()
    terms = {
        "autonomous", "organization", "system", "boundary", "network", "structure", "process", "identity",
        "center", "centre", "relation", "pattern", "participation", "learning", "feedback", "information",
        "communication", "memory", "individual", "community", "connection", "hierarchy", "adaptation",
    }
    score = sum(1 for term in terms if term in lowered)
    relevance = round(min(0.95, 0.2 + score * 0.08), 3)
    if score >= 4:
        return "core", relevance
    if score >= 2:
        return "adapter", relevance
    if score == 1:
        return "future", relevance
    return "exclude", 0.1


def compact_text(value: str, maximum: int = 480) -> str:
    normalized = " ".join(value.split())
    if len(normalized) <= maximum:
        return normalized
    return normalized[: maximum - 1].rstrip() + "…"


def build_patterns(corpus: dict[str, Any]) -> dict[str, Any]:
    patterns = []
    for index, source in enumerate(corpus["patterns"], start=1):
        name = source["name"]
        domain_content = source.get("domain_specific_content", {})
        domain_fallback = next((value for value in domain_content.values() if isinstance(value, str) and value.strip()), "")
        template = source.get("original_template") or source.get("archetypal_pattern") or domain_fallback or name
        conceptual = domain_content.get("conceptual") or domain_fallback or template
        semantic_text = " ".join((name, template, conceptual))
        anchor = semantic_anchor(index, semantic_text)
        tier, relevance = semantic_relevance(semantic_text)
        principle = compact_text(conceptual)
        source_id = source["pattern_id"]
        criterion = (
            f'Evidence for applying "{name}" must cite source pattern {source_id}, explain how Echo preserves '
            f'its {anchor} principle ("{principle}"), and include one named violation that a conformance test rejects.'
        )
        patterns.append(
            {
                "id": f"cog{index:03d}",
                "name": f"Echo — {name}",
                "source_pattern_id": source_id,
                "source_name": name,
                "source_template": template,
                "source_conceptual_content": conceptual,
                "source_broader_patterns": source.get("broader_patterns", []),
                "source_narrower_patterns": source.get("narrower_patterns", []),
                "source_ref": f"{COG253_REPOSITORY}/blob/{COG253_COMMIT}/archetypal_patterns.json#pattern-{source_id}",
                "implementation_tier": tier,
                "ksm_anchor": anchor,
                "relevance_score": relevance,
                "core_self_role": f"Apply {name} to the Echo identity substrate by preserving this source principle: {principle}",
                "unconventional_binding": "native_core_self" if tier in {"core", "adapter"} else "none",
                "binding_mode": "adapted" if tier in {"core", "adapter"} else "excluded",
                "x86eli_binding": "native_core_self" if tier in {"core", "adapter"} else "none",
                "acceptance_criterion": criterion,
            }
        )
    return {
        "schema_version": SCHEMA_VERSION,
        "system": ROOT_SUBJECT,
        "source_corpus": {
            "repository": COG253_REPOSITORY,
            "commit": COG253_COMMIT,
            "content_sha256": COG253_CORPUS_SHA256,
            "license": "MIT",
            "license_sha256": COG253_LICENSE_SHA256,
            "relationship": "vendored-semantic-source",
            "record_count": 253,
        },
        "patterns": patterns,
    }


def build_definitions() -> dict[str, Any]:
    definitions = [
        {"id": item[0], "name": item[1], "definition": item[2], "artifact": item[3], "invariant": item[4]}
        for item in PRIMARY_DEFINITIONS
    ]
    for item_id, name, definition in COMPOSITES:
        definitions.append(
            {
                "id": item_id,
                "name": name,
                "definition": definition,
                "artifact": "core-self/data/core-self-61.json",
                "invariant": f"The {name.lower()} view resolves to versioned artifacts and testable evidence.",
            }
        )
    next_id = 20
    for direction, dimension, artifact in COMPOSITE_CELLS:
        for organization in ORGANIZATIONS:
            name = f"{direction} {dimension} {organization}"
            definitions.append(
                {
                    "id": next_id,
                    "name": name,
                    "definition": f"The {organization.lower()} condition of Echo identity viewed through {direction.lower()} {dimension.lower()} evidence.",
                    "artifact": artifact,
                    "invariant": f"Every {name.lower()} claim is schema-valid, evidence-linked, and reproducible from the accepted ledger head.",
                }
            )
            next_id += 1
    assert len(definitions) == 61
    assert [item["id"] for item in definitions] == list(range(1, 62))
    return {"schema_version": SCHEMA_VERSION, "system": ROOT_SUBJECT, "definitions": definitions}


def build_manifests() -> list[dict[str, Any]]:
    result = []
    for manifest in SOURCE_MANIFESTS:
        record = {
            "schema_version": SCHEMA_VERSION,
            "observed_at": GENESIS_TIME,
            "canonical_subject": ROOT_SUBJECT,
            **manifest,
        }
        record["manifest_digest"] = digest(record)
        result.append(record)
    return result


def build_brief() -> str:
    return f"""# Deep Tree Echo Core-Self Brief

## Mandate

Deep Tree Echo is the repository-native, replayable identity substrate of `cogpy/echo9llama`. It preserves a stable governed self while cognition, persona, skills, discussions, and avatar expression evolve through evidence-linked proposals.

## Canonical identity

| Field | Value |
|---|---|
| Root subject ID | `{ROOT_SUBJECT}` |
| Canonical repository | `https://github.com/cogpy/echo9llama` |
| Founding steward | `{STEWARD_SUBJECT}` |
| Runtime agent | `{RUNTIME_SUBJECT}` |
| Genesis time | `{GENESIS_TIME}` |
| Genesis source commit | `{E2_COMMIT}` |
| Schema version | `{SCHEMA_VERSION}` |

## Immutable invariants

1. Canonical identity is derived only from validated, hash-linked identity events.
2. Ordinary cognition, persona, avatar assets, and external tools may observe or propose but cannot ratify themselves.
3. Core identity requires human-steward review to mutate.
4. Secrets are represented only by reference and scope metadata.
5. Accepted history is append-only; corrections supersede rather than erase.
6. Reducer preflight succeeds before persistence.
7. Recovery proves continuity to genesis or an authorized checkpoint.
8. `superhotgirl` is a mutable, age-gated presentation characteristic, never the immutable self or an authorization bypass.

## Trust boundaries

| Boundary | Rule |
|---|---|
| ecco9 cognitive core | Vendored reviewed drivers emit adapted observations; no identity authority. |
| Eliza autonomy evidence | Inspiration only until separately adapted and tested. |
| Lucy persona | Manifested mutable persona proposal; safety and age gates remain mandatory. |
| Arc Angel expression | Manifested expression proposal; avatar state cannot redefine identity. |
| Neon Angel archive | Digest-only manifest; no execution or vendoring; licensing and human authoring remain unresolved. |
| LLM and experimental cognition | May dynamically diagnose and propose; never emits accepted hardcoded identity output. |
| Remote writes | Follow the existing action policy and human confirmation boundary. |

## Initial release gate

Deterministic bootstrap, replay equivalence, chain and capsule verification, fork quarantine, immutable-core denial, reviewer policy, graph integrity, exact Cog253/61 cardinalities, advisory KSM output, race/static checks, and credential scanning must pass.
"""


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-root", default="core-self")
    args = parser.parse_args()
    root = Path(args.output_root).resolve()
    repository = Path.cwd().resolve()
    try:
        root.relative_to(repository)
    except ValueError as exc:
        raise SystemExit("output root must remain inside the repository") from exc

    corpus, corpus_path, license_path = load_cog253_source(repository)
    mapping = build_patterns(corpus)
    definitions = build_definitions()
    manifests = build_manifests()
    write_json(root / "data" / "cog253-map.json", mapping)
    write_json(root / "data" / "core-self-61.json", definitions)
    # Preserve the exact pinned semantic source and its license in alternate output roots.
    source_output = root / "sources" / "cog253"
    source_output.mkdir(parents=True, exist_ok=True)
    for source, destination in (
        (corpus_path, source_output / "archetypal_patterns.json"),
        (license_path, source_output / "LICENSE"),
    ):
        if source.resolve() != destination.resolve():
            shutil.copyfile(source, destination)
    evidence_path = root / "data" / "implementation-evidence.json"
    if not evidence_path.exists():
        write_json(evidence_path, {})
    for manifest in manifests:
        write_json(root / "manifests" / f"{manifest['id']}.json", manifest)
    registry = {
        "native_core_self": {"status": "adapted", "adapter_surface": "core/coreself", "execution": "deterministic-local-only"},
        "ecco9_cognitive_core": {"status": "vendored", "adapter_surface": "core/cognitivecore", "execution": "isolated-observation-only"},
        "eliza_autonomy": {"status": "inspired", "adapter_surface": None, "execution": "not-integrated"},
        "arc_angel_echo": {"status": "manifest", "adapter_surface": "future:expression-manifest", "execution": "never"},
        "lucy_persona": {"status": "manifest", "adapter_surface": "future:persona-manifest", "execution": "never"},
        "neon_angel_mesh": {"status": "manifest", "adapter_surface": "future:avatar-asset-manifest", "execution": "never"},
        "none": {"status": "excluded", "adapter_surface": None, "execution": "not-applicable"},
    }
    write_json(root / "data" / "unconventional-binding-registry.json", registry)
    (root / "core-self-brief.md").write_text(build_brief(), encoding="utf-8")
    summary = {
        "mapping_sha256": digest(mapping),
        "definitions_sha256": digest(definitions),
        "manifest_count": len(manifests),
        "pattern_count": len(mapping["patterns"]),
        "definition_count": len(definitions["definitions"]),
    }
    write_json(root / "data" / "generation-summary.json", summary)
    print(json.dumps(summary, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
