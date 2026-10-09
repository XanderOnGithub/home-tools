<!--
  The /blob feature: its switch, and a maker to try it here. The preview
  is the server's own image (the same one /blob sends), with a short
  pause while typing so it doesn't redraw on every key. Colors are the
  profile palette, picked with blob swatches like "Add profile".
-->
<script lang="ts">
  import { blobPath } from '@home-tools/ui/profiles/blob'
  import type { ProfileColor } from '@home-tools/ui/profiles/types'
  import type { Config } from '@/features/bot'
  import { ToggleSwitch } from '@/components/toggle-switch'

  let { config, save }: { config: Config; save: (change: (c: Config) => void) => Promise<string> } = $props()

  const COLORS: ProfileColor[] = ['green', 'blue', 'orange', 'purple']
  const label = (c: string) => c[0].toUpperCase() + c.slice(1)

  let busy = $state(false)
  let error = $state('')
  let name = $state('Jim')
  let color = $state<ProfileColor>('blue')
  let shown = $state('Jim') // the name the preview shows (after the typing pause)

  let timer: ReturnType<typeof setTimeout> | undefined
  $effect(() => {
    const n = name.trim()
    clearTimeout(timer)
    timer = setTimeout(() => (shown = n), 300)
    return () => clearTimeout(timer)
  })

  const src = (animated: boolean) =>
    `/api/blob?${new URLSearchParams({ name: shown, color, ...(animated ? { animated: '1' } : {}) })}`

  async function toggle(on: boolean) {
    busy = true
    error = await save((c) => (c.features.blob.enabled = on))
    busy = false
  }
</script>

<section class="feature" aria-labelledby="blob-title">
  <div class="head">
    <div>
      <h2 id="blob-title">Blob maker</h2>
      <p class="muted">
        <code>/blob</code> makes anyone a blob of their own: a name and a color. Same name, same blob, every time. A
        still PNG (works as a profile picture), or a blinking GIF.
      </p>
    </div>
    <ToggleSwitch checked={config.features.blob.enabled} label="Blob maker" {busy} onchange={toggle} />
  </div>
  {#if error}<p class="error" role="alert">{error}</p>{/if}

  <div class="maker">
    <div class="preview">
      {#if shown}
        <img src={src(false)} alt="{shown}, a {color} blob" width="160" height="160" />
      {/if}
    </div>
    <div class="controls">
      <div class="field">
        <label for="blob-name">Name</label>
        <input id="blob-name" bind:value={name} maxlength="32" autocomplete="off" />
      </div>
      <fieldset class="field">
        <legend>Color</legend>
        <div class="swatches">
          {#each COLORS as c (c)}
            <label class="swatch" data-accent={c}>
              <input class="visually-hidden" type="radio" name="blob-color" value={c} bind:group={color} />
              <svg viewBox="0 0 100 100" aria-hidden="true"><path d={blobPath(`swatch-${c}`)} /></svg>
              <span class="visually-hidden">{label(c)}</span>
            </label>
          {/each}
        </div>
      </fieldset>
      {#if shown}
        <div class="downloads">
          <a class="btn btn-secondary" href={src(false)} download="{shown}.png">Save PNG</a>
          <a class="btn btn-quiet" href={src(true)} download="{shown}.gif">Save GIF</a>
        </div>
      {/if}
    </div>
  </div>
</section>

<style>
  .feature {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
    padding: var(--space-4);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-lg);
    background: var(--color-surface);
  }

  .head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
  }

  .head > div {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  p {
    margin: 0;
  }

  .muted {
    color: var(--color-text-muted);
  }

  .error {
    color: var(--color-danger-text);
  }

  code {
    font-family: var(--font-mono);
    font-size: 0.9em;
  }

  .maker {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-5);
    padding-top: var(--space-4);
    border-top: 1px solid var(--color-border);
  }

  .preview {
    display: grid;
    flex: none;
    place-items: center;
    width: 10rem;
    height: 10rem;
    border-radius: var(--radius-full);
    background: var(--color-bg);
  }

  .preview img {
    width: 100%;
    height: 100%;
  }

  .controls {
    display: flex;
    flex: 1 1 14rem;
    flex-direction: column;
    gap: var(--space-4);
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    margin: 0;
    padding: 0;
    border: none;
  }

  label,
  legend {
    padding: 0;
    font-size: var(--text-sm);
    font-weight: var(--weight-semibold);
  }

  input {
    min-height: var(--touch-target);
    padding: 0 var(--space-3);
    border: 1px solid var(--color-border-strong);
    border-radius: var(--radius-md);
    background: var(--color-bg);
  }

  /* Same swatches as "Add profile" (ProfileDialog). */
  .swatches {
    display: flex;
    gap: var(--space-3);
    margin-top: var(--space-1);
  }

  .swatch {
    display: grid;
    place-items: center;
    width: var(--touch-target);
    height: var(--touch-target);
    padding: var(--space-1);
    border-radius: var(--radius-full);
    cursor: pointer;
  }

  .swatch svg {
    width: 100%;
    height: 100%;
    transition: scale var(--duration-fast) var(--ease-out);
  }

  .swatch path {
    fill: var(--color-accent);
  }

  .swatch:has(input:checked) {
    box-shadow: 0 0 0 2px var(--color-surface), 0 0 0 4px var(--color-text);
  }

  .swatch:has(input:focus-visible) {
    outline: var(--focus-ring);
    outline-offset: 4px;
  }

  @media (hover: hover) {
    .swatch:hover svg {
      scale: 1.1;
    }
  }

  .swatch:active svg {
    scale: 0.95;
  }

  .downloads {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
</style>
