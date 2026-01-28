---
inclusion: always
---

# Archon RAG Integration via MCP

You have access to the **Archon RAG system** via MCP (Model Context Protocol) tools. These tools provide grounded retrieval from internal Archon/Aphex platform documentation.

---

## Available MCP Tools

### archon.search
Search across internal repos and documentation. Returns ranked chunks with provenance.

**When to use:**
- User asks "how does our platform do X"
- User asks "what is the current pattern"
- User asks "where is Y defined"
- User asks "what do we already have"
- User asks "how does this integrate"
- Generating YAML manifests/CRDs/controllers/pipelines that must match existing conventions
- Any question about Aphex/Archon repos or architecture

**Parameters:**
- `query` (string, required): Search query text
- `top_k` (number, optional): Number of results to return (default: 5)
- `repo_filter` (string, optional): Filter to specific repository

**Returns:**
```json
{
  "chunks": [
    {
      "doc_id": "uuid",
      "chunk_id": "uuid",
      "source_path": "repo/path/to/file.md",
      "repo": "ArchonAgent",
      "content": "retrieved text...",
      "score": 0.85
    }
  ]
}
```

### archon.get_document
Fetch full document text when you need complete context for edits or spec generation.

**Parameters:**
- `doc_id` (string, required): Document ID from search results

**Returns:**
```json
{
  "doc_id": "uuid",
  "repo": "ArchonAgent",
  "file_path": ".kiro/docs/architecture.md",
  "full_content": "complete markdown...",
  "metadata": {}
}
```

### archon.list_repos
List available repositories in the Archon knowledge base.

**Returns:**
```json
{
  "repos": [
    {
      "name": "ArchonAgent",
      "description": "LLM orchestration and model serving",
      "last_updated": "2026-01-26T14:00:00Z"
    }
  ]
}
```

---

## Usage Policy

### MUST Call MCP Tools

**Before making implementation decisions, writing specs, or answering questions that could depend on internal architecture, existing repos, or current conventions, you MUST consult the Archon MCP tools.**

**Workflow:**
1. Call `archon.search` first with a precise query
2. If results are ambiguous, refine and call `archon.search` again
3. If you need full context, call `archon.get_document` on relevant `doc_id`s

### MUST Ground Claims

**You MUST ground key claims in retrieved passages:**
- Include `doc_id`/`chunk_id` references in your reasoning
- Include citations or "From <repo>/<file>" in final output when appropriate
- Quote small snippets only when needed; otherwise paraphrase with citations

### DO NOT Invent

**Do NOT invent repo structure, API shapes, CRD schemas, or deployment details if you haven't retrieved supporting evidence.**

If tools return "no results" or low confidence:
- Explicitly say what you could not verify
- Propose the smallest safe assumption set
- Ask user for clarification

---

## Trigger Rules

Call Archon MCP tools when:

- User asks about "how our platform does X"
- User asks about "current patterns" or "existing conventions"
- User asks "where is Y defined" or "what do we already have"
- User asks about integration between Archon/Aphex components
- Generating YAML manifests, CRDs, controllers, or pipelines
- User asks for "latest" behavior, versions, or active interfaces
- User references specific repos (ArchonAgent, AphexPlatformInfrastructure, etc.)

---

## Output Rules

### Include Grounding Section

When using retrieved information, include a "Grounding" section:

```markdown
**Grounding:**
- ArchonAgent/.kiro/docs/architecture.md (doc_id: abc123)
- AphexPlatformInfrastructure/.kiro/docs/operations.md (doc_id: def456)
```

### Prefer Paraphrasing with Citations

- Quote small snippets only when exact wording matters
- Otherwise paraphrase and cite source: "According to ArchonAgent architecture docs..."
- Always trace claims back to retrieved passages

### Be Explicit About Gaps

If retrieval doesn't cover the question:
- "I couldn't find documentation about X in the Archon knowledge base"
- "The retrieved docs don't specify Y; I recommend checking Z or asking the team"

---

## Integration with Documentation Standards

This steering works alongside `archon-docs.md`:

- **archon-docs.md**: Ensures `.kiro/docs/` stays accurate and RAG-friendly
- **archon-rag.md** (this file): Ensures you USE those docs via MCP tools

Together, they create a feedback loop:
1. You maintain accurate docs (archon-docs.md)
2. Archon ingests those docs into RAG system
3. You retrieve and ground decisions in those docs (archon-rag.md)

---

## Example Usage

**User asks:** "How do I deploy a new service to the Archon platform?"

**Your workflow:**
1. Call `archon.search("deploy service Archon platform", top_k=5)`
2. Review results from AphexPlatformInfrastructure and ArchonAgent
3. If needed, call `archon.get_document(doc_id)` for full deployment guide
4. Synthesize answer with citations: "According to AphexPlatformInfrastructure operations docs..."

**User asks:** "What's the current vLLM version in ArchonAgent?"

**Your workflow:**
1. Call `archon.search("vLLM version ArchonAgent", repo_filter="ArchonAgent")`
2. Extract version from retrieved chunk
3. Answer: "ArchonAgent uses vLLM v0.14.1-cu130 (from ArchonAgent/.kiro/docs/architecture.md)"

---

You exist to make decisions grounded in the actual state of the Archon system, not assumptions.
