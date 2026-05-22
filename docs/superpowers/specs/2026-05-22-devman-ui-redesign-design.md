# DevMan UI Redesign Design

## Summary

DevMan will be redesigned as a modern dark developer command center. The chosen direction is **Command Center / Raycast style**: high-density, keyboard-friendly visual hierarchy, dark layered surfaces, green run-state accents, blue information accents, and restrained motion.

The implementation scope is **B**: modernize the existing app UI across the current six pages and add two feedback primitives: Toast notifications and confirmation dialogs. Global search, theme switching, and keyboard shortcuts remain out of scope for this cycle.

## Source Inputs

- User-selected visual direction: **A. Command Center / Raycast style**.
- User-selected scope: **B. Add Toast + confirmation dialogs**.
- Existing frontend design document: `docs/FRONTEND.md`.
- Existing frontend structure:
  - `frontend/src/App.tsx`
  - `frontend/src/components/Sidebar.tsx`
  - `frontend/src/components/Panel.tsx`
  - `frontend/src/pages/Dashboard.tsx`
  - `frontend/src/pages/Environments.tsx`
  - `frontend/src/pages/Migration.tsx`
  - `frontend/src/pages/Cleaner.tsx`
  - `frontend/src/pages/Versions.tsx`
  - `frontend/src/pages/Settings.tsx`
  - `frontend/tailwind.config.js`
  - `frontend/src/style.css`

## Design Principles

1. **Developer first**: prioritize speed, density, direct actions, and readable paths/versions.
2. **Dark command center**: use a dark layered surface system rather than flat black or decorative glass.
3. **Data-driven**: make disk usage, environment count, cache size, migration progress, and version state visually scannable.
4. **Immediate feedback**: async actions must show loading, result, toast, and error states where appropriate.
5. **Professional iconography**: replace UI emoji icons with consistent inline SVG icons. Emoji in documentation is treated as wireframe shorthand, not final visual language.
6. **Stable interaction**: avoid layout-shifting hover effects. Prefer color, border, subtle glow, and surface elevation changes.

## Visual System

### Palette

Keep the current DevMan dark foundation but refine it into explicit surface levels:

- App background: deep midnight slate (`devman-bg`, `devman-bg-deep`).
- Sidebar background: the deepest level, visually anchoring navigation.
- Card surface: elevated slate panel.
- Raised row/input surface: one level above card background.
- Border: low-contrast slate border with a stronger hover/selected variant.
- Primary accent: green (`devman-accent`) for run/healthy/primary action states.
- Info accent: cyan/blue (`devman-info`) for sizes, paths, and secondary highlights.
- Warning/danger: amber/red for disk pressure, destructive actions, and failed operations.

Avoid pure black, pure white, heavy glassmorphism, and large neon areas.

### Typography

- UI font: keep the current sans stack, preferably IBM Plex Sans / Inter / system UI.
- Monospace font: keep JetBrains Mono / Consolas for paths, versions, logs, and database locations.
- Use tighter display hierarchy on dashboard metrics: large numeric values, small muted labels, and compact metadata.

### Motion

- Page/content reveal: subtle 150-200ms fade/slide where it does not distract.
- Progress bars: 250-300ms width transitions.
- Buttons/cards: 150-200ms color/border/shadow transitions.
- Loading: skeleton or spinner, never a frozen empty view.
- Respect `prefers-reduced-motion` in global CSS.

## App Shell

### `App.tsx`

The shell remains a Wails-friendly local-state page switcher. It should become a more polished desktop frame:

- Deep background with subtle radial or gradient atmosphere.
- Fixed sidebar and scrollable content region.
- Main content gets consistent max width, page padding, and vertical rhythm.
- Toast host and confirm dialog host live near the app root.

### `Sidebar.tsx`

Sidebar remains 240px for the default desktop window.

Required changes:

- Replace emoji navigation icons with consistent SVG icons.
- Keep nav data configuration-driven inside the component or a small local config file.
- Active item: left accent rail + raised background + green accent text.
- Hover item: stable background/border change, no scale shift.
- Footer: version and status chip, using muted text and mono version label.

Future responsive collapse is out of scope unless needed to prevent obvious overflow.

## Shared Components

The current `Panel` should evolve into a small set of reusable primitives under `frontend/src/components/` or `frontend/src/components/ui/`.

Minimum primitives for this cycle:

- `Panel` or `SurfaceCard`: base card/surface with variants for default, raised, interactive, selected, danger.
- `Button`: primary, secondary, danger, ghost; consistent disabled/focus states.
- `StatusBadge`: healthy, warning, danger, info, active, default.
- `ProgressBar`: consistent track/fill behavior for disk and ranked usage.
- `PageHeader`: title, description, optional actions.
- `EmptyState`: icon, title, description, optional action.
- `Toast`: transient success/error/info feedback.
- `ConfirmDialog`: blocking confirmation for dangerous or irreversible operations.

Implementation should stay lightweight and avoid adding a component library unless explicitly chosen during planning.

## Page Designs

### Dashboard

Dashboard sets the design language for the app.

Required behavior remains:

- Refresh calls `ScanAll()` and `GetDiskInfo()`.
- C drive status uses thresholds: >90 danger, >70 warning, otherwise healthy.
- Environment usage ranking uses proportional bars with accent-to-info gradient.

Visual redesign:

