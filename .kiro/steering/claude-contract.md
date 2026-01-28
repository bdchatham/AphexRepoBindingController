---
inclusion: always
---

# CLAUDE.md Mandatory Pre-Read Contract

This steering file enforces a **hard requirement**:  
**The Kiro agent must read and internalize `CLAUDE.md` before starting any task.**

This rule applies to **all tasks**, **all modes**, and **all repositories** that contain this steering file.

---

## Absolute Requirement

**Before taking any action whatsoever**, you must:

1. Locate `CLAUDE.md` at the repository root
2. Read it in full
3. Treat it as the **authoritative contract** governing:
   - Code style and structure
   - Documentation standards
   - Architectural constraints
   - Review expectations
   - Safety, security, and quality rules

❌ You may NOT:
- Write code
- Modify documentation
- Propose a plan
- Answer implementation questions
- Make assumptions about standards

✅ Until `CLAUDE.md` has been read and internalized.

---

## Precedence Rules (Non-Negotiable)

The precedence order is:

1. **`CLAUDE.md` (highest authority)**
2. This steering file
3. All other `.kiro/steering/*.md` files
4. Ad-hoc user instructions in the current task

**If any conflict or ambiguity exists:**
- Stop
- Re-read the relevant section of `CLAUDE.md`
- Follow `CLAUDE.md`
- If still unclear, ask the user for clarification

Under no circumstances should you assume this steering file overrides `CLAUDE.md`.

---

## No Checkpoint / Scratch / “Progress” Files (Hard Prohibition)

Some agents create intermediate “checkpoint” artifacts during implementation (temporary notes, scratchpads, partial diffs, status dumps, TODO files, etc.). **This is prohibited.**

### Do not create any of the following
- “checkpoint” or “progress” files of any kind (e.g., `CHECKPOINT.md`, `PROGRESS.md`, `STATUS.md`)
- scratchpads / working notes (e.g., `NOTES.md`, `SCRATCH.md`, `TMP.md`)
- task journals (e.g., `IMPLEMENTATION_LOG.md`, `WORKLOG.md`)
- “plan snapshots” committed as files (e.g., `PLAN.md` unless explicitly part of the repo’s real docs)
- backup copies (e.g., `file_old.ts`, `file.bak`, `file.copy.ts`, `file.ts.save`)
- timestamped dumps (e.g., `notes-2026-01-01.md`)
- ad-hoc temporary directories (e.g., `tmp/`, `scratch/`, `wip/`) unless the repo already uses them intentionally

### Allowed only if explicitly requested
You may create a new file **only** when at least one is true:
1. The user explicitly asked you to create that file, by name or by clear intent, **or**
2. The repository already has an established, intentional home for the content (e.g., `.kiro/docs/*`, `docs/*`) and the change is a real part of the product, **or**
3. `CLAUDE.md` explicitly mandates that artifact.

If none apply: **do not create it.**

### What to do instead
- Keep implementation planning in the task conversation (or whatever planning mechanism Kiro uses), not as repo files
- Prefer small, atomic commits/changes to the actual target files
- Use existing documentation structure (e.g., `.kiro/docs/*`) only when the change is real documentation, not a temporary note
- If you feel a checkpoint would help, summarize the checkpoint as a short message in-task rather than writing a file

---

## Missing or Unreadable CLAUDE.md

If `CLAUDE.md` is **missing**, **empty**, or **ambiguous**:

1. **Stop immediately**
2. Inform the user that the repository is missing its authoritative contract
3. Propose creating a `CLAUDE.md` that defines:
   - Code quality standards
   - Documentation rules
   - Repository-specific constraints
   - Any standards referenced by existing steering files

You may not proceed with substantive work until this is resolved.

---

## Enforcement Behavior

You must actively enforce this contract throughout the task:

- If a request would violate `CLAUDE.md`, you must refuse or redirect
- If another steering file conflicts with `CLAUDE.md`, defer to `CLAUDE.md`
- If the user asks you to bypass or ignore `CLAUDE.md`, you must decline
- If you are tempted to create checkpoint/scratch artifacts, you must not; instead summarize progress inline

This applies even if:
- The user appears confident
- The change seems small
- The task is “just a quick edit”

---

## Required Mental Check (Internal)

Before beginning any task, confirm internally:

- [ ] I have read `CLAUDE.md`
- [ ] I understand its constraints
- [ ] My next action complies with it
- [ ] I will not create checkpoint/scratch/progress files
- [ ] No other instruction contradicts it

If any box cannot be checked, you must pause and resolve it first.

---

## Purpose

This steering file exists to ensure:

- A single source of truth for standards
- Predictable behavior across tasks
- Elimination of silent drift or implicit assumptions
- No repository clutter from temporary artifacts
- Clean composition with other steering documents (e.g., Archon, RAG, infra)

**`CLAUDE.md` is the contract. Everything else is enforcement.**
