---
name: 'plan-implement-review'
description: 'Same as implement-review: run the plan-driven implement → review → fix → summary loop. Plans should be authored with Fable 5.1 high; review uses Grok 4.6 high fast. If the first review reports ≥1 Critical issue, run one extra review+fix round after those fixes.'
---

Read and follow the skill at `.cursor/skills/plan-implement-review/SKILL.md`, then execute its workflow on the currently approved plan.

Before doing anything else, verify the skill's preconditions:
1. An approved plan exists (from Plan mode / this conversation; plans should be authored with **Fable 5.1 high** / `claude-fable-5-1-thinking-high`). If not, stop and ask for one — and ask the user to draft it in Plan mode with Fable 5.1 high if they still need a plan.
2. You (the orchestrator) are running the `auto` model. State your current model; if it is not `auto`, stop and ask the user to switch before continuing.

Then proceed through the skill's steps: implement in the dev subagent (auto), review in a **Grok 4.6 high fast** subagent (`cursor-grok-4.6-high-fast`, /code-review), fix in the same dev subagent; if that first review had at least one Critical issue, run one more review+fix round; then write the final summary (fixed / not fixed / plan coverage / manual testing needed).
