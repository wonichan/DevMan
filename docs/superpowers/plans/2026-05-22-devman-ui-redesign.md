# DevMan UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Modernize DevMan's existing six frontend pages with a Command Center / Raycast visual direction, add Toast notifications and confirmation dialogs, and preserve existing backend behavior.

**Architecture:** Build a lightweight internal design system in `frontend/src/components/ui/`, mount feedback providers at the app shell, then migrate pages from lowest-risk to highest-risk. Keep generated Wails files untouched and avoid new third-party component libraries unless explicitly approved later.

**Tech Stack:** Wails v2 frontend, React 18, TypeScript, Tailwind CSS 4, Vite.

---

## Source Documents

- Spec: `docs/superpowers/specs/2026-05-22-devman-ui-redesign-design.md`
- Existing frontend guide: `docs/FRONTEND.md`
- Project instructions: `AGENTS.md`

## Scope

Included:

- Command Center / Raycast-style redesign.
- Existing pages: `Dashboard`, `Environments`, `Migration`, `Cleaner`, `Versions`, `Settings`.
- Shared UI primitives: surface/card, button, status badge, progress bar, page header, empty state.
- Toast notifications.
- Confirmation dialogs for migration and cleaning.

Excluded:

- Global search / command palette.
- Theme switching or light mode.
- Keyboard shortcut system.
- Backend API/model changes.
- Edits to `frontend/src/bindings/go/` or `frontend/wailsjs/`.
- Git commits unless the user explicitly requests them.

## File Responsibility Map

Create:

- `frontend/src/components/icons.tsx` — inline SVG icon set for app chrome and actions.
- `frontend/src/components/ui/SurfaceCard.tsx` — base card/surface primitive.
- `frontend/src/components/ui/Button.tsx` — shared button variants.
- `frontend/src/components/ui/StatusBadge.tsx` — status chips for health/default/active/info states.
- `frontend/src/components/ui/ProgressBar.tsx` — shared progress/usage bar.
- `frontend/src/components/ui/PageHeader.tsx` — consistent page title/description/actions.
- `frontend/src/components/ui/EmptyState.tsx` — reusable empty state.
- `frontend/src/components/ui/Toast.tsx` — toast item and stack UI.
- `frontend/src/components/ui/ToastProvider.tsx` — toast state/context.
- `frontend/src/hooks/useToast.ts` — toast hook.
- `frontend/src/components/ui/ConfirmDialog.tsx` — confirmation dialog UI and provider.
- `frontend/src/hooks/useConfirm.ts` — confirm hook.

Modify:

- `frontend/src/style.css` — reduced motion and app atmosphere utilities.
- `frontend/tailwind.config.js` — only if animation/surface tokens are needed.
- `frontend/src/components/Panel.tsx` — keep as compatibility wrapper or migrate to `SurfaceCard`.
- `frontend/src/components/Sidebar.tsx` — SVG icon navigation, active rail, polished footer.
- `frontend/src/App.tsx` — app shell atmosphere and feedback providers.
- `frontend/src/pages/Settings.tsx` — low-risk primitive adoption.
- `frontend/src/pages/Versions.tsx` — grouped version rows and badges.
- `frontend/src/pages/Environments.tsx` — cards, search, empty state.
- `frontend/src/pages/Cleaner.tsx` — summary, rows, confirm, toast.
- `frontend/src/pages/Dashboard.tsx` — stat cards, ranking, loading skeleton, toast.
- `frontend/src/pages/Migration.tsx` — four-step command-center wizard, confirm, toast.

Delete only after verification:

- `frontend/src/App.css` — stale template CSS, currently not imported by `main.tsx`.

## Verification Commands

Run from `frontend/`:

```powershell
npm run build
```

Expected result: `tsc && vite build` exits with code 0.

When TypeScript-only verification is useful:

```powershell
npx tsc --noEmit
```

Expected result: exits with code 0.

Manual QA after implementation:

- Open the app via `wails dev` or the available frontend preview route.
- Navigate all six pages from the sidebar.
- Verify scan/analyze/clean/migrate buttons still call existing flows.
- Verify Cleaner shows confirm dialog before cleaning.
- Verify Migration shows confirm dialog before starting migration.
- Verify toasts appear for success/error paths that can be exercised.

---

## Chunk 1: Foundation

### Task 1: Remove stale App.css

**Files:**

- Delete: `frontend/src/App.css`

- [ ] **Step 1: Verify no imports**

Run a text search for `App.css` under `frontend/src/`.

Expected: no imports.

- [ ] **Step 2: Delete stale CSS**

