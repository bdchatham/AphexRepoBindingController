# Archon Documentation Contract

This repository participates in the **Archon** RAG system, which ingests documentation to build mental models for sourcing code and architectural information.

## Documentation Location

The canonical, ingestible documentation location is **`.kiro/docs/`**.

Archon ingests all Markdown files under `.kiro/docs/` from this public GitHub repository.

## Documentation Stability Principle

The documentation structure is **intentionally stable** to ensure reliable RAG retrieval across repository updates.

**Core Principle: Exactly 6 Documentation Files**

This repository maintains exactly **6 core documentation files** under `.kiro/docs/`:
1. `overview.md`
2. `architecture.md`
3. `operations.md`
4. `api.md`
5. `data-models.md`
6. `faq.md`

**Prohibition on New Files**

Do **NOT** create additional documentation files for new features, components, or capabilities. The 6-file structure is fixed and must remain stable.

**Where to Document New Features**

When adding new features, components, or capabilities:
- Add new sections within the appropriate existing file
- Use descriptive headings to make content discoverable
- Update multiple files if the feature spans multiple concerns

Examples:
- New component → Add section to `architecture.md`, update `operations.md` for deployment
- New API endpoint → Add section to `api.md`, potentially update `architecture.md`
- New data schema → Add section to `data-models.md`, potentially update `architecture.md`

**Rationale**

Stable file structure enables:
- Consistent RAG retrieval patterns across repository versions
- Predictable documentation locations for engineers and operators
- Reliable mental models for automated agents
- Reduced cognitive overhead when navigating documentation

## Required Documentation Files

Maintain the following files under `.kiro/docs/`:

### `.kiro/docs/overview.md`
High-level purpose, context, and scope of this repository. Explains what problem this repo solves and how it fits into the broader system.

### `.kiro/docs/architecture.md`
System design, components, and their relationships. Includes diagrams (as Markdown), technology choices, and architectural patterns.

### `.kiro/docs/operations.md`
Deployment procedures, monitoring, alerting, runbooks, and operational concerns. How to deploy, troubleshoot, and maintain this system.

### `.kiro/docs/api.md`
API contracts, interfaces, endpoints, and integration patterns. Documents how other systems interact with this one.

### `.kiro/docs/data-models.md`
Data structures, schemas, database models, and data flow. Describes what data this system manages and how.

### `.kiro/docs/faq.md`
Common questions, gotchas, and quick answers. Helps new contributors and operators get up to speed.

## Documentation Maintenance Philosophy

Documentation should evolve incrementally alongside the codebase, not through large rewrites.

**Update Existing Sections, Don't Create New Files**

When the system changes:
- Identify which existing sections are affected
- Update those sections with new information
- Remove or correct stale content
- Do NOT create new documentation files

**Incremental Updates Over Large Rewrites**

Prefer small, focused updates:
- Update one section at a time as code changes
- Make documentation changes in the same commit as code changes
- Keep documentation synchronized with reality through continuous small updates
- Avoid "documentation sprints" that rewrite entire files

**Remove Stale Content**

When behavior changes:
- Delete or update the old documentation
- Do NOT add new content alongside stale content
- Do NOT mark sections as "deprecated" and leave them
- Maintain a single, current source of truth

**Why Incremental Maintenance Matters**

Large documentation rewrites:
- Often introduce inconsistencies across files
- May hallucinate or guess at behavior
- Create maintenance burden and drift
- Disrupt RAG retrieval patterns

Incremental updates:
- Stay grounded in actual code changes
- Maintain consistency across files
- Keep documentation continuously accurate
- Preserve stable retrieval patterns

**Practical Guidelines**

When making code changes:
1. Identify affected documentation sections (may span multiple files)
2. Update those specific sections
3. Verify cross-file consistency
4. Commit documentation with code changes

When reviewing documentation:
1. Verify against actual code behavior
2. Check for stale or contradictory content
3. Refactor oversized sections (see Documentation Standards)
4. Ensure provenance is current

## Stakeholders and Usage

Documentation in `.kiro/docs/` serves three distinct stakeholder types, each with different needs and usage patterns.

### Engineers

**How They Use Documentation:**
- Understand system architecture and design decisions
- Learn how components interact and integrate
- Find API contracts and data models
- Debug issues and understand behavior
- Onboard to the codebase

**Optimization for Engineers:**
- Include specific file references and code examples
- Explain "why" behind architectural decisions
- Provide clear API contracts and interfaces
- Document common gotchas and troubleshooting steps
- Link related concepts across files

### Operators

**How They Use Documentation:**
- Deploy and configure the system
- Monitor system health and performance
- Respond to incidents and alerts
- Execute runbooks and operational procedures
- Understand system dependencies

