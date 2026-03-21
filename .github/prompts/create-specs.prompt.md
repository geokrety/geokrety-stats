---
agent: 'specification'
tools: [vscode/askQuestions, execute/runInTerminal, read/problems, read/readFile, agent, edit/createFile, edit/editFiles, search/changes, search/codebase, search/fileSearch, search/listDirectory, search/textSearch, postgresql-mcp/pgsql_connect, postgresql-mcp/pgsql_disconnect, postgresql-mcp/pgsql_list_databases, postgresql-mcp/pgsql_query]
description: 'Generate a new specification file in the docs.'
model: 'Claude Haiku 4.5'
---

# ABSOLUTE GOAL

- create a specification file in #file:../../docs/xxx/specification.md (where `xxx` is a unique identifier for this specification).

## Instructions

- you MUST create a "specs folder" in #file:../../docs/xxx/ where `xxx` is a unique identifier for this prompt.
- you MUST add the link into the index file at #file:../../docs/index.md to ensure discoverability of the new specification.
- you MUST use agent "specification" to generate the specification file based on the user input and the context provided by the codebase and database.
- you MUST ensure that the specification file is comprehensive, clear, and well-structured, covering all relevant aspects of the feature or functionality being specified.
- you MUST use the template provided below to structure the specification file, and you MUST fill in each section with detailed and accurate information based on the requirements and design of the feature.
- once you have created the specification, you MUST use the "critical-loop" skill with agents "specification" -> "technical-writer" -> "requirements-analyst" -> "critical-thinking" to review the generated specification file and ensure that it meets the requirements, is of high quality, and is ready for implementation.

# specification file template

```md
# Specification for [Feature/Functionality Name]

## Overview
- Brief description of the feature or functionality being specified.

## Requirements
- List of functional and non-functional requirements that the feature must meet.

## Design
- Detailed design of the feature, including architecture, components, and interactions.

## Implementation Details
- Specific implementation details, such as algorithms, data structures, and technologies to be used.

## Testing
- Description of the testing strategy, including types of tests to be performed and tools to be used

## Acceptance Criteria
- Clear and measurable criteria that must be met for the feature to be considered complete and ready for deployment.

## Diagrams and Examples
- Any relevant diagrams, examples, or references to support the specification and enhance understanding.

## Deliverables
- List of deliverables that will be produced as part of the implementation of this feature.

## Implementation checklist

```
