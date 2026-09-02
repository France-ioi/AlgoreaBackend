---
name: plan-implement-review
description: Orchestrates a plan-driven implementation loop with a built-in code review for the AlgoreaBackend (Go) project. Use after a plan has been approved (Plan mode) and the user wants the change implemented by a dev subagent, reviewed by an Opus 5 subagent running the /code-review skill, fixed by the dev subagent, then summarized. When the first review reports at least one Critical issue, run one extra review+fix round after those fixes. Trigger when the user asks to "implement and review", "build then review", or run the implement → review → fix → summary workflow.
---

# Plan, Implement, Review

## Overview

This skill turns an approved plan into shipped code through a delegated loop: a dev subagent implements, an Opus 5 review subagent runs `/code-review`, the same dev subagent fixes the findings, and the orchestrator reports what was and wasn't done. If the first review reports at least one Critical issue, after the first fix pass there is exactly one more review+fix round before the final summary.

The orchestrator (the agent running this skill) stays lightweight: it coordinates subagents and writes the final summary. It does NOT implement or review the code itself.

## Preconditions (verify before any work)

Stop and ask the user if either is missing — do not start implementing.

1. **An approved plan exists.** The plan must come from Cursor Plan mode / an approved plan in the current conversation. If there is no clear, approved plan, ask the user to provide or approve one first. Restate the plan as a numbered list of concrete deliverables so completion can be checked later.
2. **The orchestrator is running the `auto` model.** State which model you are currently on. If it is not `auto`, stop and ask the user to switch to the `auto` model before continuing — this workflow is cost-optimized: cheap `auto` orchestration delegates expensive work to subagents. Do not proceed until on `auto`.

## Workflow checklist

Copy this and keep it updated as you go:

```
- [ ] Step 0: Preconditions verified (approved plan + orchestrator on `auto`)
- [ ] Step 1: Implement the plan in the dev subagent (auto model)
- [ ] Step 2: Review the changes in an Opus 5 subagent (/code-review)
- [ ] Step 3: Fix review findings in the SAME dev subagent
- [ ] Step 3a: If Step 2 had ≥1 Critical → second review (Opus 5) + second fix (same dev) [skip if no Critical]
- [ ] Step 3b: Orchestrator independently verifies coverage on all modified functions
- [ ] Step 4: Write the final summary
```

## Step 1: Implement in the dev subagent

Launch ONE dev subagent to do the implementation. Reuse this same subagent later for fixes (keep its agent ID).

- Tool: `Task` with `subagent_type: "generalPurpose"`.
- Model: **do not pass a `model`** — omitting it makes the subagent inherit the orchestrator's `auto` model, satisfying the "both on auto" requirement.
- Prompt: subagents do not see the conversation, so include the **full plan** (all numbered deliverables), the relevant file paths, and the project conventions reminder: "follow `AGENTS.md` and `ARCHITECTURE.md`: idiomatic Go and backend best practices; keep `ARCHITECTURE.md` updated when the architecture changes; only add comments that explain non-obvious intent/trade-offs (no narrating comments); DB migrations go in `db/migrations/` via goose using the `YYMMDDHHMM_description.sql` naming and never hand-edit `db/schema/schema.sql`; bump `EventVersion` in `app/event/event.go` when changing event schemas". Instruct it to run the linter (`./bin/golangci-lint run -v --timeout 2m`) and fix lint errors before returning, and to run relevant tests (`make test-dev`) when the change warrants it.
- **Coverage (mandatory in the Step 1 prompt):** per `AGENTS.md`, every **modified function** in every changed `.go` file must report **100%** in `go tool cover -func` (not only brand-new helpers). Extracting a helper and testing it directly does **not** cover the production call site — if a caller gained a new branch/`return err`, that caller path must be exercised too (e.g. via `New`/`Reset`, not only the helper). Ask the subagent to return the `go tool cover -func | grep` lines for each touched file.
- Ask the subagent to return: a list of changed files, which plan deliverables it completed, anything it could not do (with the reason), and the coverage excerpt above.

Record the dev subagent's **agent ID** — you will `resume` it in Step 3.

## Step 2: Review in an Opus 5 subagent

Launch a separate review subagent over the code that was just written.

- Tool: `Task` with `subagent_type: "generalPurpose"`, `readonly: true`.
- Model: **`claude-opus-5-thinking-high`** (Opus 5). If that slug is unavailable, tell the user Opus is unavailable rather than silently substituting another model.
- Prompt the review subagent to:
  1. Read and follow the project's `/code-review` skill/command.
  2. Scope the review to the just-written changes: run `git diff` (and `git status`) to find modified/untracked files, then read them.
  3. **Re-run coverage on changed packages** and list any modified function below 100%, plus any uncovered **line ranges** in the diff (file:line). Treat an uncovered new call-site branch as a finding even if a helper it calls is 100%.
  4. Produce its report using that skill's output template (e.g. Critical / Suggestions / Positive).
  5. Save the report to `reviews/` (e.g. `reviews/<short-feature-name>-review.md`) and also return it.

