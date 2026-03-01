---
description: UI layout standards - tabs with counts and tooltips, sortable table headers, composable reuse patterns
applyTo: '**'
---

# UI Development Standards

## Tabs with Counts and Tooltips

Every tabbed interface **must** follow these rules:

### a. Tab Labels Must Include a Count Badge
Always display the number of items in each tab label:

```vue
<v-tab value="countries" :text="`Countries (${countriesCount})`" />
<!-- or with a chip -->
<v-tab value="countries">
  Countries
  <v-chip size="x-small" class="ml-2">{{ countriesCount }}</v-chip>
</v-tab>
```

- Count must update reactively when data changes
- Show `0` explicitly, never hide the badge when empty
- Use a loading placeholder (e.g. `…`) while data is being fetched

### b. Tabs Must Have Descriptive Tooltips
Wrap each tab in a tooltip explaining what it contains:

```vue
<v-tooltip text="List of countries where this cache has been found">
  <template #activator="{ props }">
    <v-tab v-bind="props" value="countries">
      Countries
      <v-chip size="x-small" class="ml-2">{{ countriesCount }}</v-chip>
    </v-tab>
  </template>
</v-tooltip>
```

- Tooltip text must describe the **content** of the tab, not just repeat its label
- Keep tooltip text under 80 characters

---

## Table Headers

### c. Column Headers Must Have Tooltips
Every `<th>` or sortable header must expose a description via tooltip:

```vue
<v-tooltip text="The country where the log was recorded">
  <template #activator="{ props }">
    <th v-bind="props">Country</th>
  </template>
</v-tooltip>
```

- Tooltip text should clarify the **meaning or source** of the data, not just its name
- Technical column names (IDs, codes) must always have a tooltip

### d. Table Headers Must Support Backend Sorting
All data tables must implement server-side sorting. Use a consistent sort composable/helper:

```vue
<th
  class="sortable"
  :class="sortClass('country')"
  @click="setSort('country')"
>
  <v-tooltip text="Sort by country name">
    <template #activator="{ props }">
      <span v-bind="props">
        Country
        <v-icon size="x-small">{{ sortIcon('country') }}</v-icon>
      </span>
    </template>
  </v-tooltip>
</th>
```

Sorting state must be:
- Reflected in the URL query params (`?sort=country&order=asc`) for shareability
- Sent to the backend API on each change (never sort client-side on paginated data)
- Preserved across tab switches within the same page

---

## Composable Reuse Policy

When creating a **new composable**, always:

1. **Search first** — scan existing composables for any that solve the same or a similar problem
2. **Prefer extension over duplication** — if an existing composable covers 80%+ of the use case, extend or wrap it rather than creating a new one
3. **Replace, don't stack** — if the new composable fully supersedes an old one, refactor call sites to use the new one and delete the old file
4. **Document the replacement** in the PR description: what was replaced and why

```
# Example composable audit checklist (add to PR description)
- [ ] Searched for existing composables handling: sorting / filtering / pagination / tabs
- [ ] No duplicate found → new composable justified
- [ ] OR: Replaced `useOldSorting` with new `useSortableTable` in N files
```

---

## MCP Serena Workflow

When consulting or editing project files:

1. Load tools via `tool_search` using pattern: `serena` or `mcp_serena`
2. Activate project: `mcp_serena_activate_project` (point to repo root)
3. Before writing or editing code:
   - Find existing composables: `mcp_serena_find_symbol` + `mcp_serena_search_files`
   - Read relevant files: `mcp_serena_read_file`
4. Confirm file paths and symbol names exist before referencing them

---

## MCP Playwright Workflow

1. Load tools via `tool_search_tool_regex` using pattern: `^mcp_microsoft_pla_browser`
2. Navigate: `mcp_microsoft_pla_browser_navigate`
3. Resize viewport:
   - Mobile: 720×2048
   - Desktop: 1280×1024
4. Screenshot: `mcp_microsoft_pla_browser_take_screenshot`
5. Validate:
   - Responsive layout — no overflow, no hidden elements
   - No broken tables
   - No JS console errors
   - Accessibility basics
   - Dark/light theme consistency

Repeat until layout is correct.
