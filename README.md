# Archon-Ready Repository Template

This is a GitHub repository template configured for the **Archon** RAG system. It includes Kiro steering that automatically maintains accurate, RAG-ready documentation.

## What's Included

### Documentation Structure
- `.kiro/docs/` - Complete documentation skeleton optimized for RAG retrieval
- `CLAUDE.md` - Documentation contract and standards
- All required documentation files (overview, architecture, operations, api, data-models, faq)

### Kiro Integration
- `.kiro/steering/archon-docs.md` - Always-active steering that enforces documentation standards
- Automatic documentation maintenance across all Kiro tasks
- RAG-friendly structure and provenance tracking

## Getting Started

1. **Use this template** to create a new repository
2. **Customize** the documentation files under `.kiro/docs/` for your project
3. **Work with Kiro** - the steering will automatically maintain documentation standards

## Documentation Standards

All documentation follows these principles:

- **Grounded in code** - Every statement references actual code or infrastructure
- **RAG-friendly** - Structured for optimal retrieval (400-800 token sections)
- **Provenance** - Clear "Source" references to relevant files
- **No hallucinations** - Only documented, verifiable behavior
- **Always current** - Updated alongside code changes

## Archon Integration

This repository is configured to be ingested by Archon, which reads all Markdown files under `.kiro/docs/` to build mental models for sourcing code and architectural information.

See `CLAUDE.md` for the complete documentation contract.

## License

[Add your license]
