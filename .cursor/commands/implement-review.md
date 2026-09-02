---
name: 'implement-review'
description: 'Run the plan-driven implement → review → fix → summary loop. Use after a plan is approved to implement it in a dev subagent, review with an Opus 5 subagent (/code-review), fix findings, then summarize. If the first review reports ≥1 Critical issue, run one extra review+fix round after those fixes.'
---

Read and follow the skill at `.cursor/skills/plan-implement-review/SKILL.md`, then execute its workflow on the currently approved plan.

Before doing anything else, verify the skill's preconditions:
1. An approved plan exists (from Plan mode / this conversation). If not, stop and ask for one.
2. You (the orchestrator) are running the `auto` model. State your current model; if it is not `auto`, stop and ask the user to switch before continuing.

Then proceed through the skill's steps: implement in the dev subagent (auto), review in an Opus 5 subagent (/code-review), fix in the same dev subagent; if that first review had at least one Critical issue, run one more review+fix round; then write the final summary (fixed / not fixed / plan coverage / manual testing needed).