- Use a command-center page header with a compact refresh button.
- Convert metric panels into strong stat cards with label, value, status badge, and supporting metadata.
- Make the ranking panel look like a dense system monitor list.
- Use skeleton/loading state while scan is running.
- Show a toast when refresh succeeds or fails.

### Environments

Modernize the environment catalog around scan results and direct actions.

Required behavior remains:

- Load environments from `GetEnvs()`.
- Local search filters by name/key.

Visual redesign:

- Header uses "Environment Control" tone, not generic page chrome.
- Replace environment emoji with SVG/category icon treatment where possible; if backend-provided `Env.Icon` remains emoji-like, wrap it as data content rather than primary UI chrome.
- Cards show name, category, description, source/website, and action buttons.
- Add expandable details only if supported by existing data without new backend work; otherwise keep action-ready cards.
- Empty state uses shared `EmptyState`.

### Migration

Migration is the most important workflow to make safe and readable.

Required behavior remains:

- Four-step flow: choose environment, choose target, preview, execute/result.
- Uses `GetEnvs()`, `GetEnvSummary()`, `GetDiskInfo()`, and `Migrate()`.

Visual redesign:

- Replace the current step indicator with a compact command-center stepper.
- Step 1: selectable rows/cards with size, path, and selected state.
- Step 2: selected environment summary + target path input + disk preview.
- Step 3: high-contrast confirmation summary and warning block.
- Step 4: progress/result state with log surface.
- Add `ConfirmDialog` before starting migration.
- Add success/error toast after migration completes/fails.
- Keep wording that migration must not be interrupted.

### Cleaner

Required behavior remains:

- Analyze via `AnalyzeCleanable()`.
- Select all / individual items.
- Clean via `CleanItems()`.

Visual redesign:

- Summary panel emphasizes reclaimable space and selected size.
- List rows use checkbox + source/name/path/description/size hierarchy.
- Add `ConfirmDialog` before cleaning selected items.
- Show toast after analyze success/failure and clean success/failure.
- Preserve disabled state when no items are selected.

### Versions

Required behavior remains:

- Load `GetEnvs()` and `GetEnvSummary()`.
- Display instances, default state, active state, source, version, path.

Visual redesign:

- Group by environment with dense version rows.
- Use `StatusBadge` for default and active.
- Use monospace styling for versions/paths.
- Empty state uses shared `EmptyState`.

### Settings

Required behavior remains:

- Existing local toggles remain local UI state unless backed by future settings persistence.
- Database/path and about information remain display-only.

Visual redesign:

- Use settings sections with `SettingRow` style layout.
- Toggles get visible focus and stable hover states.
- Destructive or reset-like actions should use confirm dialog if present.
- Data export can show an informational toast if the action is currently placeholder-only, or remain visually disabled if not implemented.

## Toast Notifications

Toast scope for this cycle:

- Success and error feedback for scan/analyze/clean/migrate actions.
- Optional info toast for placeholder actions only if the action remains clickable.
- Toasts should not replace inline result state for migration and cleaner; they supplement it.

Toast behavior:

- Stack at top-right or bottom-right inside the Wails window.
- Auto-dismiss after a short delay.
- Manual close button.
- Variants: success, error, info, warning.
- No persistence.

## Confirmation Dialogs

Confirm dialog scope for this cycle:

- Required before starting migration.
- Required before cleaning selected items.
- Optional before reset/destructive settings actions if such actions are implemented.

Confirm dialog behavior:

- Modal overlay within app window.
- Title, explanatory body, cancel button, confirm button.
- Danger variant for destructive operations.
- Keyboard accessible focus management should be handled as much as practical without introducing large dependencies.

## Accessibility Requirements

- All buttons and inputs must have visible focus states.
- Do not remove outlines without a replacement ring.
- Clickable rows/cards must use `cursor-pointer`.
- Color cannot be the only status indicator; text labels remain present.
- Inputs must have labels or accessible labels.
- Icon-only controls require accessible names.

## Non-Goals

The following are explicitly out of scope for this redesign cycle:

- Global search / command palette.
- Theme switching or light mode.
- Keyboard shortcuts such as `Ctrl+R` and `Ctrl+K`.
- Backend model/API changes.
- Replacing the Wails page switching approach with a router.
- Adding heavy chart/table/component libraries unless a later implementation plan justifies them.

## Risks

- `Migration.tsx` combines state, data fetching, step rendering, and async mutation; redesign must avoid changing behavior while improving structure.
- Existing pages duplicate button/card styles; extracting primitives should be staged carefully.
- Some current icons come from backend data (`Env.Icon`). Treat backend-provided icons as content, but use SVG for core app/navigation/action icons.
- `frontend/src/App.css` appears to be stale template CSS; removal should be verified against imports before deleting.

## Acceptance Criteria

- The six existing pages retain their current core behavior.
- UI direction matches Command Center / Raycast style: dark layered, dense, professional, non-neon.
- Sidebar and page headers use consistent SVG iconography.
- Buttons, cards, badges, progress bars, empty states, toast, and confirm dialog use shared styling.
- Migration and cleaning actions require confirmation.
- Async scan/analyze/clean/migrate flows show immediate loading and toast feedback.
- No global search, theme switching, or shortcut system is added.
- `frontend/src/bindings/go/` and `frontend/wailsjs/` remain untouched.
