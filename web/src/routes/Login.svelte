<script lang="ts">
  import { login as apiLogin } from '../lib/api.js';
  import { isAuthenticated, username } from '../lib/stores.js';

  let loginUsername = $state('');
  let loginPassword = $state('');
  let error = $state('');
  let loading = $state(false);

  async function handleLogin() {
    error = '';
    loading = true;

    const result = await apiLogin(loginUsername, loginPassword);
    loading = false;

    if (result.status === 'ok') {
      username.set(loginUsername);
      isAuthenticated.login();
    } else {
      error = result.error || 'Login failed';
    }
  }
</script>

<div class="login-container">
  <div class="login-card">
    <div class="login-header">
      <h1>PMOS Panel</h1>
      <p>Sign in to your server</p>
    </div>

    <form onsubmit={(e) => { e.preventDefault(); handleLogin(); }}>
      {#if error}
        <div class="error">{error}</div>
      {/if}

      <div class="field">
        <label for="username">Username</label>
        <input
          id="username"
          type="text"
          bind:value={loginUsername}
          placeholder="Enter username"
          required
        />
      </div>

      <div class="field">
        <label for="password">Password</label>
        <input
          id="password"
          type="password"
          bind:value={loginPassword}
          placeholder="Enter password"
          required
        />
      </div>

      <button type="submit" disabled={loading}>
        {loading ? 'Signing in...' : 'Sign In'}
      </button>
    </form>
  </div>
</div>

<style>
  .login-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    padding: 1rem;
  }

  .login-card {
    width: 100%;
    max-width: 360px;
    background: var(--card);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 2rem;
  }

  .login-header {
    text-align: center;
    margin-bottom: 1.5rem;
  }

  .login-header h1 {
    font-size: 1.5rem;
    color: var(--text-h);
    margin: 0 0 0.5rem;
  }

  .login-header p {
    color: var(--text);
    font-size: 0.875rem;
    margin: 0;
  }

  .error {
    background: var(--error-bg);
    color: var(--error);
    padding: 0.75rem;
    border-radius: 6px;
    font-size: 0.875rem;
    margin-bottom: 1rem;
  }

  .field {
    margin-bottom: 1rem;
  }

  label {
    display: block;
    font-size: 0.875rem;
    color: var(--text);
    margin-bottom: 0.375rem;
  }

  input {
    width: 100%;
    padding: 0.625rem 0.75rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--input-bg);
    color: var(--text-h);
    font-size: 0.875rem;
    box-sizing: border-box;
    outline: none;
    transition: border-color 0.2s;
  }

  input:focus {
    border-color: var(--accent);
  }

  button {
    width: 100%;
    padding: 0.625rem;
    border: none;
    border-radius: 6px;
    background: var(--accent);
    color: white;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.2s;
    margin-top: 0.5rem;
  }

  button:hover:not(:disabled) {
    opacity: 0.9;
  }

  button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
