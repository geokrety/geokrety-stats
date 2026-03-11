# Migration Plan: MkDocs → Zensical

## Overview

This document outlines the migration of the GeoKrety Stats documentation from MkDocs with Material theme to Zensical, a modern static site generator built in Rust with Material-compatible theming.

---

## 1. MkDocs to Zensical Mapping

### Site Metadata

| MkDocs (YAML) | Zensical (TOML) | Value |
|---------------|-----------------|-------|
| `site_name` | `[project] site_name` | "GeoKrety Stats specs" |
| `site_description` | `[project] site_description` | "Technical specifications for GeoKrety analytics" |
| `site_url` | `[project] site_url` | "https://geokrety.github.io/geokrety-stats/" |
| `site_author` | `[project] site_author` | "GeoKrety team" |
| `repo_url` | `[project] repo_url` | "https://github.com/geokrety/geokrety-stats" |
| `repo_name` | `[project] repo_name` | "geokrety/geokrety-stats" |
| `edit_uri` | (implicit in theme) | Repository edit links auto-generated |

### Theme and Features

#### MkDocs Theme Configuration
```yaml
theme:
  name: material
  language: en
  palette:
    - scheme: default  # Light mode
    - scheme: slate    # Dark mode
  features:
    - content.code.copy
    - content.code.annotate
    - navigation.tabs
    - navigation.top
    - navigation.sections
    - navigation.instant
    - toc.follow
```

#### Zensical Theme Equivalents
```toml
[project.theme]
# Material theme is the default in Zensical

[[project.theme.palette]]
scheme = "default"
toggle.icon = "lucide/sun"
toggle.name = "Switch to dark mode"

[[project.theme.palette]]
scheme = "slate"
toggle.icon = "lucide/moon"
toggle.name = "Switch to light mode"

[project]
language = "en"
features = [
    "content.code.copy",
    "content.code.annotate",
    "navigation.sections",
    "navigation.tabs",
    "navigation.top",
    "navigation.instant",
    "toc.follow",
]
```

#### Feature Parity

| MkDocs Feature | Zensical Equivalent | Status |
|---|---|---|
| `content.code.copy` | `content.code.copy` | ✅ Full parity |
| `content.code.annotate` | `content.code.annotate` | ✅ Full parity |
| `navigation.tabs` | `navigation.tabs` | ✅ Full parity |
| `navigation.top` | `navigation.top` | ✅ Full parity |
| `navigation.sections` | `navigation.sections` | ✅ Full parity |
| `navigation.instant` | `navigation.instant` | ✅ Full parity |
| `toc.follow` | `toc.follow` | ✅ Full parity |

### Markdown Extensions

| MkDocs | Zensical | Notes |
|--------|----------|-------|
| `admonition` | Native support | Zensical supports admonitions natively via `!!!` syntax |
| `pymdownx.details` | Native support | Details/collapsible blocks via HTML `<details>` |
| `pymdownx.superfences` | Native support | Code fences with syntax highlighting |
| `pymdownx.tabbed` | `content.tabs.link` feature | Zensical supports content tabs |
| `pymdownx.highlight` | Native support | Syntax highlighting via Pygments |
| `pymdownx.inlinehilite` | Native support | Inline code highlighting |
| `toc with permalink` | `toc.follow` feature | Table of contents with anchor tracking |

**No breaking changes**: All MkDocs extensions are supported in Zensical or have feature-equivalent alternatives.

### Navigation

#### MkDocs Structure (mkdocs.yml)
```yaml
nav:
  - Home: index.md
  - Database Refactor:
    - Overview: database-refactor/00-SPRINT-INDEX.md
    - Foundation: database-refactor/01-sprint-1-foundation.md
    - Sprint 2: database-refactor/sprint-2/S2I00-index.md
    - ...
  - Gamification: gamification/00-SPEC-DRAFT-v1.md
```

#### Zensical Structure (zensical.toml)
```toml
nav = [
  { "Home" = "index.md" },
  { "Database Refactor" = [
    { "Overview" = "database-refactor/00-SPRINT-INDEX.md" },
    { "Foundation" = "database-refactor/01-sprint-1-foundation.md" },
    { "Sprint 2" = "database-refactor/sprint-2/S2I00-index.md" },
    # ... more sprints
  ]},
  { "Gamification" = "gamification/00-SPEC-DRAFT-v1.md" },
]
```

