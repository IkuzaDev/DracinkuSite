<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import heic2any from 'heic2any';
  import { SearchSeries, GetSeriesDetail, GetTabList, GetTabData, GetTabFeed } from '../../wailsjs/go/main/App';

  export let platform = '';
  export let searchKeyword = '';
  export let language = 'id';
  const convertedCache = new Map()

  const dispatch = createEventDispatcher();
  let dramas = [];
  let loading = false;
  let error = '';
  let pageInfo = null;
  let loadingMore = false;
  let convertedImages = {};
  
  // Filter state
  let tabs = [];
  let showFilterPopup = false;
  let selectedTab = null;
  let groupedTabs = {};

  $: if (platform && searchKeyword) {
    loadSearchResults();
  }

  $: if (platform && !searchKeyword) {
    loadInitialData();
  }

  $: if (dramas.length) {
    dramas.forEach(async (drama) => {
      if (drama.cover && isHeicUrl(drama.cover)) {
        if (!convertedImages[drama.cover]) {
          convertedImages[drama.cover] = await convertHeicToJpeg(drama.cover);
        }
      }
    });
  }


  function clearHeicCache() {
    // Revoke all blob URLs to free memory
    for (const blobUrl of convertedCache.values()) {
        if (blobUrl.startsWith('blob:')) {
            URL.revokeObjectURL(blobUrl)
        }
    }
    convertedCache.clear()
  }

  function isHeicUrl(url) {
    if (!url) {
      clearHeicCache();
      return false
    }
    return url.toLowerCase().includes('.heic')
  }

  async function convertHeicToJpeg(heicUrl) {
    // Check cache first
    if (convertedCache.has(heicUrl)) {
        return convertedCache.get(heicUrl)
    }

    try {
        // Fetch the HEIC image
        const response = await fetch(heicUrl)
        if (!response.ok) {
            throw new Error(`Failed to fetch image: ${response.status}`)
        }

        const heicBlob = await response.blob()

        // Check if it's actually HEIC
        if (!heicBlob.type.includes('heic') && !heicUrl.toLowerCase().includes('.heic')) {
            // Not HEIC, return original URL
            return heicUrl
        }

        // Convert to JPEG
        const jpegBlob = await heic2any({
            blob: heicBlob,
            toType: 'image/jpeg',
            quality: 0.8
        })

        // Create blob URL
        const resultBlob = Array.isArray(jpegBlob) ? jpegBlob[0] : jpegBlob
        const blobUrl = URL.createObjectURL(resultBlob)

        // Cache the result
        convertedCache.set(heicUrl, blobUrl)

        return blobUrl
    } catch (error) {
        console.error('[HEIC Convert] Failed:', error.message)
        // Return original URL as fallback
        return heicUrl
    }
  }

  // Group tabs by type
  function groupTabsByType(tabList) {
    const groups = {};
    for (const tab of tabList) {
      const type = tab.type || 'other';
      if (!groups[type]) {
        groups[type] = [];
      }
      groups[type].push(tab);
    }
    return groups;
  }

  async function loadSearchResults() {
    if (!searchKeyword.trim()) return;
    loading = true;
    error = '';
    pageInfo = null;
    try {
      const result = await SearchSeries(platform, searchKeyword, 1, 20, language);
      dramas = result?.book || [];
    } catch (e) {
      error = e.message || 'Failed to search';
      dramas = [];
    } finally {
      loading = false;
    }
  }

  async function loadInitialData() {
    loading = true;
    error = '';
    dramas = [];
    pageInfo = null;
    try {
      const tabList = await GetTabList(platform, language);
      tabs = tabList || [];
      groupedTabs = groupTabsByType(tabs);
      
      if (tabs.length > 0) {
        // Default: select first tab from first group
        const tab = tabs.length > 1 ? tabs[1] : tabs[0];
        await loadTabData(tab);
      }
    } catch (e) {
      console.error('[DramaList] Load failed:', e);
      error = e.message || e;
      dramas = [];
    } finally {
      loading = false;
    }
  }

  async function loadTabData(tab) {
    if (!tab) return;
    
    loading = true;
    error = '';
    dramas = [];
    pageInfo = null;
    selectedTab = tab;
    showFilterPopup = false;
    
    try {
      console.log('[DramaList] Loading tab:', tab);
      const result = await GetTabData(platform, tab.type, tab.tab_key, tab.position_index || 0, language);
      
      if (result?.book?.list) {
        dramas = result.book.list;
      } else if (result?.book?.recommend) {
        dramas = result.book.recommend;
      }
      
      if (result?.page_info) {
        pageInfo = result.page_info;
      }
    } catch (e) {
      console.error('[DramaList] Tab load failed:', e);
      error = e.message || e;
      dramas = [];
    } finally {
      loading = false;
    }
  }

  async function loadMore() {
    if (loading || loadingMore || !pageInfo || !pageInfo.has_more) return;
    
    loadingMore = true;
    try {
      const result = await GetTabFeed(platform, pageInfo, language);
      
      if (result?.book && result.book.length > 0) {
        dramas = [...dramas, ...result.book];
      }
      
      if (result?.page_info) {
        pageInfo = result.page_info;
      } else {
        pageInfo.has_more = false;
      }
    } catch (e) {
      console.error('[DramaList] Load more failed:', e);
    } finally {
      loadingMore = false;
    }
  }

  function handleScroll(e) {
    const { scrollTop, scrollHeight, clientHeight } = e.target;
    if (scrollHeight - scrollTop <= clientHeight + 100) {
      loadMore();
    }
  }

  function selectDrama(drama) {
    dispatch('select', drama);
  }

  function formatPlayCount(count) {
    if (!count) return '';
    const n = parseInt(count);
    if (isNaN(n)) return count;
    if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
    if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
    return n.toString();
  }

  function getTypeLabel(type) {
    const labels = {
      'region': 'Region',
      'audio': 'Audio/Language',
      'tag': 'Genre/Tag',
      'orderMode': 'Sort Order',
      'category': 'Category',
      'type': 'Type'
    };
    return labels[type] || type.charAt(0).toUpperCase() + type.slice(1);
  }

  function closePopup(e) {
    if (e.target.classList.contains('popup-overlay')) {
      showFilterPopup = false;
    }
  }
