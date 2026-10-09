<!--
  What people can type in Discord, and who may. Mirrors the bot's slash
  commands (tools/discord/commands.go).
-->
<section class="section" aria-labelledby="commands-title">
  <h2 id="commands-title">Commands</h2>
  <dl class="list">
    <div class="row">
      <dt><code>/sens</code></dt>
      <dd>Converts a mouse sensitivity between games (17 games, optional DPI). Anyone.</dd>
    </div>
    <div class="row">
      <dt><code>/restart</code></dt>
      <dd>Restarts a game server. Verified people only; not again within 2 minutes.</dd>
    </div>
    <div class="row">
      <dt><code>/whitelist</code></dt>
      <dd>
        Adds, removes or lists players. Verified people only. Minecraft takes a username; Valheim a SteamID64 or
        Steam profile link, and needs a restart to apply.
      </dd>
    </div>
  </dl>
</section>

<style>
  .section {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  h2 {
    font-size: var(--text-xl);
    font-weight: var(--weight-bold);
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
    margin: 0;
  }

  .row {
    display: grid;
    grid-template-columns: 6.5rem 1fr;
    gap: var(--space-3);
  }

  dt code {
    font-family: var(--font-mono);
    font-weight: var(--weight-semibold);
  }

  dd {
    margin: 0;
    color: var(--color-text-muted);
  }
</style>