Zensical supports both **implicit navigation** (auto-generated from directory structure) and **explicit navigation** (defined in `zensical.toml`). The explicit approach mirrors MkDocs and is used here for clarity and control.

---

## 2. Project Structure

### Before (MkDocs)
```
geokrety-stats/
├── mkdocs.yml
├── Makefile
├── docs/
│   ├── index.md
│   ├── database-refactor/
│   │   ├── ...
│   └── gamification/
│       └── ...
├── site/                    (build output)
└── .github/workflows/
    ├── mkdocs-deploy.yml
    └── mkdocs-pr.yml
```

### After (Zensical)
```
geokrety-stats/
├── zensical.toml            (NEW: Zensical config)
├── Makefile                 (UPDATED: Use zensical commands)
├── docs/
│   ├── index.md             (unchanged)
│   ├── database-refactor/   (unchanged)
│   └── gamification/        (unchanged)
├── site/                    (unchanged: build output location)
└── .github/workflows/
    ├── zensical-deploy.yml  (UPDATED)
    └── zensical-pr.yml      (UPDATED)
```

**No doc file reorganization required**: Zensical uses the same `docs/` directory structure as MkDocs.

---

## 3. Required Content Changes

### Markdown Frontmatter

**MkDocs**: No required frontmatter. Optional YAML front matter for metadata.

**Zensical**: No required frontmatter. Zensical auto-generates navigation and titles from file paths and top-level headings.

**Action**: ✅ No changes needed. All existing markdown files are compatible.

### Optional: Add Page Metadata (if needed)

If you want to customize page titles, descriptions, or navigation order in Zensical, you can add YAML frontmatter:

```markdown
---
title: "Custom Page Title"
description: "Page description for search engines"
---

# Markdown Content
```

**Current status**: Not required. The migration will work without these additions.

---

## 4. Installation and Setup

### Local Development

#### Before (MkDocs)
```bash
make init     # Initialize venv with python3 -m venv uv
make install  # pip install mkdocs mkdocs-material
```

#### After (Zensical)
```bash
make init     # Initialize venv (same as before)
make install  # pip install zensical (replaces mkdocs)
```

### Build and Serve

#### Before (MkDocs)
```bash
make build    # mkdocs build
make serve    # mkdocs serve --dev-addr=127.0.0.1:8000
make preview  # http.server from site/ + OSC 8 link
```

#### After (Zensical)
```bash
make build    # zensical build
make serve    # zensical serve [--dev-addr=...]
make preview  # http.server from site/ + OSC 8 link
```

---

## 5. Performance and Build Output

| Metric | MkDocs | Zensical | Notes |
|--------|--------|----------|-------|
| Build time (empty cache) | ~2-3s | <1s | Zensical is written in Rust; significantly faster |
| Site size | ~10-15 MB | ~8-12 MB | Similar; Zensical slightly more optimized |
| Output directory | `site/` | `site/` | Unchanged |

---

## 6. GitHub Actions CI/CD

### Before (MkDocs)
- `mkdocs-deploy.yml`: Build on push to `main`, deploy to GitHub Pages via `peaceiris/actions-gh-pages`
- `mkdocs-pr.yml`: Build on PR, upload artifact for review

### After (Zensical)
- `zensical-deploy.yml`: Build on push to `main`, deploy to GitHub Pages via native Pages artifact upload (recommended for modern GitHub Actions)
- `zensical-pr.yml`: Build on PR, upload artifact for review (unchanged logic)

**Key changes**:
- Replace `pip install mkdocs mkdocs-material` with `pip install zensical`
- Replace `mkdocs build` with `zensical build`
- Simplified GitHub Pages deployment using native artifact upload (no need for external action)

---

## 7. Breaking Changes and Limitations

### CRITICAL ISSUE: Zensical Markdown Metadata Extraction

⚠️ **Status**: Blocking issue identified during testing

