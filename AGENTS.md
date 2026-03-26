# AGENTS.md - Guidelines for LLM Agents

This document provides guidelines for LLM agents (including AI assistants) working in this repository. These guidelines ensure consistency, quality, and proper semantic versioning practices.

## Overview

FIFA Bot is a real-time FIFA match monitoring bot that sends notifications to Slack. When making changes to this repository, follow the guidelines in this document to ensure your work aligns with project standards.

## Semantic Versioning Requirements

All pull requests MUST be labeled with exactly ONE of the following labels to indicate the type of change:

### Label: `major` 🔴

Use this label for **breaking changes** that require a major version bump (e.g., 1.0.0 → 2.0.0).

**Examples of breaking changes:**
- Removing public API functions or methods
- Changing function signatures in a non-backward-compatible way
- Changing required environment variables
- Removing or renaming exported types or interfaces
- Changing behavior in a way that breaks existing deployments
- Removing support for a previously supported feature

### Label: `minor` 🟢

Use this label for **new features** that are backward-compatible (e.g., 1.0.0 → 1.1.0).

**Examples of new features:**
- Adding new public functions or methods
- Adding new optional configuration options
- Expanding functionality without breaking existing code
- Adding new supported match events or notification types
- Enhancing existing features with new capabilities

### Label: `patch` 🔵

Use this label for **bug fixes** and other non-breaking changes (e.g., 1.0.0 → 1.0.1).

**Examples of patch changes:**
- Fixing bugs in match event detection
- Fixing Redis connection issues
- Optimizing performance without API changes
- Fixing notification formatting
- Updating dependencies with bugfixes
- Fixing typos in documentation or code
- Refactoring code without changing behavior

## PR Requirements

When creating a pull request, you MUST:

1. **Add exactly one semver label** (`major`, `minor`, or `patch`)
   - Use `gh pr create` with labels or add labels manually after creation
   - Example: `gh pr create --title "feat: new feature" --label "minor" --body "..."`

2. **Use conventional commit format**
   - Commit subjects: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, etc.
   - Example: `feat: add support for penalty miss notifications`
   - Example: `fix: resolve duplicate notification bug`

3. **Include AI disclosure statement** (required for all AI-assisted work)
   ```markdown
   __Disclosure__
   This change was developed with the assistance of AI, but was reviewed and tested by a human.
   ```

4. **Write a clear PR description**
   - Summarize what changed and why
   - List key changes as bullet points
   - Reference relevant issues if applicable

## Making Code Changes

### Code Style

- Follow existing code patterns in the repository
- Match indentation and formatting style
- Use the same naming conventions (camelCase for Go)
- Keep functions focused and well-documented

### Testing

- Run `go test ./...` before committing
- Ensure all tests pass
- Add tests for new functionality when possible
- Don't break existing tests

### Documentation

- Update README.md if you change configuration options
- Add inline comments for complex logic
- Document new exported functions and types
- Keep documentation in sync with code changes

## Project Structure

```
FIFA Bot/
├── cmd/server.go              # Application entry point
├── pkg/
│   ├── app/                   # Core application logic
│   ├── fifa/                  # FIFA API integration
│   ├── database/              # Redis operations
│   └── models/                # Data structures
├── README.md                  # Project documentation
└── AGENTS.md                  # This file
```

## Common Task Examples

### Adding a New Feature

1. Branch: `git checkout -b feat/feature-name`
2. Implement feature with tests
3. Commit: `git commit -m "feat: add new feature"`
4. Push and create PR with `minor` label

### Fixing a Bug

1. Branch: `git checkout -b fix/bug-name`
2. Fix the issue with tests
3. Commit: `git commit -m "fix: resolve bug description"`
4. Push and create PR with `patch` label

### Making a Breaking Change

1. Branch: `git checkout -b feat/breaking-change`
2. Implement the breaking change
3. Commit: `git commit -m "feat!: breaking change description"`
4. Push and create PR with `major` label
5. Add a note in PR body explaining the breaking change

### Updating Documentation

1. Branch: `git checkout -b docs/update-name`
2. Update documentation files
3. Commit: `git commit -m "docs: update documentation"`
4. Push and create PR with `patch` label

## Environment Variables

When working on configuration or environment handling:

- **Required variables**: `SLACK_WEBHOOK_URL`, `REDIS_ADDRESS`, `REDIS_DB`
- **Optional variables**: See README.md for full list
- Changes to required variables = `major` label
- New optional variables = `minor` label
- Fixes to env handling = `patch` label

## Slack Integration

When modifying Slack notification logic:

- Ensure messages are properly formatted
- Test with actual Slack webhook if possible
- Maintain backward compatibility with Slack API
- Breaking changes to message format = `major` label

## Redis Database

When modifying Redis persistence:

- Ensure state is properly serialized/deserialized
- Test connection retry logic
- Avoid breaking schema changes without migration path
- Breaking schema changes = `major` label

## Decision Matrix

Use this matrix to determine the correct label:

| Change Type | Breaking? | Label |
|-------------|-----------|-------|
| Bug fix | No | `patch` |
| Performance optimization | No | `patch` |
| Refactoring | No | `patch` |
| Documentation | No | `patch` |
| New optional feature | No | `minor` |
| New public API | No | `minor` |
| Enhanced functionality | No | `minor` |
| Removed feature | Yes | `major` |
| API change | Yes | `major` |
| Required env var change | Yes | `major` |
| Config schema change | Yes | `major` |

## Questions or Ambiguity?

If you're unsure which label to use:

1. Default to `patch` if uncertain (it's safer)
2. Document your reasoning in the PR description
3. Consider the impact on users running the bot in production
4. When in doubt, ask: "Will this require users to change their setup?"

## Reminders

✅ DO:
- Add exactly one semver label to every PR
- Write clear commit messages
- Include the AI disclosure statement
- Run tests before pushing
- Keep commits focused and atomic
- Update documentation when needed

❌ DON'T:
- Forget to label your PR
- Add multiple version labels to one PR
- Skip testing
- Make unrelated changes in one commit
- Leave TODO comments instead of implementing
- Skip the disclosure statement for AI-assisted work

---

**Last Updated**: March 26, 2026

This document applies to all LLM agents and AI assistants working on this repository.
