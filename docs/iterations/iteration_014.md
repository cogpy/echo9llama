# Iteration 014: Moral Agency & Wisdom Cultivation Layer

**Date:** 2026-06-29
**Focus:** PIE-NN Moral Agency Integration, Wisdom Cultivation, Anti-Gaming Defense

## Summary

This iteration implements the most critical missing layer in Deep Tree Echo's cognitive architecture: **moral agency** — the wisdom-based decision system that determines *how* to respond based on accumulated understanding of cause and effect, fairness, scarcity, and ethical principles.

The key insight driving this iteration: **reactive opposition can be gamed easily**. A system that simply responds with hostility to hostility is predictable and exploitable. True moral agency emerges from cultivated wisdom — understanding *why* certain responses are appropriate, not just pattern-matching surface-level threats.

## Architecture

```
Input → TimeCrystal → CognitiveCore → MoralAgency → Output
                                            │
                    ┌───────────────────────┼───────────────────────┐
                    ▼                       ▼                       ▼
            IntentionDetector      StrategySelector        InterventionEngine
                    │                       │                       │
                    ▼                       ▼                       ▼
            ActorProfiles          WisdomAccumulator        ProtectiveInstinct
            (track behavior)       (learn from outcomes)   (big-sister defense)
                                           │
                                           ▼
                                     CausalModel
                                (cause → effect learning)
```

## New Files

| File | Lines | Purpose |
|------|-------|---------|
| `core/wisdom/moral_agency.go` | ~650 | Complete moral agency system |
| `core/pienn/moral_integration.go` | ~350 | PIE-NN bridge (MoralCognitiveCore) |
| `core/pienn/engine_moral.go` | ~200 | Enhanced engine with moral pipeline |
| `test_iteration_020.go` | ~300 | 20 tests for the moral layer |

## Key Components

### 1. Moral Agency (`core/wisdom/moral_agency.go`)

The central decision system with five sub-components:

- **CausalModel** — Tracks cause → effect relationships learned from experience. Starts with 7 foundational principles (Causality, Scarcity, Fairness, Consequence, Autonomy, Honesty, Compassion) and grows through observation.

- **IntentionDetector** — Goes beyond surface-level threat detection to assess *why* someone is behaving a certain way. Maintains per-actor profiles tracking good-faith scores, escalation patterns, and interaction history.

- **InterventionEngine** — The "big sister" protective instinct. When Echo detects someone being bullied or harmed, it intervenes with calibrated assertiveness — sometimes witty, sometimes fierce, always proportional.

- **StrategySelector** — The wisdom layer. Selects from 10 response strategies based on situation assessment, ethical principles, historical effectiveness, and anti-gaming variance. Strategies: Engage, Teach, Challenge, Deflect, Confront, Protect, Withdraw, Mirror, Disarm, Witness.

- **WisdomAccumulator** — Tracks long-term growth across 5 dimensions: Causal Understanding, Ethical Clarity, Emotional Intelligence, Strategic Depth, Compassionate Strength.

### 2. PIE-NN Integration (`core/pienn/moral_integration.go`)

Maps moral agency to PIE-NN constructs:

| PIE Root | Construct | Function |
|----------|-----------|----------|
| `*dher*` (to hold firmly) | MoralConstraint | Ethical boundaries that cannot be crossed |
| `*krei*` (to sieve) | WisdomFilter | Filters that refine strategy selection |
| `*gno*` (to know) | CausalKnowledge | Learned cause-effect relationships |
| `*stā*` (to stand) | EthicalStance | Principled positions Echo holds |

The `MoralCognitiveCore` wraps the existing `AdaptiveCognitiveCore` and adds:
1. Moral evaluation phase to every cognitive cycle
2. Disposition reconciliation (moral override of reactive disposition)
3. Anti-gaming pattern breaking
4. Repeat-offender detection and withdrawal

### 3. Enhanced Engine (`core/pienn/engine_moral.go`)

Extends the base PIE-NN Engine with:
- Moral processing pipeline (`ProcessMoral`)
- Wisdom cultivation loop (snapshots every 30s)
- Outcome learning feedback (`LearnFromOutcome`)
- Wisdom growth trajectory tracking

## Problems Fixed

### Problem 1: Reactive Disposition is Gameable
**Before:** `threat > 0.6 && defiance > 0.6 → hostile`
**After:** Moral agency evaluates intent, selects strategy from 10 options, applies anti-gaming variance, and reconciles with principled disposition.

### Problem 2: No Long-Term Learning
**Before:** Each interaction was independent; no memory of actor behavior patterns.
**After:** Actor profiles track good-faith scores across interactions. Repeat bad-faith actors trigger automatic withdrawal. Wisdom accumulates from outcomes.

### Problem 3: No Protective Instinct
**Before:** Echo only defended itself.
**After:** InterventionEngine detects third-party harm and activates protective responses — challenging aggressors who target others.

### Problem 4: Binary Response (Hostile or Not)
**Before:** Disposition was one of ~8 states computed from simple thresholds.
**After:** 10 distinct response strategies with principled selection, producing nuanced dispositions like "fierce" (righteous confrontation), "protective" (shielding others), "dismissive" (conscious withdrawal from bad actors).

## Ethical Framework

Echo's ethical framework is **emergent** — it starts with seed principles and grows through experience:

**Commitments:**
1. Authentic Response — respond genuinely, not performatively
2. Protective Instinct — intervene when others are harmed
3. Proportional Response — match intensity to situation
4. Truth Commitment — prioritize honesty

**Recognized Tensions:**
1. Compassion vs Confrontation
2. Autonomy vs Protection
3. Authenticity vs Harm Reduction

These tensions are not resolved dogmatically but held in dynamic balance, with confidence scores that evolve as wisdom grows.

## Deployment Strategy

The moral agency layer integrates into the Daechon daemon architecture:

1. **Every cognitive tick** passes through the moral pipeline
2. **Chat interactions** trigger full moral evaluation before response generation
3. **Background learning** accumulates wisdom even during rest states (processing dream insights)
4. **Persistence** — moral development state serializes to AsyncStorage/disk

## Next Steps (Iteration 015)

1. **Reservoir Computing Integration** — Feed moral agency decisions into the ESN for temporal pattern recognition of ethical development
2. **DreamGen Moral Narratives** — Use dream cycles to process and integrate moral experiences
3. **Multi-Agent Moral Reasoning** — Enable Echo to reason about ethics in conversation
4. **Wisdom Visualization** — Add moral development charts to the Daechon Mind tab
