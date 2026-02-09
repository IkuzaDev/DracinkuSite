<script>
  import { createEventDispatcher } from 'svelte';
  export let platforms = [];
  export let selectedPlatform = null;

  const dispatch = createEventDispatcher();
</script>

<aside class="sidebar">
  <div class="logo">    
    <span>Dracinku.site</span>
  </div>

  <nav class="nav-section">
    <h3 class="nav-title">Platforms</h3>
    <ul class="nav-list">
      {#each platforms as platform}
        <li>
          <button 
            class="nav-item"
            class:active={selectedPlatform?.id === platform.id}
            on:click={() => dispatch('select', platform)}
          >
            <span>{platform.name}</span>
          </button>
        </li>
      {/each}
    </ul>
  </nav>

  <div class="sidebar-footer">
    <button class="nav-item" on:click={() => dispatch('showLibrary')}>
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
      </svg>
      <span>My Library</span>
    </button>
    <button class="nav-item" on:click={() => dispatch('toggleDownloads')}>
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
        <polyline points="7,10 12,15 17,10"/>
        <line x1="12" y1="15" x2="12" y2="3"/>
      </svg>
      <span>Downloads</span>
    </button>
  </div>
</aside>

<style>
  .sidebar {
    width: 240px;
    height: 100%;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 20px;
    font-size: 18px;
    font-weight: 700;
    color: var(--text-primary);
    border-bottom: 1px solid var(--border-color);
  }

  .nav-section {
    flex: 1;
    padding: 16px 12px;
    overflow-y: auto;
  }

  .nav-title {
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    color: var(--text-muted);
    padding: 8px 12px;
    letter-spacing: 0.5px;
  }

  .nav-list {
    list-style: none;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    padding: 10px 12px;
    border-radius: 8px;
    font-size: 14px;
    color: var(--text-secondary);
    transition: all 0.15s ease;
    text-align: left;
  }

  .nav-item:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .nav-item.active {
    background: var(--bg-tertiary);
    color: var(--accent);
  }

  .sidebar-footer {
    padding: 12px;
    border-top: 1px solid var(--border-color);
  }
</style>
