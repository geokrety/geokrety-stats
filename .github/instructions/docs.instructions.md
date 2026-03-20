---
description: 'Instructions for building the GeoKrety Stats Documentation, including usage examples, technical implementation details, and updated help documentation.'
applyTo: 'docs/**/*.md,zensical.toml'
---

# GeoKrety Stats Documentation Instructions

- When a new folder is added to the `docs/` directory, it MUST be automatically included in the documentation index #file:./../../zensical.toml
When a new file is added to the `docs/` directory, it MUST be included in the documentation index if relevant, and linked from the appropriate parent page.
- All documentation files MUST follow the existing style and formatting conventions used in the current `#file:../../docs/log-at-home/` directory.
- After each update the documentation build MUST be tested locally to ensure all links work and formatting is correct, using the #file:../../Makefile.docs
