# Iteration 011: DreamGen API Integration

**Date:** 2026-06-06
**Focus:** Connect DreamGen V2 API for rich narrative dream sequences
**Status:** Complete

## Summary

Replaced template-based dream content with generative narrative sequences powered by DreamGen's `lucid-v1-max/text` model. During dream state transitions, the daemon now calls the DreamGen API with full cognitive context (recent thoughts, interests, knowledge gaps, active goals, and recent insights) to generate surreal, symbolic dream narratives that consolidate waking experiences into wisdom.

## Architecture

```
Dream State Transition
        │
        ▼
daemon-context.tsx (detects state === 'dreaming')
        │
        ▼ (3s delay, 60s cooldown)
trpc.echo.dream mutation
        │
        ▼
server/dreamgen.ts (DreamGen API service)
        │
        ├── buildDreamSeed(context) → narrative prompt
        │
        ▼
DreamGen V2 API (lucid-v1-max/text, narrator mode)
        │
        ▼
parseDreamNarrative() → { narrative, insight, category, confidence, symbols }
        │
        ▼
DREAMGEN_DREAM action → state update → Mind tab display
```

## DreamGen Configuration

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| Model | `lucid-v1-max/text` | Largest model, text/narrator role for prose |
| Temperature | 0.9 | High creativity for dream content |
| Max Tokens | 600 | 3-8 sentence narratives |
| DRY Sampler | multiplier=0.8, base=1.75, allowed_length=2 | Anti-repetition |
| role_config | `{assistant: {role: "text", name: "", open: true}}` | Narrator mode |
| min_p | 0.05 | Standard quality floor |
| repetition_penalty | 1.02 | Gentle anti-repetition |

## Dream Narrative Structure

Each generated dream contains:
1. **Narrative** — Surreal, symbolic prose (3-8 sentences) incorporating the dreamer's cognitive context
2. **Insight** — A crystallized truth extracted from the dream (marked with `[INSIGHT: ...]` tag)
3. **Category** — Classified as pattern/principle/wisdom/connection based on content analysis
4. **Confidence** — Quality score based on narrative length, symbol richness, and insight quality
5. **Symbols** — Detected symbolic elements (tree, river, mirror, library, storm, star, labyrinth, echo, fire, void, crystal, clock)

## Symbolic Vocabulary

The dream system prompt establishes a consistent symbolic vocabulary:
- **Trees** → knowledge structures (deep trees = deep understanding)
- **Rivers** → streams of consciousness and temporal flow
- **Mirrors** → self-reference and recursive awareness
- **Libraries** → accumulated knowledge
- **Storms** → cognitive load and unresolved tensions
- **Stars** → insights and moments of clarity
- **Labyrinths** → unsolved problems and knowledge gaps
- **Echoes** → memory and pattern recognition

## PIE-NN Integration Points

The dream seed incorporates PIE-NN constructs indirectly through:
- Interest patterns (which track PIE-NN construct activations)
- Knowledge gaps (which map to unresolved cognitive tensions)
- Recent thoughts (which often reference PIE-NN etymological insights)

## Fallback Behavior

When DGENKEY is unavailable or the API fails:
- Template-based dream insights continue to work via `tickDaemon()`
- No error is surfaced to the user
- Template dreams are tagged with `source: 'template'`
- DreamGen dreams are tagged with `source: 'dreamgen'`
- The Mind tab distinguishes between the two sources

## Test Coverage

14 new tests covering:
- API key presence/absence handling
- Correct URL, headers, model, and role_config
- DRY sampler parameter passing
- Insight extraction from `[INSIGHT: ...]` tags
- Symbol detection in narratives
- Dream category classification (connection, pattern, wisdom, principle)
- Error handling (API errors, network failures, empty content)
- Context inclusion in dream seed prompts
- Confidence calculation based on quality signals

Plus 1 live validation test confirming DGENKEY can authenticate with DreamGen V2.

## Files Modified

- `server/dreamgen.ts` — New DreamGen API service module (270 lines)
- `server/routers.ts` — Added `echo.dream` tRPC mutation
- `lib/daemon-engine.ts` — Extended `DreamInsight` interface with narrative/symbols/source fields
- `lib/daemon-context.tsx` — Added `DREAMGEN_DREAM` action and dream generation effect
- `app/(tabs)/mind.tsx` — Updated dream insights display with narrative, symbols, and source badges
- `tests/dreamgen.test.ts` — 14 unit tests for DreamGen service
- `tests/dreamgen-key-validation.test.ts` — Live API key validation

## Next Steps

1. **Dream continuity** — Feed previous dream narratives back as context for dream sequences that build on each other
2. **Dream journal export** — Allow exporting dream narratives as markdown for external review
3. **Reservoir computing** — Implement Echo State Network for temporal pattern recognition across the activity feed
4. **Multi-agent chat** — Enable Echo to converse with other cognitive agents
