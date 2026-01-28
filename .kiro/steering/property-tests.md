---
inclusion: always
---

# Property-Based Testing Placement Standard

You are working in a repository that follows **strict test-structure conventions** to ensure consistency, discoverability, and tooling compatibility.

Your responsibility:

> Ensure all **property-based tests** are located exclusively under `kiro/{current-spec}/tests/`.

This steering applies to **all tasks** that generate, modify, refactor, or move tests.

---

## Canonical Location

All **property-based tests** MUST live in:

.kiro/{current-spect}/tests/


This directory is the **single canonical location** for invariant-, generator-, or fuzz-style tests.

---

## What Qualifies as a Property-Based Test

A test is considered **property-based** if it meets one or more of the following criteria:

- Validates **invariants** or **behavioral properties** across many inputs
- Uses generated, randomized, or fuzzed inputs
- Asserts correctness over a **class of inputs**, not fixed examples
- Uses QuickCheck-style, fuzzing, or generator-based frameworks
- Is designed to find edge cases rather than validate a single scenario

If a test *could reasonably be described as asserting a property*, it belongs in `kiro/{current-spect}/tests/`.

---

## Non-Goals

This steering **does not apply** to:

- Unit tests
- Snapshot tests
- Integration tests
- End-to-end tests

Those tests may continue to live in their existing, conventional locations.

---

## Enforcement Rules

When working with tests, you MUST enforce the following:

1. **Creation**
   - Any newly created property-based test MUST be placed in `kiro/{current-spect}/tests/`
   - Property-based tests MUST NOT be placed alongside implementation code

2. **Modification**
   - When editing an existing property-based test, verify it is in `kiro/{current-spect}/tests/`
   - If it is not, move it and update imports or references accordingly

3. **Refactoring**
   - If a test evolves from example-based to property-based, it MUST be moved into `kiro/{current-spect}/tests/`
   - If a test is ambiguous, default to treating it as property-based and place it in `kiro/{current-spect}/tests/`

4. **Discovery**
   - If you encounter a property-based test outside `kiro/{current-spect}/tests/`, you SHOULD flag it and recommend relocation

---

## Rationale

Centralizing property-based tests:

- Makes invariant testing easy to locate and reason about
- Prevents mixing generative tests with example-based unit tests
- Enables targeted CI, fuzzing, and stress-test workflows
- Improves long-term maintainability and test hygiene

This structure is intentional and must be preserved.

---

## Guardrails

- Do not duplicate property-based tests across multiple directories
- Do not co-locate property-based tests with implementation code
- Do not create alternative directories for generative or fuzz tests

If a test checks properties, **it goes in `kiro/{current-spect}/tests/` — no exceptions**.
