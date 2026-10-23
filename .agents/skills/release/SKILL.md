---
name: release
description: Cut a new AlgoreaBackend release - changelog entry from merged PRs, version bump, commit, tag and push. Use when the user wants to release, bump the version, tag, or write a changelog entry. Supports auto mode (no human interaction).
---

# Release

Steps run in order. Two modes:

- **Interactive** (default): pause where a step says so.
- **Auto mode**: the caller asked for it (`/release auto`, "release without asking", ...). Never pause: accept the proposed version and entries, run the release command, and report everything in the final summary.

A **stop** below applies in both modes: report the reason and end the run, with nothing committed or pushed.

## 1. Verify state

```bash
git branch --show-current
git status --porcelain
git fetch origin
git rev-list --left-right --count master...origin/master
```

Stop unless: on `master`, empty porcelain output, and the count is `0	0` (local `master` neither ahead of nor behind `origin/master`).

## 2. Read current version

The latest section heading at the top of `CHANGELOG.md`:

```
## [vX.Y.Z](https://github.com/France-ioi/AlgoreaBackend/compare/vA.B.C...vX.Y.Z) - YYYY-MM-DD
```

Stop if this version differs from `git describe --tags --abbrev=0` (the changelog is out of sync with the tags).

## 3. Collect changes

```bash
git log --oneline --merges $(git describe --tags --abbrev=0)..HEAD
git log --oneline $(git describe --tags --abbrev=0)..HEAD
```

Stop if there is no commit since the last tag. For each merged PR, `gh pr view <number> --json title,body` gives the context the merge commit lacks.

Write one entry per meaningful change:

- the **API-facing impact** (endpoints, fields, events, behaviour), not the implementation;
- concise plain English, starting with a lowercase verb or noun phrase (`- add ...`, `- fix ...`, `- new service: ...`, `- new event dispatched: ...`);
- no PR or commit links.

Skip internal-only changes (refactors, CI, tests) unless they visibly affect reliability or performance. If only internal changes remain, the release is a patch with the single entry `- internal maintenance`.

## 4. Decide the version bump

| Changes | Bump |
|---|---|
| Only fixes, corrections, internal improvements, no API change | patch (x.y.**Z+1**) |
| Any API change: new endpoints, fields, events, options, anything an API consumer may notice | minor (x.**Y+1**.0) |
| Clearly breaking (removed endpoints, renamed fields, changed behaviour) | minor too, but **flag it** so the user can choose a major bump |

When in doubt between patch and minor, pick minor. In auto mode, a flagged breaking change stays minor and the flag goes in the final summary.

## 5. Proposal

Show the proposed version (`vX.Y.Z`, was `vA.B.C`) and the CHANGELOG section of step 6, dated today.

**Interactive: stop here** until the user approves, edits the entries, or changes the version. Auto mode: continue.

## 6. Update CHANGELOG.md

Insert the new section right after the `# Changelog` header and its subtitle line, before the previous latest section. Exact layout (blank line after the heading and after the bullets):

```
## [vX.Y.Z](https://github.com/France-ioi/AlgoreaBackend/compare/vA.B.C...vX.Y.Z) - YYYY-MM-DD

- entry one
- entry two

```

Leave the rest of the file untouched.

## 7. Commit, tag and push

The release command, with the real version:

```bash
VERSION=X.Y.Z; git add CHANGELOG.md && git commit -m "Release $VERSION" && git tag -a -m "Release v$VERSION" v$VERSION && git push origin v$VERSION && git push origin master
```

- **Interactive**: show it, explain it commits the changelog, creates the annotated tag and pushes both, and tell the user they can either run it or reply **"do it"**. On "do it", run exactly this command, nothing else.
- **Auto mode**: run it.

If any part fails, stop and report the output; leave recovery (retries, tag deletion, force pushes) to the user.

## 8. After the push

The tag triggers CI, which builds the binaries, publishes the GitHub release with them, generates and deploys the API doc, and notifies ops. Nothing else to do: end with a summary (version, entries, anything flagged).
