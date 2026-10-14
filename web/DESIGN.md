# Design rules

How every home-tools UI looks and behaves. Decision #23 in `AGENTS.md`.
Tokens, base styles and shared pieces live in `@home-tools/ui`
(`web/packages/ui/src`: `tokens.css`, `app.css`, `components/`, `profiles/`);
every app imports `@home-tools/ui/app.css` once in `main.ts`.

## 1. Tokens only
- Components use semantic tokens: `var(--color-text)`, `var(--space-4)`.
  Never hex codes, never raw `px` for type or spacing.
- Need a value that doesn't exist? Add a token (and say why), don't inline it.
- Raw palette values appear **only** in `tokens.css`.

## 2. Color
- Light and dark mode come free: every color token is `light-dark(...)`.
  Check every screen in both before calling it done.
- **Accent = the profile's color** (`<html data-accent="blue">`). Use it for
  the primary action, selection, focus and links. One primary action per view.
- **Red only means danger:** errors and destructive actions. Never decoration.
- Never use color as the only signal. An error also gets text and/or an icon;
  a selected item also gets a check or weight change.

## 3. Type and space
- Font: Inter (variable, bundled via `@fontsource-variable/inter`, so it
  works offline). Neutral and precise: the product should feel like a
  tool, not a toy (decision #57). Weights from `--weight-*`: page titles
  `--weight-bold`, section titles and buttons `--weight-semibold`, labels
  `--weight-medium`. `--weight-extrabold` only for hero numbers (the weight
  you're lifting), never for titles.
- Sizes from the `--text-*` scale only; body is `--text-md`. Page `<h1>` =
  `--text-title` (24px on phones → 28px), `<h2>` = `--text-lg`, `<h3>` =
  `--text-md`: set once in `app.css`, so components don't restyle headings
  (a card's own title may drop to `--text-md`). Headings get
  `--tracking-tight`.
- Inputs are at least `--text-md` (iOS zooms the page on smaller inputs).
- Numbers that update (reps, kg, timers) use tabular figures (set in `app.css`).
- Spacing from the 4px `--space-*` scale. Related things sit closer together
  than unrelated things; that's most of what makes a layout feel organized.

## 4. Semantics (HTML first)
- Use the element that already means the thing: `<button>` for actions,
  `<a href>` for navigation, `<label>` for every input, `<fieldset>` +
  `<legend>` for groups (like the color picker), `<ul>` for lists.
- One `<h1>` per view; heading levels never skip (h1 → h2 → h3).
- Landmarks: `<header>`, `<nav>`, `<main>`, `<footer>`.
- ARIA only when HTML can't express it (e.g. `aria-live` for "Saved").
  No ARIA beats wrong ARIA.
- Never put a click handler on a `<div>`.

## 5. Accessibility (WCAG 2.2 AA)
- Contrast is handled by the tokens (checked: text ≥ 7:1, muted/accent text
  ≥ 4.5:1, borders/fills ≥ 3:1). Don't invent new color pairs without checking.
- Everything works by keyboard: Tab order follows reading order, Esc closes
  dialogs, focus moves into a dialog and back out when it closes.
- Focus is always visible (`:focus-visible` ring from `app.css`). Never
  `outline: none` without a replacement.
- Touch targets ≥ `--touch-target` (44px). This is a gym app on a phone.
- Images: meaningful ones get `alt`; decorative ones get `alt=""`.
- Motion uses `--duration-*` tokens, which drop to 0 with reduced motion.
- Form errors: shown next to the field as text, linked with `aria-describedby`.

## 6. States checklist
An interactive element isn't done until each state is designed:

| State | Question to answer |
|---|---|
| Rest | Does it look interactive? |
| Hover | Does it respond before I click? (pointer devices only: `@media (hover: hover)`) |
| Pressed | Does it react the instant I tap? (`:active`) |
| Focus | Is the keyboard ring visible and not clipped? |
| Disabled | Is it clear *why*? (Prefer explaining over disabling.) |
| Loading | Does it stop double-submits and show progress? |

And each **view** answers: empty (first use), loading, error, and full/long
content (long names, 50 items).

## 7. Shape and surfaces (decision #57)
- Corners: `--radius-md` (8px) for buttons, inputs and rows;
  `--radius-lg` (12px) for cards, dialogs and menus; `--radius-full` only
  for tags, toggles and avatars. **Buttons are never pills.**
- Surfaces: the page is `--color-bg`; cards, the header and inputs are
  `--color-surface` with a 1px `--color-border`; `--color-surface-sunken`
  for wells (the log), skeletons and row hover. No tinted "hero" cards:
  the accent marks actions and selection, not areas.
- Shadows only on things that float (menus, dialogs, toasts) and on a
  hovered link-card. Cards at rest are flat.
- Blobs are **avatars only** (a person's profile, the Discord persona).
  Tools have glyph icons on colored tiles (`ToolIcon`); days, states and
  decoration use plain shapes.
- Say less: a page or section gets at most one line of description, and
  only when the title doesn't say it. No greetings, no exclamation marks.

## 8. Shared building blocks (use these, don't restyle)
Every page in every tool is built from the same pieces, so the four apps
read as one product.

- **Shell:** `<AppShell tool="fitness" …>`: the tool switcher (top left:
  every tool's icon and name, plus "All tools"), the page links (tabs
  underlined in the header on wide screens, a bottom tab bar on phones),
  the profile menu, and the toasts.
- **Page header:** `<PageHeader title description back actions meta
  leading eyebrow>` (`components/page-header`) starts every page: the
  `<h1>`, an optional back link ("‹ Plans"), one line of description,
  actions on the right. Never a hand-rolled `<h1>` or `.back`.
- **Sections:** `<Section title description count actions>` for a titled
  part of a page; `<Card padding as href>` for a bordered group (a list,
  controls, a chart). Rows inside `<Card padding="none">` go in
  `<ul class="card-list">` (dividers come free).
- **Widths:** pages use `--content-narrow` (forms, articles) or
  `--content-medium` (settings, editors), or the full `--page-width`
  (dashboards). No other `max-width` on a page.
- **Forms:** `<Field label hint error optional>` wires the label, hint and
  error to the control (`aria-describedby`, `aria-invalid`). Inputs and
  selects are styled in `app.css`; components only set width.
- **Loading:** a skeleton in the shape of the content, never a blank page
  or "Loading…": `<PageSkeleton variant="list|grid|form">`, or
  `<SkeletonGroup label>` around `<Skeleton>`s for a custom shape. It waits
  150 ms before showing, so fast loads don't flash.
- **Empty:** `<EmptyState icon title description action>`.
- **Saved / failed in the background:** `toast('Plan saved')`,
  `toast.error(…)`, or with `{ action: { label: 'Undo', run } }`
  (`@home-tools/ui/toast`). Errors a person must fix in a form stay next
  to the field; failed page loads are `LoadError`.
- **Status:** `<Badge tone dot pulse>`: success = up/done, warning =
  in progress or needs a look, danger = failed, neutral = off.
- **Switches:** `<ToggleSwitch checked label busy onchange>`.
- **Menus:** `popover="auto"` + `class="menu"`, items `.menu-item`
  (`.menu-item-danger` for destructive), positioned with `placeUnder`
  (`@home-tools/ui/menu`).
- **Buttons:** `.btn` + one of `.btn-primary`, `.btn-secondary`,
  `.btn-quiet`, `.btn-danger`, `.btn-text-danger`, `.btn-icon` (icon only;
  needs `aria-label`); `.btn-sm` for dense rows. Same classes on an `<a>`
  that looks like a button. All states are in `app.css`; pressed = a
  slight shrink plus a color change; disabled filled buttons go gray.
- **Dialogs:** `<ConfirmDialog title message confirmLabel danger
  onconfirm>` + `.open()` for "are you sure?" (safe choice focused first).
  Other dialogs: `<dialog class="dialog">` + `showModal()`.
- **Failed loads:** `<LoadError what="your week" onretry={load} />`.
- **Destructive actions:** `ConfirmDialog` for actions that affect other
  people (stopping a server); an inline confirm in place of the buttons
  for a row's own delete ("Keep it" / red "Discard"), focus on the safe
  choice; or do it and offer Undo in a toast (archiving a plan).
- **Screen-reader-only text:** `.visually-hidden`.

## 9. Before calling a view done
1. `pnpm --dir web check` passes, with no a11y warnings from svelte-check.
2. Checked in light and dark, on a phone-sized window.
3. Used it with only the keyboard.
4. axe DevTools (browser extension): no issues.
5. Tried it with VoiceOver (Cmd+F5): every control announces a sensible name.