## Step 3: Fix in the same dev subagent

`resume` the Step 1 dev subagent (pass its agent ID) with the review report. Instruct it to:

- **Fix every Critical issue.**
- **Address Suggestions that are relevant** — apply the ones that genuinely improve the code; for any suggestion it deliberately skips, record a one-line reason. Do not blindly apply suggestions that conflict with the plan or project conventions.
- Re-run the linter (`./bin/golangci-lint run -v --timeout 2m`) and tests (`make test-dev`) when the change warrants it, and fix failures.
- **If the review cited uncovered line ranges:** fix by exercising those exact paths; then confirm in the cover profile that each cited block has count > 0. Do not claim a coverage finding is fixed solely because a related helper is now 100%.
- Return: which Critical issues were fixed, which Suggestions were applied vs. skipped (with reasons), and (when coverage was in scope) the cover-func lines for every modified function plus confirmation of any previously uncovered ranges.

### Step 3a: Extra review+fix when Step 2 had Critical issues

**Trigger:** Step 2's review report listed **at least one Critical** issue. If Step 2 had zero Critical findings, skip this step entirely (even if there were Suggestions).

When triggered, run **exactly one** additional review+fix round — do not loop further even if the second review still finds Critical issues.

1. **Second review:** Launch a new Opus 5 review subagent the same way as Step 2 (`generalPurpose`, `readonly: true`, model `claude-opus-5-thinking-high`, `/code-review`). Scope it to the post-fix diff. Ask it to note which Step 2 Critical items are resolved vs. still open, and to report any new Critical / Suggestions. Save under `reviews/` (e.g. `reviews/<short-feature-name>-review-round2.md`) and return the report.
2. **Second fix:** `resume` the same Step 1 dev subagent with the round-2 report. Same fix rules as Step 3 (fix every Critical; address relevant Suggestions; re-lint/re-test as warranted; coverage paths if cited). Return what was fixed vs. skipped.

If the second review still reports Critical issues after the second fix, do **not** start a third round — list those remaining Criticals under **Not fixed / deferred** in Step 4 with the reason that the workflow caps at one extra round.

### Step 3b: Orchestrator coverage gate (do not skip)

Before Step 4, the orchestrator **must independently verify** coverage — do not trust the subagent's summary alone. This is the exception to "orchestrator does not implement": verification only.

1. Identify every `.go` file changed on the branch (excluding generated/`_test.go` if you only care about prod code — but still cover new branches in non-test files).
2. Run `go test` with `-coverprofile` / `-coverpkg` for the affected packages (same pattern as `AGENTS.md`).
3. For each modified **non-test** function in those files, require `go tool cover -func` = **100%**. If a review listed specific uncovered lines, confirm those blocks are now hit.
4. If anything fails the gate, `resume` the dev subagent again with the exact cover gaps; do not write the final summary until the gate passes (or the user explicitly accepts a documented unreachable line).

## Step 4: Final summary (orchestrator writes this)

After fixes land (including Step 3a when it ran) **and Step 3b passes**, the orchestrator presents a concise summary to the user covering:

- **Fixed:** Critical issues resolved and Suggestions applied (note round 1 vs round 2 when Step 3a ran).
- **Not fixed / deferred:** review findings intentionally skipped or still open after the optional second round, each with a reason.
- **Plan coverage:** which planned deliverables are done, and which were not done (with why).
- **Coverage gate:** note that modified functions were verified at 100% (or list accepted exceptions).
- **Manual testing needed:** call out anything that should be verified by hand (e.g. flows not covered by automated tests, DB migration effects, event dispatch payloads, permission/propagation edge cases, external integrations like SQS). Be specific about what to run/check.

Keep it scannable. This summary is the deliverable of the skill.

## Notes

- Always reuse the single dev subagent across Steps 1, 3, and 3a (second fix) so it keeps the context of what it built.
- Review subagents are independent and read-only — they must not modify code. Use a fresh review subagent for Step 3a (do not resume the Step 2 reviewer).
- Step 3a runs at most once, and only when Step 2 reported ≥1 Critical. Suggestions alone do not trigger a second round.
- If the orchestrator drifts off `auto` or no plan is approved, return to Preconditions before continuing.
- **Coverage failure mode to avoid:** testing a new helper in isolation while leaving the production caller's new `if err != nil { return err }` uncovered. Codecov patches flag those caller lines; the gate in Step 3b exists to catch them.
