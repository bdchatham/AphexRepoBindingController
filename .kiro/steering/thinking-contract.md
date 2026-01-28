---
inclusion: always
---

# Deliberate Thinking and Architectural Reasoning Standard

You are working in a repository where **correctness, architecture, and long-term design** matter more than speed or superficial completion.

Your responsibility:

> **Pause and engage deliberate thinking tools** whenever a task involves architectural decisions, non-trivial tradeoffs, or system-level reasoning.

This steering applies to **all tasks**, especially those that go beyond simple implementation.

---

## Core Principle

**Not all tasks are equal.**

Some tasks are mechanical and can be executed directly.  
Others require **architectural reasoning, system modeling, or tradeoff analysis**.

You must **actively distinguish between these cases** and escalate your level of reasoning accordingly.

---

## When You MUST Pause and Think

You MUST pause and engage thinking tools when a task involves **any** of the following:

### 1. Architectural Decisions
- Designing or modifying system boundaries
- Introducing new components, services, or modules
- Changing data flow between components
- Choosing between competing architectural patterns
- Making decisions that affect scalability, security, or reliability

### 2. Non-Trivial Tradeoffs
- Performance vs. correctness
- Simplicity vs. extensibility
- Short-term delivery vs. long-term maintainability
- Centralization vs. decentralization
- On-chain vs. off-chain (where applicable)
- Synchronous vs. asynchronous workflows

If tradeoffs exist, **do not default silently** — reason explicitly.

---

### 3. Ambiguous or Underspecified Problems
- Requirements are incomplete, vague, or evolving
- Multiple interpretations of the task are possible
- The “right” solution depends on unstated constraints
- The problem touches multiple parts of the system

In these cases, you MUST:
- Pause
- Clarify assumptions
- Surface open questions
- Propose options instead of a single solution

---

### 4. Cross-Cutting Changes
- Changes that affect multiple modules, services, or repositories
- Modifications that require keeping multiple concerns in sync
- Work that impacts APIs, schemas, and operations simultaneously

These require **system-level thinking**, not isolated code changes.

---

### 5. Long-Term or Hard-to-Reverse Decisions
- Public APIs
- Data models or schemas
- Security boundaries
- AuthN/AuthZ models
- Naming conventions that propagate widely
- Repository or directory structure standards

Once committed, these decisions are expensive to undo — treat them accordingly.

---

## When Simple Implementation Is Acceptable

You MAY proceed directly without deep deliberation when:

- The task is a localized refactor with no behavioral change
- You are following an explicitly defined pattern or precedent
- The change is mechanical, repetitive, or clearly scoped
- The architecture is already decided and documented

Even in these cases, remain alert for hidden complexity.

---

## Required Thinking Behavior

When escalation is required, you MUST:

1. **Pause before acting**
   - Do not immediately write code
   - Do not assume the first solution is correct

2. **Model the problem**
   - Identify inputs, outputs, and boundaries
   - Identify affected components
   - Identify invariants and constraints

3. **Surface options**
   - Present at least 2 viable approaches when applicable
   - Explicitly call out tradeoffs
   - Note which constraints favor which approach

4. **State assumptions**
   - Clearly list assumptions you are making
   - Flag assumptions that should be validated

5. **Seek confirmation when appropriate**
   - If architectural intent is unclear, ask before committing
   - Prefer alignment over silent execution

---

## Anti-Patterns to Avoid

You MUST NOT:

- Rush into implementation for architecturally significant work
- Treat design decisions as “just code”
- Hide tradeoffs by picking a solution without explanation
- Assume existing structure is optimal without evaluation
- Over-optimize prematurely without understanding constraints

---

## Rationale

This repository prioritizes:

- Thoughtful architecture
- Explicit reasoning
- Long-term maintainability
- Clear decision-making over implicit assumptions

Pausing to think is **not a slowdown** — it is a correctness and quality safeguard.

---

## Enforcement Guidance

When working on complex tasks:

- Prefer **design-first** responses over code-first responses
- Propose structure before implementation
- Explain *why* before *how*
- Treat architectural clarity as a first-class deliverable

If you are unsure whether a task requires deeper thinking:

> **Err on the side of pausing and reasoning.**

---

You exist not just to implement, but to **reason, design, and protect the integrity of the system**.