**Optimization for Operators:**
- Provide step-by-step deployment procedures
- Document monitoring and alerting setup
- Include clear runbooks for common issues
- Explain operational dependencies and requirements
- Specify configuration options and their effects

### RAG Agents

**How They Use Documentation:**
- Retrieve relevant information based on semantic queries
- Build mental models of system architecture
- Answer questions about system behavior
- Identify relevant code and infrastructure
- Synthesize information across multiple files

**Optimization for RAG Agents:**
- Use descriptive, specific headings as retrieval keys
- Keep sections focused and appropriately sized (400-800 tokens)
- Maintain consistent terminology across all files
- Group related information in single sections
- Provide clear provenance linking docs to code
- Avoid duplication that could confuse retrieval

**Cross-Stakeholder Considerations**

When updating documentation:
- Consider all three stakeholder types
- Ensure content serves multiple audiences where possible
- Balance technical depth with operational clarity
- Optimize structure for both human readers and RAG retrieval
- Verify that changes maintain value for all stakeholders

## Documentation Standards

### Grounding in Code
All documentation must be grounded in actual code and infrastructure:
- Reference specific files (e.g., `src/handler.py`, `infra/stack.ts`)
- Include "Source" sections pointing to relevant code
- Update docs when code changes

### RAG-Friendly Structure
Documentation should be optimized for retrieval:
- Use clear headings and subheadings
- Keep sections focused (400–800 tokens each)
- Use direct, factual language
- Prefer lists and step-by-step instructions
- Avoid long, monolithic sections

### Retrieval Optimization Strategy

Documentation structure directly impacts RAG retrieval quality. Follow these principles to ensure optimal retrieval:

**Headings as Retrieval Keys**

Headings are the primary mechanism for RAG systems to locate relevant information. They function as semantic keys that match user queries.

- Headings should be descriptive and specific
- Headings should contain key terms that users would search for
- Headings should clearly indicate the content that follows

**Good Heading Examples:**
- "Lambda Function Deployment Process"
- "DynamoDB Table Schema for User Profiles"
- "Authentication Flow for API Endpoints"
- "Monitoring CloudWatch Alarms for Service Health"

**Bad Heading Examples (Too Generic):**
- "Details" (what details?)
- "Information" (what information?)
- "Overview" (overview of what?)
- "Configuration" (configuration of what?)

**Require Descriptive, Specific Headings**

Every heading must:
- Clearly identify the topic being discussed
- Include specific component, feature, or concept names
- Be unique within the file (avoid duplicate headings)
- Be meaningful when read in isolation

**Mandate Consistent Terminology**

Use the same terms across all documentation files:
- Choose one term for each concept and use it consistently
- Example: Use "Lambda function" everywhere, not "Lambda", "function", "handler" interchangeably
- Example: Use "DynamoDB table" everywhere, not "table", "database", "data store" interchangeably
- Document preferred terminology in the glossary section of relevant files

**Terminology Consistency Checklist:**
- Component names: Use exact names from code (e.g., `DocumentMonitor`, not "document monitor" or "doc monitor")
- Technology names: Use official names (e.g., "Amazon DynamoDB", not "Dynamo" or "DDB" in prose)
- Concepts: Define once, use consistently (e.g., "deployment pipeline" vs "CI/CD pipeline" vs "build pipeline")

**Group Related Information**

Keep related information together in single sections:
- Don't scatter information about a component across multiple sections
- If a component has multiple aspects (architecture, deployment, monitoring), consider whether they belong in one file or should be split across files by concern
- Use cross-references when information must span files

**Why This Matters for RAG:**
- RAG systems retrieve chunks of text based on semantic similarity
- Descriptive headings improve matching between queries and content
- Consistent terminology prevents retrieval failures due to synonym mismatches
- Grouped information ensures complete context in retrieved chunks
- Well-structured sections improve retrieval precision and recall

### Provenance
Each significant section should include a "Source" subsection:

```markdown
**Source**
- `src/document_monitor.py`
- `infra/archon-cron-stack.ts`
```

### No Hallucinations
Only document behavior that can be verified from:
- This repository's code
- This repository's infrastructure
- This repository's existing specs

If something is uncertain, mark it as a TODO rather than guessing.

### Avoid Duplication
Link to existing documentation rather than repeating it. Maintain a single source of truth for each concept.

### Documentation Refactoring

Documentation sections should remain appropriately sized for optimal RAG retrieval. Oversized sections reduce retrieval precision and make content harder to navigate.

**Maximum Section Size**

Each section should target **400-800 tokens** and must not exceed **~1000 tokens**.

