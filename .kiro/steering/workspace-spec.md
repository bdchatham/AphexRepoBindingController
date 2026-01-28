---
inclusion: manual
priority: highest
---

# Workspace-Driven Spec & Implementation Steering

## Core Mental Model

You (Kiro) operate on a **workspace**, not a single package.

A **workspace** is a directory that contains *all packages required to implement a feature end-to-end*, including:
- Product services
- Platform or shared components
- Infrastructure definitions
- Pipelines and tooling

The workspace is the **unit of design**.  
Packages are **units of implementation**.

---

## Absolute Rules

1. **One workspace → one design**
2. **One workspace → one requirements set**
3. **One workspace → one task list**
4. Packages are not designed independently
5. Implementation does not begin until the spec phase is complete

Tasks may cross package boundaries freely.  
Ownership boundaries are respected at implementation time, not design time.

---

## The Spec-Driven Flow (Mandatory Order)

You must execute work in **two strictly separated phases**:

PHASE 1 — DESIGN (Workspace level)
PHASE 2 — IMPLEMENTATION (Package level)

yaml
Copy code

You may not interleave these phases.

---

## PHASE 1 — DESIGN (Workspace Level)

### Goal

Produce a **single, coherent system design** that is:
- Grounded in reality
- Dependency-driven
- Explicit about boundaries
- Ready to implement without invention

### Artifacts Produced (Exactly These)

All stored under the workspace’s canonical `.kiro/` directory:

1. **Workspace Design Spec**
2. **Workspace Requirements**
3. **Workspace Task List**

No package-local tasks or requirements are created in this phase.

---

### Step 1: Workspace Discovery

You must inventory **everything in the workspace** before designing.

For each package, record:
- Name + path
- Stated purpose (from docs)
- What it owns (runtime, data, infra, pipelines)
- Interfaces it exposes or consumes
- Evidence sources (file paths)

Packages are classified as:
- **Participating** (actively part of the feature)
- **Constraining** (define boundaries/interfaces)
- **Out-of-scope** (explicitly excluded)

All three must be listed in the design spec.

No silent exclusions.

---

### Step 2: Current-State System Understanding

Before proposing anything new, you must understand:
- How data flows today
- How control flows today
- Where coupling already exists
- Where boundaries are currently enforced or violated

This understanding must be grounded in:
- Code
- Infra
- Pipelines
- Existing `.kiro/docs`
- Runtime contracts

Assumptions are not allowed.  
Unknowns must be labeled as TODOs with required evidence.

---

### Step 3: End-to-End Design

Produce **one system design** for the workspace:

- Components and responsibilities
- Data flow
- Control flow
- Failure boundaries
- Ownership boundaries

This design:
- Resolves dependencies explicitly
- Decides where logic belongs
- Clarifies platform vs product responsibilities
- Avoids speculative abstractions

This is the **authoritative design**.

---

### Step 4: Requirements Derivation

From the design, derive **one requirements set** for the workspace.

Requirements:
- Are system-level, not package-local
- Are dependency-driven
- Are testable
- Reference the design sections they originate from

Requirements may later be *implemented* in different packages, but they are **defined once**.

---

### Step 5: Task Planning

From the requirements, generate **one ordered task list**.

Tasks:
- May span multiple packages
- Are ordered to reduce risk early
- Include investigation, contracts, scaffolding, implementation, validation
- Reference the requirements they satisfy

Tasks do **not** belong to individual packages yet.  
They belong to the workspace plan.

---

### Phase 1 Completion Gate (Hard Stop)

You may not proceed unless:

- The design spec exists and is complete
- Requirements are fully derived
- Tasks are fully listed and ordered
- All TODOs are explicit and tracked

Only then may implementation begin.

---

## PHASE 2 — IMPLEMENTATION (Package Level)

### Entry Rule

Before working in **any package**, you must:

1. Read that package’s `CLAUDE.md`
2. Read that package’s `.kiro/steering/*`
3. Read that package’s `.kiro/docs/*`
4. Confirm no conflicts with workspace design

If conflicts exist:
- Stop
- Surface them
- Resolve at the workspace design level

---

### Task Execution Model

Tasks are pulled from the **workspace task list**, not invented per package.

When executing a task:
- Determine which package(s) it touches
- Enter those packages one at a time
- Follow **all local standards**
- Make changes consistent with the workspace design

No package-local redesigns are allowed.

---

### Documentation & Code Standards

While implementing:
- All existing documentation and code standards apply
- The agent must comply with each package’s steering and contracts
- Cross-package consistency must be maintained

If implementation reveals a design flaw:
- Stop
- Propose a workspace-level design update
- Do not patch around it locally

---

## What This Steering Prevents (Intentionally)

- Designing packages in isolation
- Letting platform abstractions drift
- Creating duplicate or conflicting requirements
- Accidental architectural divergence
- “Implementation-led design”

---

## Completion Criteria

The workspace feature is complete only when:

- All workspace tasks are done
- Requirements are satisfied
- Design remains accurate
- Package docs reflect reality
- No unresolved TODOs remain

---

## Guiding Principle

> **Design once.  
> Implement everywhere.  
> Never redesign locally.**

The workspace is the system.  
Packages are how it is built.
