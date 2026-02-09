<script>
  import { onMount } from 'svelte';
  import { GetLibraryItems, RemoveFromLibrary } from '../../wailsjs/go/main/App';
  import { createEventDispatcher } from 'svelte';

  const dispatch = createEventDispatcher();

  let libraryItems = [];
  let loading = true;
  let groupedItems = {};
  let expandedDramas = {}; // Track which drama groups are expanded

  onMount(async () => {
    await loadLibrary();
  });

  async function loadLibrary() {
    loading = true;
    try {
      libraryItems = await GetLibraryItems() || [];
      groupItems();
    } catch (e) {
      console.error('Failed to load library:', e);
      libraryItems = [];
    } finally {
      loading = false;
    }
  }

  function groupItems() {
    groupedItems = {};
    for (const item of libraryItems) {
      const key = item.dramaId || item.dramaName;
      if (!groupedItems[key]) {
        groupedItems[key] = {
          dramaName: item.dramaName,
          dramaCover: item.dramaCover,
          platform: item.platform,
          episodes: []
        };
      }
      groupedItems[key].episodes.push(item);
    }
    
    // Sort episodes by index
    for (const key of Object.keys(groupedItems)) {
      groupedItems[key].episodes.sort((a, b) => a.episodeIndex - b.episodeIndex);
    }
  }

  function toggleDrama(dramaId) {
    expandedDramas[dramaId] = !expandedDramas[dramaId];
    expandedDramas = expandedDramas; // Trigger reactivity
  }

  function playOffline(item, dramaEpisodes) {
    dispatch('playOffline', { 
      item, 
      episodes: dramaEpisodes 
    });
  }

  async function removeItem(item, deleteFile = false) {
    if (confirm(`Remove "${item.dramaName} - ${item.episodeName}" from library${deleteFile ? ' and delete file' : ''}?`)) {
      try {
        await RemoveFromLibrary(item.id, deleteFile);
        await loadLibrary();
      } catch (e) {
        console.error('Failed to remove:', e);
      }
    }
  }

  function formatSize(bytes) {
    if (!bytes) return '';
    const mb = bytes / (1024 * 1024);
    if (mb < 1024) return `${mb.toFixed(1)} MB`;
    return `${(mb / 1024).toFixed(2)} GB`;
  }

  function formatDate(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleDateString();
  }
</script>

<div class="library-container">
  <div class="library-header">
    <h1>My Library</h1>
    <p class="subtitle">{libraryItems.length} videos downloaded</p>
  </div>

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Loading library...</p>
    </div>
  {:else if libraryItems.length === 0}
    <div class="empty-state">
      <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
        <polyline points="7,10 12,15 17,10"/>
        <line x1="12" y1="15" x2="12" y2="3"/>
      </svg>
      <h3>No Downloads Yet</h3>
      <p>Download videos to watch them offline</p>
    </div>
  {:else}
    <div class="drama-list">
      {#each Object.entries(groupedItems) as [dramaId, drama]}
        <div class="drama-group">
          <div class="drama-group-header" on:click={() => toggleDrama(dramaId)} role="button" tabindex="0" on:keypress={(e) => e.key === 'Enter' && toggleDrama(dramaId)}>
            <svg class="expand-icon" class:expanded={expandedDramas[dramaId]} width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="9,18 15,12 9,6"/>
            </svg>
            {#if drama.dramaCover}
              <img src={drama.dramaCover} alt={drama.dramaName} class="drama-thumbnail" />
            {:else}
              <div class="drama-thumbnail-placeholder">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <rect x="2" y="2" width="20" height="20" rx="2"/>
                  <path d="M10 8l6 4-6 4V8z"/>
                </svg>
              </div>
            {/if}
            <div class="drama-info">
              <h3>{drama.dramaName}</h3>
              <span class="episode-count">{drama.episodes.length} episodes</span>
            </div>
          </div>
          
          {#if expandedDramas[dramaId]}
          <div class="episode-list">
            {#each drama.episodes as item}
              <div class="episode-item">
                <div class="episode-left">
                  <span class="episode-name">{item.episodeName}</span>
                  <div class="episode-meta">
                    <span>{formatSize(item.fileSize)}</span>
                    {#if item.hasSubtitles}
                      <span class="badge">Subtitles</span>
                    {/if}
                    {#if item.watchProgress > 0}
                      <span class="progress-badge">{Math.round(item.watchProgress)}% watched</span>
                    {/if}
                  </div>
                </div>
                <div class="episode-actions">
                  <button class="btn-play" on:click={() => playOffline(item, drama.episodes)} title="Play">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                      <polygon points="5,3 19,12 5,21"/>
                    </svg>
                  </button>
                  <button class="btn-delete" on:click={() => removeItem(item, true)} title="Delete">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M3 6h18"/>
                      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/>
                      <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
                    </svg>
                  </button>
                </div>
              </div>
            {/each}
          </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .library-container {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .library-header {
    margin-bottom: 24px;
  }

  .library-header h1 {
    font-size: 28px;
    font-weight: 700;
  }

  .subtitle {
    color: var(--text-secondary);
    font-size: 14px;
    margin-top: 4px;
  }

  .loading-state,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 300px;
    gap: 16px;
    color: var(--text-secondary);
    text-align: center;
  }

  .empty-state h3 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--bg-tertiary);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .drama-list {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .drama-group {
    background: var(--bg-secondary);
    border-radius: 12px;
    overflow: hidden;
    border: 1px solid var(--border-color);
  }

  .drama-group-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px;
    background: var(--bg-tertiary);
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .drama-group-header:hover {
    background: var(--bg-hover);
  }

  .expand-icon {
    flex-shrink: 0;
    color: var(--text-secondary);
    transition: transform 0.2s ease;
  }

  .expand-icon.expanded {
    transform: rotate(90deg);
  }

  .drama-thumbnail {
    width: 60px;
    height: 80px;
    object-fit: cover;
    border-radius: 6px;
  }

  .drama-thumbnail-placeholder {
    width: 60px;
    height: 80px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-hover);
    border-radius: 6px;
    color: var(--text-secondary);
  }

  .drama-info h3 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 4px;
  }

  .episode-count {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .episode-list {
    padding: 8px;
  }

  .episode-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px;
    border-radius: 8px;
    transition: background 0.15s ease;
  }

  .episode-item:hover {
    background: var(--bg-hover);
  }

  .episode-left {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .episode-name {
    font-size: 14px;
    font-weight: 500;
  }

  .episode-meta {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--text-secondary);
  }

  .badge {
    padding: 2px 6px;
    background: #60597260;
    color: #fff;
    border-radius: 4px;
    font-size: 10px;
    font-weight: 500;
  }

  .progress-badge {
    padding: 2px 6px;
    background: var(--bg-hover);
    border-radius: 4px;
    font-size: 10px;
  }

  .episode-actions {
    display: flex;
    gap: 8px;
  }

  .btn-play {
    padding: 10px;
    background: var(--bg-secondary);
    color: #fff;
    border: 2px;
    border-color: #fff;
    border-radius: 8px;
    transition: all 0.15s ease;
  }

  .btn-play:hover {
    transform: scale(1.05);
  }

  .btn-delete {
    padding: 10px;
    color: var(--text-secondary);
    border-radius: 8px;
    transition: all 0.15s ease;
  }

  .btn-delete:hover {
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
  }
</style>
