<script lang="ts">
  // Placeholder: proves the UI can reach the Go API through the proxy.
  let status = $state('Connecting…')

  fetch('/api/exercises')
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      return res.json()
    })
    .then((list: unknown[]) => (status = `Connected: ${list.length} exercises`))
    .catch((err) => (status = `API unreachable (${err.message}). Is \`make run\` running?`))
</script>

<main>
  <h1>Fitness</h1>
  <p>{status}</p>
</main>
