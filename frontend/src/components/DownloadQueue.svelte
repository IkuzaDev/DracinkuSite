<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';
  import { GetAllDownloads, CancelDownload } from '../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher();
  let downloads = [];
  let interval;

  onMount(() => {
    loadDownloads();
    interval = setInterval(loadDownloads, 1000);
  });

  onDestroy(() => {
    clearInterval(interval);
  });

  async function loadDownloads() {
    try {
      const allDownloads = await GetAllDownloads() || [];
      // Filter out completed and cancelled downloads as requested
      downloads = allDownloads.filter(d => d.status !== 'completed' && d.status !== 'cancelled');
    } catch (e) {
      console.error('Failed to load downloads:', e);
    }
  }

  async function cancelDownload(id) {
    try {
      await CancelDownload(id);
      await loadDownloads();
    } catch (e) {
      console.error('Failed to cancel:', e);
    }
  }

  function formatBytes(bytes) {
    if (!bytes) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function getStatusColor(status) {
    switch (status) {
      case 'completed': return 'var(--success)';
      case 'failed': return 'var(--error)';
      case 'cancelled': return 'var(--warning)';
      case 'converting': return 'var(--accent)';
      case 'installing_ffmpeg': return 'var(--warning)';
      default: return 'var(--accent)';
    }
  }
  
  function isInProgress(status) {
    return status === 'downloading' || status === 'converting' || status === 'pending' || status === 'installing_ffmpeg';
  }
</script>

<div class="download-panel">
  <div class="panel-header">
    <h2>Downloads</h2>
    <button class="btn-icon" on:click={() => dispatch('close')}>
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="18" y1="6" x2="6" y2="18"/>
        <line x1="6" y1="6" x2="18" y2="18"/>
      </svg>
    </button>
  </div>

  <div class="panel-content">
    {#if downloads.length === 0}
      <div class="empty-state">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
          <polyline points="7,10 12,15 17,10"/>
          <line x1="12" y1="15" x2="12" y2="3"/>
        </svg>
        <p>No downloads yet</p>
      </div>
    {:else}
      <div class="download-list">
        {#each downloads as download}
          <div class="download-item">
            <div class="download-info">
              <span class="download-name">{download.filename}</span>
              <div class="download-meta">
                <span class="download-status" style="color: {getStatusColor(download.status)}">
                  {download.status}
                </span>
                {#if isInProgress(download.status)}
                  <span class="download-size">
                    {formatBytes(download.downloadedBytes)} / {formatBytes(download.totalBytes)}
                  </span>
                {/if}
              </div>
              {#if isInProgress(download.status)}
                <div class="progress-bar">
                  <div class="progress-fill" style="width: {download.progress}%"></div>
                </div>
              {/if}
            </div>
            
            {#if isInProgress(download.status)}
              <button class="btn-icon danger" on:click={() => cancelDownload(download.id)}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="18" y1="6" x2="6" y2="18"/>
                  <line x1="6" y1="6" x2="18" y2="18"/>
                </svg>
              </button>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .download-panel {
    position: fixed;
    right: 0;
    top: 0;
    bottom: 0;
    width: 380px;
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    z-index: 100;
    box-shadow: -8px 0 24px rgba(0, 0, 0, 0.3);
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px;
    border-bottom: 1px solid var(--border-color);
  }

  .panel-header h2 {
    font-size: 18px;
    font-weight: 600;
  }

  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 200px;
    gap: 12px;
    color: var(--text-muted);
  }

  .download-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .download-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px;
    background: var(--bg-tertiary);
    border-radius: 8px;
  }

  .download-info {
    flex: 1;
    min-width: 0;
  }

  .download-name {
    font-size: 14px;
    font-weight: 500;
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 4px;
  }

  .download-meta {
    display: flex;
    gap: 12px;
    font-size: 12px;
    margin-bottom: 8px;
  }

  .download-status {
    text-transform: capitalize;
    font-weight: 500;
  }

  .download-size {
    color: var(--text-muted);
  }

  .progress-bar {
    height: 4px;
    background: var(--bg-primary);
    border-radius: 2px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: var(--accent);
    transition: width 0.2s ease;
  }

  .btn-icon.danger:hover {
    background: rgba(248, 113, 113, 0.2);
    color: var(--error);
  }
</style>
