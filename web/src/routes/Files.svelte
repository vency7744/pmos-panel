<script lang="ts">
  import { onMount } from 'svelte';

  interface FileEntry {
    name: string;
    path: string;
    size: number;
    mode: string;
    is_dir: boolean;
    mod_time: string;
  }

  let files = $state<FileEntry[]>([]);
  let currentPath = $state('/');
  let loading = $state(false);
  let newFolderName = $state('');
  let showNewFolder = $state(false);
  let error = $state('');
  let uploading = $state(false);
  let fileInput: HTMLInputElement;

  onMount(() => loadFiles('/'));

  async function loadFiles(path: string) {
    loading = true;
    error = '';
    try {
      const res = await fetch(`/api/files?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        files = await res.json();
        currentPath = path;
      } else {
        const data = await res.json();
        error = data.error || 'Failed to load';
      }
    } catch {
      error = 'Connection error';
    }
    loading = false;
  }

  function navigate(entry: FileEntry) {
    if (entry.is_dir) {
      loadFiles(entry.path);
    } else {
      window.open(`/api/files/read?path=${encodeURIComponent(entry.path)}`, '_blank');
    }
  }

  function goUp() {
    const parts = currentPath.split('/').filter(Boolean);
    parts.pop();
    loadFiles('/' + parts.join('/'));
  }

  function download(entry: FileEntry) {
    const a = document.createElement('a');
    a.href = `/api/files/download?path=${encodeURIComponent(entry.path)}`;
    a.download = entry.name;
    a.click();
  }

  async function deleteFile(entry: FileEntry) {
    if (!confirm(`Delete ${entry.name}?`)) return;
    await fetch('/api/files/delete', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: entry.path }),
    });
    loadFiles(currentPath);
  }

  async function createFolder() {
    if (!newFolderName) return;
    const path = currentPath === '/' ? `/${newFolderName}` : `${currentPath}/${newFolderName}`;
    await fetch('/api/files/mkdir', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    });
    newFolderName = '';
    showNewFolder = false;
    loadFiles(currentPath);
  }

  async function uploadFiles(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;
    uploading = true;
    error = '';
    const formData = new FormData();
    formData.append('path', currentPath);
    for (const file of input.files) {
      formData.append('file', file);
    }
    try {
      const res = await fetch('/api/files/upload', { method: 'POST', body: formData });
      if (!res.ok) {
        const data = await res.json();
        error = data.error || 'Upload failed';
      }
    } catch {
      error = 'Upload failed';
    }
    input.value = '';
    uploading = false;
    loadFiles(currentPath);
  }

  function formatSize(bytes: number): string {
    if (bytes === 0) return '-';
    const units = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
  }

  function icon(entry: FileEntry): string {
    if (entry.is_dir) return '📁';
    const ext = entry.name.split('.').pop()?.toLowerCase();
    if (['js', 'ts', 'go', 'py', 'rs'].includes(ext || '')) return '📄';
    if (['jpg', 'png', 'gif', 'svg'].includes(ext || '')) return '🖼';
    if (['zip', 'tar', 'gz'].includes(ext || '')) return '📦';
    if (['md', 'txt', 'log'].includes(ext || '')) return '📝';
    return '📄';
  }
</script>

<div class="files-page">
  <div class="toolbar">
    <div class="breadcrumb">
      <button class="nav-btn" onclick={() => loadFiles('/')} disabled={currentPath === '/'}>/</button>
      {#each currentPath.split('/').filter(Boolean) as part, i}
        <span class="sep">/</span>
        <button
          class="nav-btn"
          onclick={() => loadFiles('/' + currentPath.split('/').filter(Boolean).slice(0, i + 1).join('/'))}
        >{part}</button>
      {/each}
    </div>
    <div class="toolbar-actions">
      <button onclick={() => fileInput.click()} disabled={uploading}>{uploading ? 'Uploading...' : 'Upload'}</button>
      <input type="file" multiple bind:this={fileInput} onchange={uploadFiles} style="display:none" />
      <button onclick={() => showNewFolder = !showNewFolder}>New Folder</button>
      <button onclick={() => loadFiles(currentPath)} disabled={loading}>Refresh</button>
    </div>
  </div>

  {#if showNewFolder}
    <div class="new-folder">
      <input type="text" placeholder="Folder name" bind:value={newFolderName} />
      <button onclick={createFolder}>Create</button>
      <button onclick={() => { showNewFolder = false; newFolderName = ''; }}>Cancel</button>
    </div>
  {/if}

  {#if error}
    <div class="error">{error}</div>
  {/if}

  <div class="file-list">
    {#if currentPath !== '/'}
      <div class="file-row dir" role="button" tabindex="0" onclick={goUp} onkeydown={(e) => e.key === 'Enter' && goUp()}>
        <span class="file-icon">⬆</span>
        <span class="file-name">..</span>
        <span class="file-size">-</span>
        <span class="file-date">-</span>
      </div>
    {/if}

    {#each files as file}
      <div class="file-row" class:dir={file.is_dir} role="button" tabindex="0" onclick={() => navigate(file)} onkeydown={(e) => e.key === 'Enter' && navigate(file)}>
        <span class="file-icon">{icon(file)}</span>
        <span class="file-name">{file.name}</span>
        <span class="file-size">{file.is_dir ? '-' : formatSize(file.size)}</span>
        <span class="file-date">{file.mod_time}</span>
        <div class="file-actions" role="none" onclick={(e) => e.stopPropagation()}>
          {#if !file.is_dir}
            <button class="action-btn" onclick={() => download(file)}>↓</button>
          {/if}
          <button class="action-btn delete" onclick={() => deleteFile(file)}>×</button>
        </div>
      </div>
    {/each}

    {#if files.length === 0 && !loading}
      <div class="empty">Empty directory</div>
    {/if}
  </div>
</div>

<style>
  .files-page { padding: 0; }
  .toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; flex-wrap: wrap; gap: 0.5rem; }
  .breadcrumb { display: flex; align-items: center; flex-wrap: wrap; }
  .sep { color: var(--text); margin: 0 0.125rem; font-size: 0.875rem; }
  .nav-btn {
    padding: 0.25rem 0.375rem; border: none; border-radius: 4px; background: transparent;
    color: var(--accent); font-size: 0.8125rem; cursor: pointer; font-family: var(--mono);
  }
  .nav-btn:hover { background: var(--hover); }
  .nav-btn:disabled { color: var(--text); cursor: default; }
  .toolbar-actions { display: flex; gap: 0.5rem; }
  button {
    padding: 0.375rem 0.625rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--card); color: var(--text); font-size: 0.8125rem; cursor: pointer;
  }
  button:hover { background: var(--hover); }
  .new-folder { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
  .new-folder input {
    flex: 1; padding: 0.375rem 0.625rem; border: 1px solid var(--border); border-radius: 6px;
    background: var(--input-bg); color: var(--text-h); font-size: 0.8125rem; outline: none;
  }
  .error {
    padding: 0.5rem 0.75rem; background: var(--error-bg); border: 1px solid var(--error);
    border-radius: 6px; color: var(--error); font-size: 0.8125rem; margin-bottom: 1rem;
  }
  .file-list { border: 1px solid var(--border); border-radius: 8px; overflow: hidden; }
  .file-row {
    display: flex; align-items: center; padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--border);
    cursor: pointer; gap: 0.75rem; font-size: 0.8125rem;
  }
  .file-row:last-child { border-bottom: none; }
  .file-row:hover { background: var(--hover); }
  .file-row.dir { color: var(--accent); }
  .file-icon { width: 1.25rem; text-align: center; flex-shrink: 0; }
  .file-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-h); }
  .file-size { width: 80px; text-align: right; font-family: var(--mono); color: var(--text); flex-shrink: 0; }
  .file-date { width: 140px; color: var(--text); flex-shrink: 0; }
  .file-actions { display: flex; gap: 0.25rem; }
  .action-btn {
    width: 1.5rem; height: 1.5rem; padding: 0; border: 1px solid var(--border); border-radius: 4px;
    background: transparent; color: var(--text); cursor: pointer; display: flex; align-items: center;
    justify-content: center; font-size: 0.75rem;
  }
  .action-btn:hover { background: var(--hover); }
  .action-btn.delete { color: var(--error); }
  .empty { padding: 2rem; text-align: center; color: var(--text); }
</style>
