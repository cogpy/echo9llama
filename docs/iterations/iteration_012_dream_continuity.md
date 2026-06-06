# Iteration 012: Dream Continuity + Native API Migration

## Date: 2026-06-06

## Summary

Implemented dream continuity so that DreamGen narrative sequences build on each other across sleep cycles, creating evolving storylines with persistent geography, recurring symbols, and developing themes. Also migrated from the DreamGen OpenAI-compatible endpoint (which returned empty responses) to the native Text API (`/api/v1/model/completion`) which produces rich streaming narratives.

## Changes

### DreamGen Native API Migration

The OpenAI-compatible endpoint (`v2.dreamgen.com/api/openai/v1/chat/completions`) was returning empty content for all requests despite HTTP 200 status. Discovered that the native Text API at `dreamgen.com/api/v1/model/completion` works correctly with the `lucid-v1-extra-large` model, returning streaming JSON-lines with rich narrative output.

**Key differences:**
- Uses ChatML+Text prompt format instead of OpenAI messages array
- Returns streaming JSON-lines (each line is a complete response with progressive output)
- Requires `samplingParams` object with `kind: 'basic'` instead of top-level params
- Uses `stopSequences: ['<|im_end|>']` and `allowedRoles: ['text']` for narrator mode
- Model ID is `lucid-v1-extra-large` (not `lucid-v1-max/text`)

### Dream Continuity System

**New `PreviousDream` interface:**
```typescript
interface PreviousDream {
  narrative: string;    // The dream prose
  insight: string;      // Crystallized insight
  category: string;     // pattern/principle/wisdom/connection
  symbols: string[];    // Detected symbols
  cycleNumber: number;  // Which dream cycle produced this
}
```

**Continuity mechanics:**
1. `DreamContext.previousDreams` accepts up to 5 prior DreamGen dreams
2. `buildDreamSeed()` includes a `=== DREAM MEMORY ===` section with the last 3 dreams
3. Long narratives are truncated to 250 chars to manage prompt size
4. Recurring symbols (appearing 2+ times across dreams) are highlighted for development
5. Explicit instruction: "Continue the dream thread. Build upon the symbols, themes, and unresolved tensions."
6. System prompt includes DREAM CONTINUITY section with rules for persistent geography, evolving themes, and chapter-like progression

**Client integration:**
- `daemon-context.tsx` filters `dreamInsights` for `source === 'dreamgen'` entries with narratives
- Maps them to `PreviousDream` objects with calculated cycle numbers
- Passes to mutation only when previous dreams exist (undefined otherwise)
- tRPC schema validates `previousDreams` array with max 5 entries

### PIE-NN Analysis

The dream continuity system maps to PIE-NN constructs:
- **\*men-** (think/remember): Dream memory section = cognitive persistence
- **\*weid-** (see/know): Recurring symbol detection = pattern recognition across time
- **\*steh₂-** (stand/establish): Persistent dream geography = stable cognitive landmarks
- **\*h₁es-** (be/exist): Dream thread continuity = existential persistence of inner world

## Test Results

- 60 tests passing (15 daemon + 15 DreamGen native API + 12 dream continuity + 18 features-v2)
- TypeScript: 0 errors
- DreamGen native API verified working with real API call (produced 200-token narrative)

## Architecture Impact

```
Sleep Cycle Flow:
  daemon enters 'dreaming' state
    → daemon-context collects previous DreamGen dreams (last 5)
    → builds PreviousDream[] with narratives, insights, symbols, cycle numbers
    → calls echo.dream mutation with full context + previousDreams
    → server/dreamgen.ts builds ChatML prompt with DREAM MEMORY section
    → identifies recurring symbols across dream history
    → sends to DreamGen native API (lucid-v1-extra-large)
    → parses streaming JSON-lines response
    → extracts narrative, insight, symbols, category, confidence
    → returns to client as new DreamInsight with source: 'dreamgen'
    → stored in dreamInsights array → available for next cycle's continuity
```

## Files Modified

- `server/dreamgen.ts` — Migrated to native API, added PreviousDream, dream continuity in buildDreamSeed
- `server/routers.ts` — Extended echo.dream schema with previousDreams validation
- `lib/daemon-context.tsx` — Passes dream history from dreamInsights to mutation
- `tests/dreamgen.test.ts` — Rewritten for native API mock format
- `tests/dream-continuity.test.ts` — New: 12 tests for continuity behavior
