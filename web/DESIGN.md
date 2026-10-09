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
- Font: Figtree (friendly geometric; bundled, so it works offline).
  Weights only from `--weight-*`: only the five bundled weights (400–800)
  render for real; anything else is faked by the browser. Page titles use
  `--weight-extrabold`.
- Sizes from the `--text-*` scale only; body is `--text-md`.
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

## 7. Shared building blocks (use these, don't restyle)
- **Buttons:** `.btn` + one of `.btn-primary`, `.btn-secondary`,
  `.btn-quiet`, `.btn-danger`, `.btn-text-danger`, `.btn-icon` (icon only;
  needs `aria-label`). Same classes on an `<a>` that looks like a button.
  All states (hover, pressed, disabled) are in `@home-tools/ui/app.css`.
- **Dialogs:** `<dialog class="dialog">` + `showModal()`. Shared look and
  backdrop in `@home-tools/ui/app.css`; the component sets only its size
  and padding.
- **Failed loads:** `<LoadError what="your week" onretry={load} />`
  (`@home-tools/ui/components/load-error`).
- **Destructive actions:** an inline confirm in place of the buttons
  ("Keep it" / red "Discard"), with focus moved to the safe choice.
- **Screen-reader-only text:** `.visually-hidden`.

## 8. Before calling a view done
1. `pnpm --dir web check` passes, with no a11y warnings from svelte-check.
2. Checked in light and dark, on a phone-sized window.
3. Used it with only the keyboard.
4. axe DevTools (browser extension): no issues.
5. Tried it with VoiceOver (Cmd+F5): every control announces a sensible name.