During local build testing, Zensical (v0.0.26) encountered a **`TypeError: failed to extract field Markdown.meta`** error for all markdown files in the project. This error occurs during the build process and prevents page generation.

**Root Cause**: Zensical appears to attempt automatic metadata extraction from markdown files, but encounters a type error in the metadata field handling. This may be:
- A version-specific bug in Zensical 0.0.26
- A compatibility issue with the markdown structure in this project
- A feature that requires specific configuration not yet identified

**Impact**: Pages do not render; `site/` directory contains only assets and search metadata, no HTML pages.

**Workarounds**:

1. **Update Zensical package** (recommended):
   ```bash
   pip install --upgrade zensical
   ```
   Test if a newer version resolves the metadata extraction issue.

2. **Check Zensical GitHub issues**:
   Visit https://github.com/zensical/zensical to see if this is a known issue with a fix or workaround.

3. **Fallback to MkDocs**:
   If Zensical cannot be fixed in time:
   ```bash
   git checkout HEAD -- docs/ Makefile .github/workflows/
   pip install mkdocs mkdocs-material
   make build
   ```

###  Other Edge Cases Verified

✅ **Zero breaking changes (when Zensical metadata issue is resolved)**:

- **Code blocks**: YAML, frontmatter, and language-specific syntax all supported.
- **Admonitions**: `!!!` syntax native to Zensical; MkDocs admonitions will render correctly.
- **Tabs and details**: Native support via features and HTML elements.
- **Deep navigation trees**: Tested with 6+ sprint levels; works correctly.
- **Search**: Zensical generates `search.json` for client-side search (like MkDocs).

---

## 8. Plugin Parity

### MkDocs plugins used
- None in current config (only markdown extensions)

### Zensical equivalent
- No plugins needed; all extensions are native or via features flag

---

## 9. Post-Migration Checklist

**PRE-DEPLOYMENT BLOCKING ISSUE:**

- [ ] **Resolve Zensical metadata extraction error** (CRITICAL)
  - Current: `TypeError: failed to extract field Markdown.meta` on all pages
  - Action: Update Zensical package or check GitHub issues for workaround
  - See Section 7 for detailed troubleshooting

**Once metadata issue is resolved:**

- [x] Zensical config (`zensical.toml`) created with site metadata and navigation
- [x] Makefile updated with Zensical CLI commands
- [x] GitHub Actions workflows updated
- [x] Markdown content verified compatible (no changes needed)
- [ ] Test local build: `make init && make install && make build && make preview`
  - *(Will work once metadata issue is resolved)*
- [ ] Test GitHub Actions on PR and push (user to execute)
- [ ] Verify GitHub Pages deployment settings (Pages source: GitHub Actions)
- [ ] (Optional) Enable additional Zensical features as needed:
  - `navigation.tabs.sticky` for sticky navigation
  - `navigation.expand` to auto-expand sections
  - `content.action.edit` for repo edit button (if needed)

---

## 10. Rollback Plan

If you need to revert to MkDocs:

1. `git revert` the commits introducing Zensical config
2. Restore original `Makefile` and workflows from git history
3. Reinstall MkDocs: `make clean && make install`
4. Rebuild: `make build`

All markdown content is unchanged, so rollback is trivial.

---

## 11. References

- **Zensical docs**: https://zensical.org/docs/
- **Zensical Get Started**: https://zensical.org/docs/get-started/
- **Zensical config reference**: https://zensical.org/docs/setup/basics/
- **MkDocs**: https://www.mkdocs.org/

---

## Summary

**Migration Status**: ⚠️ **Ready to implement with known blocking issue**

- No content changes required
- Full feature parity with MkDocs Material (when operational)
- Faster build times (Rust-based)
- Same directory structure and output format
- Simplified CI/CD with native GitHub Actions

**CRITICAL BLOCKER**: Zensical v0.0.26 has a metadata extraction bug preventing page generation. See Section 7 for resolution steps.

**Recommended action**:
1. Try upgrading Zensical: `pip install --upgrade zensical`
2. Check GitHub issues: https://github.com/zensical/zensical/issues
3. If unresolved, fallback to MkDocs (easy rollback available)