Delete `frontend/src/App.css`.

- [ ] **Step 3: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 2: Add global motion and atmosphere utilities

**Files:**

- Modify: `frontend/src/style.css`
- Modify if needed: `frontend/tailwind.config.js`

- [ ] **Step 1: Add reduced-motion protection**

Add a `prefers-reduced-motion: reduce` rule that minimizes animations/transitions.

- [ ] **Step 2: Add app atmosphere utility**

Add a reusable dark radial-gradient background utility for the app shell.

- [ ] **Step 3: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 3: Create inline SVG icon set

**Files:**

- Create: `frontend/src/components/icons.tsx`

- [ ] **Step 1: Add shared icon prop type**

Use `React.SVGProps<SVGSVGElement>`.

- [ ] **Step 2: Export consistent icons**

Minimum icons: Dashboard, Environments, Migration, Cleaner, Versions, Settings, Refresh, Search, Close, Check, Warning, Info, ArrowRight, ArrowLeft, Trash, Loader.

- [ ] **Step 3: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

---

## Chunk 2: UI Primitives

### Task 4: Create SurfaceCard and preserve Panel compatibility

**Files:**

- Create: `frontend/src/components/ui/SurfaceCard.tsx`
- Modify: `frontend/src/components/Panel.tsx`

- [ ] **Step 1: Implement `SurfaceCard`**

Variants: `default`, `raised`, `interactive`, `selected`, `danger`.

- [ ] **Step 2: Keep `Panel` as a wrapper**

Update `Panel.tsx` to render `SurfaceCard` so existing pages keep compiling while migration proceeds.

- [ ] **Step 3: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 5: Create Button

**Files:**

- Create: `frontend/src/components/ui/Button.tsx`

- [ ] **Step 1: Implement variants**

Variants: `primary`, `secondary`, `danger`, `ghost`.

- [ ] **Step 2: Implement sizes**

Sizes: `sm`, `md`, `lg`.

- [ ] **Step 3: Add accessible focus and disabled states**

Use visible focus rings and disabled opacity/cursor states.

- [ ] **Step 4: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 6: Create StatusBadge

**Files:**

- Create: `frontend/src/components/ui/StatusBadge.tsx`

- [ ] **Step 1: Implement status variants**

Statuses: `healthy`, `warning`, `danger`, `info`, `active`, `default`.

- [ ] **Step 2: Include text labels**

Do not rely on color alone.

- [ ] **Step 3: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 7: Create ProgressBar

**Files:**

- Create: `frontend/src/components/ui/ProgressBar.tsx`

- [ ] **Step 1: Implement bounded value handling**

Clamp input value to 0-100.

- [ ] **Step 2: Implement visual variants**

Colors: `accent`, `info`, `warning`, `danger`, plus optional gradient.

- [ ] **Step 3: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 8: Create PageHeader

**Files:**

- Create: `frontend/src/components/ui/PageHeader.tsx`

- [ ] **Step 1: Implement title/description/actions layout**

Use compact command-center spacing.

- [ ] **Step 2: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 9: Create EmptyState

**Files:**

- Create: `frontend/src/components/ui/EmptyState.tsx`

- [ ] **Step 1: Implement icon/title/description/action props**

Action should use the shared `Button`.

- [ ] **Step 2: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

---

## Chunk 3: Feedback And Shell

### Task 10: Create Toast system

**Files:**

- Create: `frontend/src/components/ui/Toast.tsx`
- Create: `frontend/src/components/ui/ToastProvider.tsx`
- Create: `frontend/src/hooks/useToast.ts`

- [ ] **Step 1: Implement provider state**

Manage a stack of toasts with id, variant, title/message, and dismiss callback.

- [ ] **Step 2: Implement hook API**

Expose `toast.success`, `toast.error`, `toast.info`, `toast.warning`.

- [ ] **Step 3: Implement toast stack UI**

Top-right or bottom-right inside the Wails window, auto-dismiss, manual close.

- [ ] **Step 4: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 11: Create ConfirmDialog system

**Files:**

- Create: `frontend/src/components/ui/ConfirmDialog.tsx`
- Create: `frontend/src/hooks/useConfirm.ts`

- [ ] **Step 1: Implement provider state**

Support a promise-based `confirm(options)` API.

- [ ] **Step 2: Implement dialog UI**

Title, description, cancel button, confirm button, danger/default variants.

- [ ] **Step 3: Add practical focus behavior**

Focus the cancel or safe action on open; close on cancel.

- [ ] **Step 4: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 12: Redesign Sidebar

**Files:**