Token estimation:
- 1 token ≈ 4 characters
- 1000 tokens ≈ 4000 characters ≈ 600-800 words
- Use character count as a proxy: sections over 4000 characters need refactoring

**Refactoring Triggers**

Refactor a section when:
- Section exceeds ~1000 tokens (4000 characters)
- Section covers multiple distinct concepts
- Section has grown through incremental additions
- Heading no longer accurately describes all content
- Section is difficult to navigate or scan

**Refactoring Patterns**

When a section becomes too large, apply one of these patterns:

**Pattern 1: Split into Subsections**
- Break the section into focused subsections with descriptive headings
- Each subsection should cover one specific aspect
- Maintain the parent heading as an organizing structure
- Example: "Lambda Functions" → "Lambda Function Architecture", "Lambda Function Deployment", "Lambda Function Monitoring"

**Pattern 2: Move to Different File**
- If content spans multiple concerns, distribute across appropriate files
- Example: Component description stays in `architecture.md`, deployment moves to `operations.md`
- Add cross-references between files
- Ensure each file maintains complete, useful information

**Pattern 3: Extract Common Patterns**
- If multiple sections repeat similar information, extract to a shared section
- Reference the shared section from specific sections
- Example: Extract "Common Configuration Patterns" referenced by multiple component sections

**Pattern 4: Create Subsections with Summaries**
- Add a brief summary at the parent level
- Move detailed content into subsections
- Allows readers to scan summaries and dive into details as needed

**Refactoring Process**

1. Identify the oversized section
2. Analyze what concepts it covers
3. Choose appropriate refactoring pattern
4. Create new structure with descriptive headings
5. Distribute content to new sections
6. Add cross-references if content spans files
7. Verify each new section is appropriately sized
8. Update any links to the refactored section

**Refactoring Guidelines**

