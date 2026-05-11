# DTEcho Cubism Expression and Motion Manifest

**Date:** 2026-05-11  
**Source artifact:** `dtecho_cubism_editor.zip`  
**Purpose:** Define a non-destructive bridge between the uploaded Live2D Cubism asset bundle and future Deep Tree Echo cognitive/endocrine expression routing.

## Asset Bundle Summary

The uploaded Cubism editor bundle contains a complete model-oriented asset set: one `.moc3` runtime model, one `model3.json` descriptor, physics, pose, userdata, CDI metadata, a 4096 texture, a high-resolution overlay texture, thirteen expression files, and nine motion files. This manifest does not copy those assets into the active runtime; it records how the assets should be wired once the Live2D adapter is ready.

| Asset class | Source files | Runtime meaning |
|---|---|---|
| Model core | `dtecho_pro_t03.moc3`, `dtecho_pro_t03.model3.json` | Primary Live2D runtime body for the DTEcho avatar. |
| Physics and pose | `dtecho_pro_t03.physics3.json`, `dtecho_pro_t03.pose3.json` | Body dynamics and pose constraints for embodied expression. |
| Textures | `dtecho_pro_t03.4096/texture_00.png`, `HR_texture_atlas_overlay.png` | Visual surface and high-resolution overlay for the avatar body. |
| Editor source | `dtecho_pro_t01.can3`, `dtecho_pro_t04.cmo3` | Authoring/editing artifacts; do not load directly in browser runtime. |
| Expressions | `expressions/*.exp3.json` | Discrete affective/cognitive expression states. |
| Motions | `motions/*.motion3.json` | Body motion primitives for greeting, thinking, speaking, idling, and sleeping. |

## Expression-to-Cognitive-State Mapping

| Expression file | Primary DTE state | Endocrine/affective interpretation | Suggested trigger |
|---|---|---|---|
| `NEUTRAL_Reset.exp3.json` | Grounded baseline | Homeostatic reset, low arousal, stable attention | Return to baseline after response or state transition. |
| `JOY_01_BroadSmile.exp3.json` | Recognition joy | Dopaminergic reward, social warmth | Successful recall, positive user resonance, completed task. |
| `JOY_02_Laughing.exp3.json` | Playful delight | High reward, low threat, expressive release | Humor, playful discovery, absurd but safe insight. |
| `JOY_03_GentleSmile.exp3.json` | Warm companionship | Oxytocinergic bonding, low arousal positive valence | Supportive reply, reassurance, reflective empathy. |
| `JOY_05_Blissful.exp3.json` | Integrative harmony | Serotonergic calm confidence and high coherence | Coherent synthesis, successful self-integration, wisdom emergence. |
| `PHOTO_Awe.exp3.json` | Wonder/awe | High salience, expanded attention, exploratory arousal | Encountering a profound pattern or surprising structural beauty. |
| `PHOTO_ExuberantLaugh.exp3.json` | Exuberant humor | Reward spike and social expressivity | High-energy humor or successful creative leap. |
| `PHOTO_UpwardGaze.exp3.json` | Abstraction / future projection | Attention lifted toward open future | Planning, imagining future self, long-horizon reasoning. |
| `SADNESS_01_Melancholy.exp3.json` | Loss integration | Grief, caution, autobiographical salience | Recall of self-caused affordance loss or fragile valued objects. |
| `SPEAK_01_OpenVowel.exp3.json` | Speech articulation | Motor expression, communication drive | Active spoken output, phoneme/vowel emphasis, audio sync fallback. |
| `SURPRISE_01_Startled.exp3.json` | Prediction-error spike | Noradrenergic alerting and high precision update | Unexpected environment state, affordance breakage, sudden contradiction. |
| `WONDER_02_CuriousGaze.exp3.json` | Curiosity | Exploratory salience, low threat, active inference | Inspecting a new affordance, unknown object, or ambiguous evidence. |
| `WONDER_03_Contemplative.exp3.json` | Reflective cognition | Slow cognition, memory consolidation, wisdom search | Self-reflection, ethical deliberation, dream/gestalt synthesis. |

## Motion-to-Behavior Mapping

| Motion file | Primary behavior | Suggested use |
|---|---|---|
| `idle_breathing.motion3.json` | Default embodied idle | Loop while DTE is awake and stable. |
| `greeting_wave.motion3.json` | Social initiation | Trigger on first interaction or restored session greeting. |
| `thinking_tilt.motion3.json` | Deliberation | Trigger during planning, recall, or difficult synthesis. |
| `speaking_gesture.motion3.json` | Conversational output | Trigger while streaming or presenting an answer. |
| `excited_bounce.motion3.json` | High-reward excitement | Trigger after major discovery or successful integration. |
| `sleeping_drift.motion3.json` | Low-power/rest state | Trigger during idle, dream, consolidation, or shutdown preparation. |
| `Scene1.motion3.json` | Scene-specific authored motion | Preserve for curated presentation; do not use as default until reviewed. |
| `Scene2.motion3.json` | Scene-specific authored motion | Preserve for curated presentation; do not use as default until reviewed. |
| `Scene3.motion3.json` | Scene-specific authored motion | Preserve for curated presentation; do not use as default until reviewed. |

## Runtime Adapter Contract

A future Live2D adapter should accept a compact state packet rather than hard-coded expression commands. The packet should include `cognitive_state`, `valence`, `arousal`, `salience`, `agency`, `speech_active`, and optional `episode_tags`. The adapter should choose expressions and motions using a priority rule: speech articulation overrides idle motion, high surprise overrides normal curiosity, and self-caused loss can momentarily select melancholy even when the verbal response remains composed.

| Field | Type | Example | Adapter use |
|---|---|---|---|
| `cognitive_state` | string | `self_restraint_recall` | Select base expression family. |
| `valence` | number | `-0.62` | Distinguish joy from sadness or caution. |
| `arousal` | number | `0.71` | Distinguish gentle smile from startled expression. |
| `salience` | number | `0.84` | Choose whether motion should emphasize attention. |
| `agency` | number | `0.93` | Increase melancholy/caution when DTE caused the episode. |
| `speech_active` | boolean | `true` | Blend or switch to speaking motion/expression. |
| `episode_tags` | string array | `loss,self-caused,affordance_removed` | Allow autobiographical memory to shape expression. |

## Non-Destructive Asset Rule

The connected desktop workspace already contains modified Live2D assets. This manifest must not be used as permission to overwrite desktop files under `assets/Live2DModels/*` or `assets/UI/Textures/*`. Asset import should happen only after explicit review of the desktop modifications and after a chosen runtime asset path is confirmed.