</script>

<div class="drama-list-wrapper">
  <!-- Header with filter button -->
  <div class="list-header">
    <button class="filter-btn" on:click={() => showFilterPopup = true}>
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <polygon points="22,3 2,3 10,12.46 10,19 14,21 14,12.46"/>
      </svg>
      <span>Filter</span>
      {#if selectedTab}
        <span class="filter-badge">{selectedTab.name}</span>
      {/if}
    </button>
  </div>

  <!-- Main grid container -->
  <div class="drama-grid-container" on:scroll={handleScroll}>
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>Loading...</p>
      </div>
    {:else if error}
      <div class="error-state">
        <p>{error}</p>
        <button class="btn btn-secondary" on:click={loadInitialData}>Retry</button>
      </div>
    {:else if dramas.length === 0}
      <div class="empty-state">
        <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/>
          <line x1="7" y1="2" x2="7" y2="22"/>
          <line x1="17" y1="2" x2="17" y2="22"/>
          <line x1="2" y1="12" x2="22" y2="12"/>
        </svg>
        <h3>No dramas found</h3>
        <p>Try selecting a different filter</p>
      </div>
    {:else}
      <div class="drama-grid">
        {#each dramas as drama}
          <button class="drama-card" on:click={() => selectDrama(drama)}>
            <div class="drama-cover">
              {#if drama.cover}
                <img src={convertedImages[drama.cover] || drama.cover} alt={drama.name} loading="lazy" />
              {:else}
                <div class="cover-placeholder">
                  <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <rect x="2" y="2" width="20" height="20" rx="2"/>
                    <circle cx="8.5" cy="8.5" r="1.5"/>
                    <polyline points="21,15 16,10 5,21"/>
                  </svg>
                </div>
              {/if}
              {#if drama.chapterCount}
                <span class="episode-badge">{drama.chapterCount} EP</span>
              {/if}
            </div>
            <div class="drama-info">
              <h4 class="drama-title">{drama.name}</h4>
              {#if drama.playCount}
                <span class="drama-views">{formatPlayCount(drama.playCount)} views</span>
              {/if}
            </div>
          </button>
        {/each}
      </div>
      {#if loadingMore}
        <div class="loading-more">
          <div class="spinner-small"></div>
          <span>Loading more...</span>
        </div>
      {/if}
    {/if}
  </div>
</div>

<!-- Filter Popup -->
{#if showFilterPopup}
  <div class="popup-overlay" on:click={closePopup}>
    <div class="popup-content">
      <div class="popup-header">
        <h3>Filter by Category</h3>
        <button class="close-btn" on:click={() => showFilterPopup = false}>
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="popup-body">
        {#each Object.entries(groupedTabs) as [type, tabItems]}
          <div class="filter-group">
            <h4 class="group-title">{getTypeLabel(type)}</h4>
            <div class="filter-chips">
              {#each tabItems as tab}
                <button 
                  class="filter-chip" 
                  class:active={selectedTab && selectedTab.tab_key === tab.tab_key && selectedTab.type === tab.type}
                  on:click={() => loadTabData(tab)}
                >
                  {tab.name}
                </button>
              {/each}
            </div>
          </div>
        {/each}
      </div>
    </div>
  </div>
{/if}

<style>
  .drama-list-wrapper {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .list-header {
    display: flex;
    align-items: center;
    padding: 16px 24px;
    border-bottom: 1px solid var(--border);
  }

  .filter-btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: var(--bg-secondary);
    border-radius: 8px;
    font-size: 14px;
    transition: background 0.15s ease;
  }

  .filter-btn:hover {
    background: var(--bg-tertiary);
  }

  .filter-badge {
    padding: 2px 8px;
    background: var(--accent);
    color: var(--bg-primary);
    border-radius: 4px;
    font-size: 12px;
    font-weight: 600;
  }

  .drama-grid-container {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .drama-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 20px;
  }

  .drama-card {
    background: var(--bg-secondary);
    border-radius: 12px;
    overflow: hidden;
    transition: all 0.2s ease;
    text-align: left;
  }

  .drama-card:hover {
    transform: translateY(-4px);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  }

  .drama-cover {
    position: relative;
    aspect-ratio: 3/4;
    background: var(--bg-tertiary);
    overflow: hidden;
  }

  .drama-cover img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .cover-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
  }

  .episode-badge {
    position: absolute;
    top: 8px;
    right: 8px;
    padding: 4px 8px;
    background: rgba(0, 0, 0, 0.7);
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
  }

  .drama-info {
    padding: 12px;
  }

  .drama-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    margin-bottom: 4px;
  }

  .drama-views {
    font-size: 12px;
    color: var(--text-muted);
  }

  .loading-state,
  .error-state,
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 300px;
    gap: 16px;
    color: var(--text-secondary);
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

  .empty-state svg {
    color: var(--text-muted);
  }

  .empty-state h3 {
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .loading-more {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 16px;
    color: var(--text-muted);
    font-size: 12px;
  }

  .spinner-small {
    width: 16px;
    height: 16px;
    border: 2px solid var(--bg-tertiary);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  /* Popup Styles */
  .popup-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
  }

  .popup-content {
    background: var(--bg-primary);
    border-radius: 16px;
    width: 90%;
    max-width: 600px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.5);
    border: 1px solid var(--border);
  }

  .popup-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid var(--border);
  }

  .popup-header h3 {
    font-size: 18px;
    font-weight: 600;
  }

  .close-btn {
    padding: 4px;
    border-radius: 6px;
    transition: background 0.15s ease;
  }

  .close-btn:hover {
    background: var(--bg-tertiary);
  }

  .popup-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .filter-group {
    margin-bottom: 24px;
  }

  .filter-group:last-child {
    margin-bottom: 0;
  }

  .group-title {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-secondary);
    margin-bottom: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .filter-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .filter-chip {
    padding: 8px 16px;
    background: var(--bg-secondary);
    border-radius: 20px;
    font-size: 13px;
    transition: all 0.15s ease;
  }

  .filter-chip:hover {
    background: var(--bg-tertiary);
  }

  .filter-chip.active {
    background: var(--accent);
    color: var(--bg-primary);
    font-weight: 600;
  }
</style>