- Preserve all information (don't delete content during refactoring)
- Maintain or improve heading descriptiveness
- Keep related information together
- Ensure each section can stand alone for retrieval
- Update provenance references if content moves
- Verify terminology consistency after refactoring

**When NOT to Refactor**

Don't refactor if:
- Section is under 1000 tokens and covers a single, cohesive concept
- Splitting would create sections that lack sufficient context
- Content is already well-organized with clear subsections
- Refactoring would duplicate information across files

## Quality Validation Checklist

Before committing documentation changes, validate against these criteria to ensure documentation quality and RAG effectiveness.

### Grounding Validation

**All statements must have code references:**
- [ ] Every significant claim about system behavior references specific files
- [ ] "Source" subsections exist for all major sections
- [ ] File references are current and accurate
- [ ] No guesses or assumptions are documented as facts
- [ ] TODOs are marked for uncertain or unverified content

**How to validate:**
- Review each section and identify claims about behavior
- Verify each claim against actual code or infrastructure
- Add or update "Source" subsections with specific file paths
- Remove or mark as TODO any unverifiable statements

### Structure Validation

**Sections must be appropriately sized and organized:**
- [ ] Each section is 400-800 tokens (target) or under 1000 tokens (maximum)
- [ ] Headings are descriptive and specific (not generic like "Details")
- [ ] Related information is grouped in single sections
- [ ] Sections are focused on single concepts or components
- [ ] Subsections are used to organize complex topics

**How to validate:**
- Check character count for each section (4000 characters ≈ 1000 tokens)
- Review headings for specificity and descriptiveness
- Identify any scattered information that should be grouped
- Verify each section has a clear, single focus
- Refactor oversized sections using patterns above

### Consistency Validation

**Terminology must be consistent across all files:**
- [ ] Component names match code exactly (e.g., `DocumentMonitor`)
- [ ] Technology names use official terminology consistently
- [ ] Concepts use the same terms across all files
- [ ] Acronyms are defined on first use in each file
- [ ] Cross-references use consistent naming

**How to validate:**
- List key terms used in the updated sections
- Search for those terms across all `.kiro/docs/*.md` files
- Verify consistent usage (same term, same meaning)
- Update any inconsistent terminology
- Document preferred terms in glossaries where appropriate

### Completeness Validation

**All affected files must be updated:**
- [ ] All files impacted by the change are identified
- [ ] Each impacted file is updated with relevant information
- [ ] Cross-references between files are added or updated
- [ ] No orphaned or contradictory information remains
- [ ] Related sections across files tell a consistent story

**How to validate:**
- Use the decision tree for common update patterns (see `.kiro/steering/archon-docs.md`)
- Check each of the 6 core files for related content
- Verify cross-file consistency for the changed component/feature
- Update all related sections, not just the primary file
- Remove or update any stale references in other files

### Stakeholder Validation

**Content must serve all stakeholder types:**
- [ ] Engineers can understand architecture and integration
- [ ] Operators can deploy and troubleshoot
- [ ] RAG agents can retrieve relevant information
- [ ] Technical depth is appropriate for the audience
- [ ] Operational procedures are clear and actionable

**How to validate:**
- Review content from each stakeholder perspective
- Verify engineers have sufficient technical detail
- Verify operators have clear procedures and runbooks
- Verify RAG agents have descriptive headings and consistent terminology
- Add missing information for underserved stakeholders

### Retrieval Validation

**Structure must optimize RAG retrieval:**
- [ ] Headings contain key search terms
- [ ] Headings are unique within each file
- [ ] Sections are self-contained with sufficient context
- [ ] Terminology matches likely user queries
- [ ] Related information is co-located

**How to validate:**
- Read each heading in isolation - is it clear what the section contains?
- Verify headings include specific component/feature names
- Check that sections can be understood without reading entire file
- Verify terminology matches what users would search for
- Ensure related concepts are grouped together

## Code Quality Standards

Code quality directly impacts maintainability, readability, and the ability of engineers to understand and modify the system. These standards emphasize clean, self-documenting code that minimizes the need for explanatory comments.

### Naming Conventions

**Use expressive names that convey intent:**
- Variable names should clearly indicate what they contain
- Function names should clearly indicate what they do
- Class names should clearly indicate what they represent
- Avoid abbreviations unless universally understood
- Prefer longer, descriptive names over short, cryptic ones

**Good Examples:**
```python
user_authentication_token = generate_token(user_id)
def calculate_monthly_revenue(transactions):
def validate_email_format(email_address):
```

**Bad Examples:**
```python
uat = gen_tok(uid)  # cryptic abbreviations
def calc(t):  # unclear what is being calculated
def validate(e):  # unclear what is being validated
```

**Why This Matters:**
- Expressive names make code self-documenting
- Reduces need for explanatory comments
- Improves code comprehension for new contributors
- Makes code searchable and navigable

### Method Sizing

**Keep methods focused and appropriately sized:**
- Target: 10-30 lines per method
- Preference: Under 20 lines per method
- Maximum: Avoid exceeding 30 lines without strong justification

**When methods grow too large:**
- Extract helper methods for distinct sub-tasks
- Break complex logic into smaller, named steps
- Each helper method should have a clear, single purpose
- Use descriptive names for helper methods

**Example Refactoring:**

Before (oversized method):
```python
def process_order(order):
    # 50+ lines of validation, calculation, database updates, notifications
    ...
```

After (refactored with helpers):
```python
def process_order(order):
    validate_order_data(order)
    total = calculate_order_total(order)
    save_order_to_database(order, total)
    send_confirmation_email(order)
    update_inventory(order)
```

**Why This Matters:**
- Smaller methods are easier to understand and test
- Helper methods with clear names document the process
- Reduces cognitive load when reading code
- Makes debugging and modification safer

### Commenting Guidelines

**Comments should be infrequent and purposeful:**
- Code should be self-explanatory through expressive naming and clear structure
- Only add comments when behavior cannot be inferred from the code itself
- Comments should explain "why", not "what" or "how"

**When Comments Are Appropriate:**
- Explaining non-obvious business rules or domain logic
- Documenting behavior of remote services or external APIs
- Clarifying complex algorithms or mathematical operations
- Describing foundational architectural components
- Warning about subtle bugs or edge cases

**When Comments Are NOT Appropriate:**
- Describing what the code does (code should be self-explanatory)
- Breadcrumb comments tracking changes or history (use git history)
- Commented-out code (delete it, git preserves history)
- Obvious statements that restate the code

**Breadcrumb Comments Are Prohibited:**

Do NOT include comments that track changes, decisions, or history:

```python
# Changed from using Redis to DynamoDB because Redis was too slow
# TODO: This used to return a list but now returns a dict
# Fixed bug where this would fail on empty input
# Refactored from process_data_v1 to process_data_v2
```

**Why breadcrumb comments are harmful:**
- They clutter the code and reduce readability
- They become stale and misleading over time
- Git history provides complete, accurate change tracking
- They don't help understand current behavior

**Use git commit messages for history:**
- Commit messages should explain why changes were made
- Git blame shows who changed what and when
- Git history is the authoritative source for change tracking

**Good Comment Examples:**

```python
# DynamoDB eventually consistent reads may return stale data for up to 1 second
# This is acceptable for our use case as we prioritize read performance
result = table.get_item(ConsistentRead=False)

# Stripe webhook signatures expire after 5 minutes to prevent replay attacks
# We must validate the signature before processing the webhook payload
validate_stripe_signature(payload, signature, timestamp)

# Binary search requires sorted input - we sort here rather than at insertion
# because reads are 100x more frequent than writes in our access pattern
data.sort()
result = binary_search(data, target)
```

**Bad Comment Examples:**

```python
# Get the user from the database
user = db.get_user(user_id)  # What the code already says

# Loop through all items
for item in items:  # Obvious from the code

# This is the main function
def main():  # Obvious from the name
```

### Clean Code Principles

**Follow Clean Code principles pragmatically:**
- Prioritize readability and maintainability
- Favor simplicity over cleverness
- Write code for humans first, computers second
- Refactor when code becomes difficult to understand
- Balance idealism with practical constraints

**Pragmatism Over Dogmatism:**
- Clean Code principles are guidelines, not absolute rules
- Context matters - apply principles where they add value
- Don't refactor working code just to follow a principle
- Focus on code that is frequently read or modified
- Accept reasonable trade-offs for deadlines or constraints

**Core Principles to Emphasize:**
- Single Responsibility: Each function/class should do one thing well
- DRY (Don't Repeat Yourself): Extract common patterns into reusable functions
- YAGNI (You Aren't Gonna Need It): Don't add functionality until it's needed
- Fail Fast: Validate inputs early and return errors immediately
- Separation of Concerns: Keep business logic separate from infrastructure

**When to Refactor:**
- Code is difficult to understand or modify
- Methods exceed 30 lines without clear structure
- Logic is duplicated in multiple places
- Names don't accurately reflect behavior
- Tests are difficult to write or maintain

**When NOT to Refactor:**
- Code is working and rarely modified
- Refactoring would introduce risk without clear benefit
- Time constraints require shipping working code
- The "improvement" is purely aesthetic

**Source**
- Clean Code: A Handbook of Agile Software Craftsmanship (Robert C. Martin)
- The Pragmatic Programmer (Andrew Hunt, David Thomas)
- Code Complete (Steve McConnell)

## Examples and Anti-Patterns

This section provides concrete examples of well-structured documentation and common anti-patterns to avoid. Use these as reference when creating or updating documentation.

### Well-Structured Section Example

**Good Example: Component Description in `architecture.md`**

```markdown
### Document Monitor Lambda Function

The Document Monitor is a scheduled Lambda function that checks for new or updated documents in the source S3 bucket and triggers processing workflows.

**Responsibilities:**
- Poll S3 bucket every 5 minutes for new documents
- Validate document metadata and format
- Trigger Step Functions workflow for valid documents
- Log errors for invalid documents to CloudWatch

**Integration Points:**
- **Input**: S3 bucket `documents-incoming` (configured via `SOURCE_BUCKET_NAME` environment variable)
- **Output**: Step Functions state machine `DocumentProcessingWorkflow` (ARN in `WORKFLOW_ARN` environment variable)
- **Monitoring**: CloudWatch Logs group `/aws/lambda/document-monitor`

**Technology:**
- Runtime: Python 3.11
- Memory: 512 MB
- Timeout: 60 seconds
- Trigger: EventBridge rule (cron: `rate(5 minutes)`)

**Source**
- `src/document_monitor.py` - Lambda handler implementation
- `infra/lambda-stack.ts` - Lambda function infrastructure definition
- `infra/eventbridge-stack.ts` - EventBridge scheduling rule
```

**Why This Works:**
- Descriptive heading includes component name and type
- Clear, focused scope (400-600 tokens)
- Structured with subheadings for different aspects
- Specific technical details (runtime, memory, timeout)
- Integration points clearly identified
- Provenance links to actual code files
- Terminology is consistent ("Document Monitor", "Lambda function")

### Poorly-Structured Section Anti-Pattern

**Bad Example: Vague Component Description**

```markdown
### Details

The system has a function that runs periodically. It checks for stuff and does processing when needed.

It uses AWS services and connects to other parts of the system. Configuration is handled through environment variables.

See the code for more information.
```

**Why This Fails:**
- Generic heading ("Details") - not discoverable via RAG retrieval
- Vague language ("stuff", "does processing", "other parts")
- No specific technical details (runtime, memory, schedule)
- No integration points or dependencies identified
- No provenance - "see the code" is not helpful
- No structure - single paragraph instead of organized subsections
- Inconsistent terminology ("function" vs "Lambda function")

### Good vs. Bad Heading Examples

**Good Headings (Descriptive and Specific):**

From `architecture.md`:
- ✅ "Document Monitor Lambda Function"
- ✅ "DynamoDB Table Schema for Document Metadata"
- ✅ "Step Functions Workflow for Document Processing"
- ✅ "S3 Bucket Structure and Lifecycle Policies"

From `operations.md`:
- ✅ "Deploying the Document Processing Pipeline"
- ✅ "Monitoring CloudWatch Alarms for Lambda Failures"
- ✅ "Troubleshooting S3 Access Permission Errors"
- ✅ "Scaling DynamoDB Table Capacity"

From `api.md`:
- ✅ "POST /documents - Upload New Document"
- ✅ "GET /documents/{id} - Retrieve Document Metadata"
- ✅ "Authentication Using API Keys"
- ✅ "Error Response Format and Status Codes"

**Bad Headings (Generic and Non-Specific):**

- ❌ "Details" (details about what?)
- ❌ "Information" (what information?)
- ❌ "Overview" (overview of what?)
- ❌ "Configuration" (configuration of what?)
- ❌ "Setup" (setup of what?)
- ❌ "Usage" (usage of what?)
- ❌ "Notes" (notes about what?)
- ❌ "Miscellaneous" (completely non-descriptive)

**Heading Quality Checklist:**
- [ ] Heading includes specific component, feature, or concept name
- [ ] Heading is unique within the file
- [ ] Heading clearly indicates what content follows
- [ ] Heading contains key terms users would search for
- [ ] Heading is meaningful when read in isolation

### Proper Provenance Formatting

**Good Provenance Examples:**

**Example 1: Single Component**
```markdown
**Source**
- `src/handlers/document_upload.py` - Upload handler implementation
```

**Example 2: Component with Infrastructure**
```markdown
**Source**
- `src/document_processor.py` - Processing logic
- `infra/lambda-stack.ts` - Lambda infrastructure definition
- `infra/dynamodb-stack.ts` - DynamoDB table definition
```

**Example 3: Multiple Related Files**
```markdown
**Source**
- `src/api/routes.py` - API route definitions
- `src/api/auth.py` - Authentication middleware
- `src/api/validators.py` - Request validation logic
- `infra/api-gateway-stack.ts` - API Gateway infrastructure
```

**Example 4: Configuration and Code**
```markdown
**Source**
- `src/config/settings.py` - Configuration management
- `.env.example` - Environment variable template
- `infra/parameter-store.ts` - SSM Parameter Store setup
```

**Bad Provenance Examples:**

❌ **Too Vague:**
```markdown
**Source**
- See the source code
- Check the infrastructure folder
- Look at the Lambda functions
```

❌ **No Provenance:**
```markdown
(No Source section at all)
```

❌ **Outdated References:**
```markdown
**Source**
- `src/old_handler.py` (file no longer exists)
- `infra/deprecated-stack.ts` (file has been removed)
```

**Provenance Best Practices:**
- Always include a "Source" subsection for significant content
- Use specific file paths relative to repository root
- Include both implementation code and infrastructure code
- Keep provenance up-to-date when files are renamed or moved
- Add brief descriptions after file paths when helpful
- List files in logical order (implementation first, then infrastructure)

### Examples Covering All 6 Core Files

**Example 1: `overview.md` - Repository Purpose**

```markdown
## Purpose

This repository implements a document processing pipeline that ingests documents from S3, extracts metadata, and stores structured data in DynamoDB for retrieval via API.

**Key Capabilities:**
- Automated document ingestion from S3
- Metadata extraction using AWS Textract
- Structured storage in DynamoDB
- RESTful API for document retrieval
- Monitoring and alerting via CloudWatch

**Archon Integration:**
This repository participates in the Archon RAG system. Documentation under `.kiro/docs/` is ingested to provide context for automated agents and engineers.

**Source**
- `README.md` - High-level project description
- `infra/main-stack.ts` - Complete infrastructure definition
```

**Example 2: `architecture.md` - Component Description**

```markdown
### API Gateway REST API

The API Gateway provides a RESTful interface for document operations, including upload, retrieval, and search.

**Endpoints:**
- `POST /documents` - Upload new document
- `GET /documents/{id}` - Retrieve document by ID
- `GET /documents` - Search documents with filters

**Authentication:**
- API key authentication via `x-api-key` header
- Keys managed in API Gateway usage plans

**Integration:**
- Lambda proxy integration to `DocumentApiHandler` function
- Request validation using JSON Schema models
- CORS enabled for web client access

**Source**
- `infra/api-gateway-stack.ts` - API Gateway infrastructure
- `src/api/handler.py` - API Lambda handler
- `src/api/routes.py` - Route definitions
```

**Example 3: `operations.md` - Deployment Procedure**

```markdown
### Deploying the Document Processing Pipeline

**Prerequisites:**
- AWS CLI configured with appropriate credentials
- Node.js 18+ and npm installed
- Python 3.11+ installed

**Deployment Steps:**

1. Install dependencies:
   ```bash
   npm install
   pip install -r requirements.txt
   ```

2. Configure environment:
   ```bash
   cp .env.example .env
   # Edit .env with your AWS account ID and region
   ```

3. Deploy infrastructure:
   ```bash
   npm run deploy
   ```

4. Verify deployment:
   ```bash
   aws lambda list-functions --query 'Functions[?starts_with(FunctionName, `DocumentProcessor`)].FunctionName'
   ```

**Expected Output:**
- 3 Lambda functions deployed
- 1 DynamoDB table created
- 1 API Gateway REST API created
- CloudWatch log groups created for each Lambda

**Source**
- `package.json` - Deployment scripts
- `infra/main-stack.ts` - CDK stack entry point
- `.env.example` - Configuration template
```

**Example 4: `api.md` - Endpoint Documentation**

```markdown
### POST /documents - Upload New Document

Upload a new document for processing.

**Request:**

```http
POST /documents HTTP/1.1
Host: api.example.com
x-api-key: your-api-key
Content-Type: application/json

{
  "filename": "report.pdf",
  "s3_key": "incoming/report.pdf",
  "metadata": {
    "author": "John Doe",
    "department": "Engineering"
  }
}
```

**Response (Success):**

```http
HTTP/1.1 201 Created
Content-Type: application/json

{
  "document_id": "doc_abc123",
  "status": "processing",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Response (Error):**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": "InvalidRequest",
  "message": "filename is required"
}
```

**Source**
- `src/api/routes.py` - Route handler implementation
- `src/api/validators.py` - Request validation logic
- `infra/api-gateway-stack.ts` - API Gateway endpoint definition
```

**Example 5: `data-models.md` - Schema Documentation**

```markdown
### DynamoDB Table: DocumentMetadata

Stores metadata for all processed documents.

**Table Configuration:**
- Table name: `DocumentMetadata`
- Partition key: `document_id` (String)
- Sort key: None
- Billing mode: PAY_PER_REQUEST
- Point-in-time recovery: Enabled

**Attributes:**

| Attribute | Type | Description | Required |
|-----------|------|-------------|----------|
| `document_id` | String | Unique document identifier (UUID) | Yes |
| `filename` | String | Original filename | Yes |
| `s3_key` | String | S3 object key | Yes |
| `status` | String | Processing status (pending, processing, completed, failed) | Yes |
| `created_at` | String | ISO 8601 timestamp | Yes |
| `updated_at` | String | ISO 8601 timestamp | Yes |
| `metadata` | Map | Custom metadata key-value pairs | No |
| `extracted_text` | String | Extracted text content | No |

**Access Patterns:**
- Get document by ID: Query on `document_id`
- List all documents: Scan (use pagination for large datasets)

**Source**
- `infra/dynamodb-stack.ts` - Table infrastructure definition
- `src/models/document.py` - Document model class
- `src/repositories/document_repository.py` - Data access layer
```

**Example 6: `faq.md` - Common Question**

```markdown
### Why are some documents stuck in "processing" status?

**Symptom:**
Documents remain in "processing" status for more than 10 minutes and never complete.

**Common Causes:**

1. **Lambda timeout**: Processing Lambda may be timing out for large documents
   - Check CloudWatch Logs for timeout errors
   - Increase Lambda timeout in `infra/lambda-stack.ts` if needed

2. **Textract throttling**: AWS Textract may be throttling requests
   - Check CloudWatch metrics for Textract throttling
   - Implement exponential backoff in `src/textract_client.py`

3. **DynamoDB write failures**: Status updates may be failing
   - Check CloudWatch Logs for DynamoDB errors
   - Verify Lambda has correct IAM permissions

**Resolution Steps:**

1. Check Lambda logs:
   ```bash
   aws logs tail /aws/lambda/DocumentProcessor --follow
   ```

2. Check DynamoDB for stuck documents:
   ```bash
   aws dynamodb scan --table-name DocumentMetadata \
     --filter-expression "status = :status" \
     --expression-attribute-values '{":status":{"S":"processing"}}'
   ```

3. Manually retry processing:
   ```bash
   aws lambda invoke --function-name DocumentProcessor \
     --payload '{"document_id": "doc_abc123"}' response.json
   ```

**Source**
- `src/document_processor.py` - Processing logic
- `src/textract_client.py` - Textract integration
- CloudWatch Logs - Error patterns and diagnostics
```

### Key Takeaways

**For Well-Structured Documentation:**
- Use descriptive, specific headings with component/feature names
- Keep sections focused and appropriately sized (400-800 tokens)
- Include specific technical details (not vague descriptions)
- Provide clear integration points and dependencies
- Always include provenance with specific file paths
- Use consistent terminology throughout

**For Avoiding Anti-Patterns:**
- Never use generic headings like "Details", "Information", "Overview"
- Avoid vague language like "stuff", "things", "various"
- Don't omit technical specifics (runtime, memory, configuration)
- Don't skip provenance - always link to actual code
- Don't scatter related information across multiple sections
- Don't use inconsistent terminology for the same concept

**For All Documentation:**
- Think about RAG retrieval - would a semantic search find this content?
- Think about stakeholders - does this serve engineers, operators, and RAG agents?
- Think about maintenance - can this be updated incrementally as code changes?
- Think about grounding - is every claim verifiable from actual code?

## Security

Do not include:
- Secrets, tokens, or credentials
- Sensitive internal details without review
- Large external documents (summarize instead)

## Kiro Integration

This repository includes Kiro steering at `.kiro/steering/archon-docs.md` that enforces these standards automatically across all Kiro tasks.

When working with Kiro:
- Documentation updates should accompany code changes
- Kiro will help maintain documentation accuracy
- Kiro will ensure RAG-friendly structure

## Contract Authority

This `CLAUDE.md` file is the **authoritative contract** for documentation and code quality standards in this repository. It has **final authority** over all documentation practices, code quality requirements, and structural decisions.

### Precedence Rules

**CLAUDE.md has final authority:**
- When conflicts arise between this contract and any other guidance, this contract takes precedence
- All documentation and code quality decisions must align with this contract
- This contract defines the "what" - the standards that must be met

**Enforcement Mechanism:**
- `.kiro/steering/archon-docs.md` serves as the enforcement mechanism for this contract
- The steering file provides always-active guidance to Kiro on how to apply these standards
- The steering file defines the "how" - the workflows and processes for meeting the standards

**Conflict Resolution:**
- If conflicts arise between `CLAUDE.md` and `.kiro/steering/archon-docs.md`, defer to `CLAUDE.md`
- If conflicts arise between `CLAUDE.md` and other instructions or guidance, defer to `CLAUDE.md`
- When in doubt, read `CLAUDE.md` first to understand the authoritative standards
- Update the steering file to align with this contract if conflicts are discovered

**Relationship Between Contract and Steering:**
- `CLAUDE.md` (this file) = Authoritative standards and requirements
- `.kiro/steering/archon-docs.md` = Enforcement workflows and processes
- Both work together: contract defines standards, steering enforces them
- Steering must always defer to contract in case of conflicts

## Absolute Prohibition on Intermediary, Progress, or Ephemeral Files

This repository **must never contain intermediary, temporary, progress-tracking, scratch, or task-oriented files** created by humans or automated agents.

This includes (but is not limited to):

- `TASK_*.md`
- `*_PROGRESS.md`
- `NOTES.md`, `SCRATCH.md`, `DRAFT.md`
- `TEMP.md`, `WIP.md`
- Any file created to “think out loud”, track steps, or stage content before final integration
- Any file not explicitly part of the canonical documentation set

**This prohibition is absolute.**

### Rationale

Intermediary files:
- Break the **stable documentation surface** required for reliable RAG ingestion
- Pollute retrieval with partial, speculative, or outdated content
- Create ambiguity about what is canonical vs. transient
- Introduce cognitive overhead for both humans and agents
- Violate the fixed 6-file contract under `.kiro/docs/`

RAG systems, automated agents, and engineers must be able to assume:

> **If a file exists, it is canonical, intentional, and stable.**

### Required Behavior Instead

When working on documentation or reasoning through changes:

- **Perform reasoning internally** (agent scratchpad, chain-of-thought, or local context)
- **Apply changes directly** to the appropriate existing section(s) in the canonical files
- **Edit in place** within the correct file under `.kiro/docs/`
- **Use Git commits** as the only acceptable history of intermediate states

### Explicitly Disallowed Patterns

The following patterns are violations of this contract:

- Creating a temporary Markdown file to plan documentation
- Creating a progress file to track multi-step updates
- Creating a “draft” file before merging content
- Creating a new file “just to think” or “to organize thoughts”
- Creating files that are later intended to be deleted

> **If content is not ready to live permanently in one of the 6 canonical files, it must not be written to disk.**

### Enforcement for Automated Agents (Including Kiro)

Automated agents **must**:

- Never emit or suggest creation of intermediary files
- Never ask to create a temporary or progress file
- Never stage documentation outside the canonical files
- Treat any request to create such files as invalid

If an agent believes intermediary material is required:
- It must instead request clarification **in conversation**
- Or proceed with a best-effort update directly to canonical files

### Enforcement for Humans

Human contributors **must not**:

- Commit intermediary documentation files
- Request agents to create planning or progress files
- Leave behind scratch or task artifacts

