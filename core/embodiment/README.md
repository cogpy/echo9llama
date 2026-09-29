# embodiment — GTAngelEcho link

echo9llama half of the Deep Tree Echo ⇄ GTAngel avatar loop. The Archecho desk
(`cogpy/GTAngelEcho`, `Archecho/archecho-desk/lib/echo9llama-link.js`) streams
`dte.embodiment/v1` frames; this package folds them into a smoothed echo of the
avatar's affect trajectory and replies with a reflection.

```
bridge.tick() ──frame──▶ POST /api/dte/embodiment ──▶ Hub.Ingest
      ▲                                                  │ EWMA echo (valence, arousal, flow)
      │                                                  │ gradient → flow attractor (0.6, 0.6, 0.8)
      └── triggerEvent(suggestedEvent) ◀──reflection─────┘ + systemPrompt for /api/chat
```

| Endpoint | Method | Result |
|---|---|---|
| `/api/dte/embodiment` | POST | ingest a frame → `Reflection` (422 on contract violation) |
| `/api/dte/embodiment` | GET | latest `Reflection` (404 before the first frame) |
| `/api/dte/embodiment/history` | GET | retained frames, oldest first |

Steering (circumplex quadrant of the smoothed echo → gameplay event):
tilted → `FLOW_STATE`, deflated → `EPIC_PLAY`, bored → `CLUTCH_MOMENT`,
engaged (low flow) → `FLOW_STATE`, flow → none.

`testdata/frame.json` is the contract fixture; a copy lives in the desk's
`test/fixtures/embodiment-frame.json` and both test suites check against it.