- Modify: `frontend/src/components/Sidebar.tsx`

- [ ] **Step 1: Replace emoji nav icons**

Use icons from `frontend/src/components/icons.tsx`.

- [ ] **Step 2: Add active rail and raised active state**

Selected item gets left accent rail, raised background, and accent text.

- [ ] **Step 3: Add stable hover and footer polish**

No layout-shifting scale effects.

- [ ] **Step 4: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 13: Redesign App shell

**Files:**

- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: Mount feedback providers**

Mount Toast and ConfirmDialog providers/hosts near the app root.

- [ ] **Step 2: Add atmosphere and content layout**

Keep the existing page switcher, but polish the background and scroll container.

- [ ] **Step 3: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

---

## Chunk 4: Lower-Risk Pages

### Task 14: Redesign Settings

**Files:**

- Modify: `frontend/src/pages/Settings.tsx`

- [ ] **Step 1: Replace inline header with PageHeader**

- [ ] **Step 2: Convert sections to SurfaceCard**

- [ ] **Step 3: Improve toggles with role/focus states**

- [ ] **Step 4: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 15: Redesign Versions

**Files:**

- Modify: `frontend/src/pages/Versions.tsx`

- [ ] **Step 1: Use PageHeader and SurfaceCard groups**

- [ ] **Step 2: Use StatusBadge for default/active states**

- [ ] **Step 3: Use EmptyState when no environments exist**

- [ ] **Step 4: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 16: Redesign Environments

**Files:**

- Modify: `frontend/src/pages/Environments.tsx`

- [ ] **Step 1: Use PageHeader and shared Button/Search styling**

- [ ] **Step 2: Convert environment cards to SurfaceCard**

- [ ] **Step 3: Use EmptyState for no results**

- [ ] **Step 4: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 17: Redesign Cleaner

**Files:**

- Modify: `frontend/src/pages/Cleaner.tsx`

- [ ] **Step 1: Use PageHeader, SurfaceCard, Button, Progress/summary styling**

- [ ] **Step 2: Add ConfirmDialog before `CleanItems()`**

- [ ] **Step 3: Add toast feedback for analyze/clean success and failure**

- [ ] **Step 4: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

---

## Chunk 5: Higher-Risk Pages

### Task 18: Redesign Dashboard

**Files:**

- Modify: `frontend/src/pages/Dashboard.tsx`

- [ ] **Step 1: Use PageHeader and shared refresh Button**

- [ ] **Step 2: Convert metrics to command-center stat cards**

- [ ] **Step 3: Convert ranking bars to shared ProgressBar**

- [ ] **Step 4: Add loading skeleton and toast feedback**

- [ ] **Step 5: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

### Task 19: Redesign Migration

**Files:**

- Modify: `frontend/src/pages/Migration.tsx`

- [ ] **Step 1: Preserve handlers and state logic**

Do not change the signatures or core behavior of `loadEnvs`, `loadDisks`, or `handleMigrate` during the first visual pass.

- [ ] **Step 2: Replace stepper and step layouts with shared primitives**

- [ ] **Step 3: Add ConfirmDialog before migration begins**

- [ ] **Step 4: Add toast feedback for migration success/failure**

- [ ] **Step 5: Build**

Run: `npm run build` in `frontend/`.

Expected: build passes.

- [ ] **Step 6: Manual workflow check**

Check step navigation, environment selection, target path editing, preview, confirmation, and result/log display.

---

## Execution Order

1. Chunk 1: Foundation.
2. Chunk 2: UI primitives.
3. Chunk 3: feedback providers and shell.
4. Chunk 4: Settings, Versions, Environments, Cleaner.
5. Chunk 5: Dashboard, then Migration.

Dashboard and Migration should not start until shared primitives and providers are stable.

## Review Gates

- After Chunk 2: verify primitives are consistent before page rewrites.
- After Chunk 3: verify providers mount without breaking routing.
- After Chunk 4: verify low-risk pages establish reusable patterns.
- After Chunk 5: run full frontend build and manual UI pass.

## Final Acceptance Checklist

- [ ] `npm run build` passes in `frontend/`.
- [ ] Six pages render without runtime errors.
- [ ] Sidebar uses SVG icons for app chrome.
- [ ] Toast and ConfirmDialog are implemented and mounted.
- [ ] Cleaner asks for confirmation before cleaning.
- [ ] Migration asks for confirmation before migrating.
- [ ] Existing backend calls are preserved.
- [ ] No global search, theme switching, or keyboard shortcut system is added.
- [ ] Generated Wails files are untouched.
