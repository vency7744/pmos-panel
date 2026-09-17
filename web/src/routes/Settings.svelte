<script lang="ts">
  let settings = $state({});

  async function loadSettings() {
    const res = await fetch('/api/me');
    if (res.ok) settings = await res.json();
  }
</script>

<div class="settings-page">
  <h2>Settings</h2>
  <div class="settings-card">
    <p>Settings configuration will be available in a future update.</p>
  </div>
</div>

<style>
  .settings-page { padding: 0; }
  h2 { margin: 0 0 1.5rem; font-size: 1rem; color: var(--text-h); }
  .settings-card {
    background: var(--card); border: 1px solid var(--border); border-radius: 8px; padding: 1.5rem;
    color: var(--text); font-size: 0.875rem;
  }
</style>
