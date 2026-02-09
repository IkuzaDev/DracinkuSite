<script>
  import { onMount, createEventDispatcher } from 'svelte';
  import heic2any from 'heic2any';
  import { GetSeriesDetail, StartDownload, AddToLibrary, GetDownloadDir, GetDownloadProgress } from '../../wailsjs/go/main/App';
  import DownloadDialog from './DownloadDialog.svelte';

  export let platform = '';
  export let language = 'id';
  export let drama = null;
  const convertedCache = new Map()
  let imageSrc = drama.cover || '';

  onMount(async () => {
    if (isHeicUrl(drama.cover)) {
      imageSrc = await convertHeicToJpeg(drama.cover);
    }
  });

  const dispatch = createEventDispatcher();
  let episodes = [];
  let loading = true;
  let error = '';
  let downloadingEps = new Set();
  
  // Download dialog state
  let showDownloadDialog = false;
  let selectedEpisodeForDownload = null;
  
  // Bulk download state
  let isBulkDownload = false;
  let bulkSubtitleSelection = null;

  onMount(async () => {
    if (drama?.id) {
      await loadDetail();
    }
  });


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

  async function loadDetail() {
    loading = true;
    error = '';
    try {
      const result = await GetSeriesDetail(platform, drama.id, 720, language);
      episodes = result?.chapters || [];
      if (result?.book) {
        drama = { ...drama, ...result.book };
      }
    } catch (e) {
      error = e.message || 'Failed to load detail';
    } finally {
      loading = false;
    }
  }

  function playEpisode(episode) {
    dispatch('play', { episode, episodes });
  }

  function openDownloadDialog(episode) {
    isBulkDownload = false;
    selectedEpisodeForDownload = episode;
    showDownloadDialog = true;
  }

  async function handleDownloadConfirm(event) {
    const { episode, subtitle } = event.detail;
    
    // Handle bulk download
    if (isBulkDownload) {
      bulkSubtitleSelection = subtitle;
      isBulkDownload = false; // Reset before processing
      await processBulkDownload();
      return;
    }
    
    await downloadSingleEpisode(episode, subtitle);
  }

  async function downloadSingleEpisode(episode, subtitle) {
    if (downloadingEps.has(episode.index)) return;
    downloadingEps.add(episode.index);
    downloadingEps = downloadingEps;

    try {
      const filename = `${drama.name} - ${episode.eps}`.replace(/[<>:"/\\|?*]/g, '');
      
      // Pass only the selected subtitle (or empty array if no subtitle)
      const subtitles = subtitle ? [{
        url: subtitle.subtitle || subtitle.url || '',
        language: subtitle.language || '',
        display_name: subtitle.display_name || subtitle.language || ''
      }] : [];
      
      const embedSubs = subtitle !== null;
      
      const downloadId = await StartDownload(episode.videoPath, subtitles, embedSubs, filename, '');
      
      // Get proper download directory from backend
      const downloadDir = await GetDownloadDir();
      const filePath = `${downloadDir}\\${filename}.mp4`;
      
      // Add to library for offline playback
      await AddToLibrary({
        id: downloadId,
        dramaId: String(drama.id),
        dramaName: drama.name || '',
        dramaCover: drama.cover || '',
        episodeIndex: episode.index || 0,
        episodeName: episode.eps || `Episode ${episode.index + 1}`,
        platform: platform,
        filePath: filePath,
        hasSubtitles: embedSubs,
        downloadedAt: new Date().toISOString(),
        watchProgress: 0
      });

      return downloadId;
    } catch (e) {
      console.error('Download failed:', e);
      return null;
    }
  }

  function waitForDownload(id) {
    return new Promise((resolve, reject) => {
      const interval = setInterval(async () => {
        try {
          const progress = await GetDownloadProgress(id);
          if (!progress) {
             clearInterval(interval);
             reject("Download not found");
             return;
          }
          if (progress.status === 'completed') {
            clearInterval(interval);
            resolve();
          } else if (progress.status === 'failed' || progress.status === 'cancelled') {
            clearInterval(interval);
            reject(progress.error || "Download failed");
          }
        } catch (e) {
          clearInterval(interval);
          reject(e);
        }
      }, 1000);
    });
  }

  function downloadAll() {
    const refEpisode = episodes.find(ep => ep.videoPath && ep.subtitle && ep.subtitle.length > 0) || episodes[0];
    if (!refEpisode) return;
    
    isBulkDownload = true;
    selectedEpisodeForDownload = refEpisode;
    showDownloadDialog = true;
  }

  async function processBulkDownload() {
    // If subtitles are selected (hardsub), process sequentially to avoid high CPU usage
    const isHardsub = !!bulkSubtitleSelection;
    const BATCH_SIZE = isHardsub ? 1 : 3;
    
    const pendingEpisodes = episodes.filter(ep => ep.videoPath && !downloadingEps.has(ep.index));
    
    for (let i = 0; i < pendingEpisodes.length; i += BATCH_SIZE) {
      const batch = pendingEpisodes.slice(i, i + BATCH_SIZE);
      
      // Process batch and wait for completion
      await Promise.all(
        batch.map(async (ep) => {
           const id = await downloadSingleEpisode(ep, bulkSubtitleSelection);
           if (id && isHardsub) {
               // Wait for completion for hardsubs (strict sequential)
               await waitForDownload(id);
           } else if (id) {
               // Also wait for batch completion for direct downloads
               await waitForDownload(id);
           }
        })
      );

      // Add delay between batches/episodes if hardsub
      if (isHardsub) {
          await new Promise(r => setTimeout(r, 2000));
      }
    }
    
    bulkSubtitleSelection = null;
  }
</script>

<DownloadDialog 
  episode={selectedEpisodeForDownload}
  dramaName={drama?.name}
  {isBulkDownload}
  bind:show={showDownloadDialog}
  on:confirm={handleDownloadConfirm}
/>

<div class="detail-container">
  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Loading detail...</p>
    </div>
  {:else if error}
    <div class="error-state">
      <p>{error}</p>
      <button class="btn btn-secondary" on:click={loadDetail}>Retry</button>
    </div>
  {:else}
    <div class="detail-header">
      <div class="cover-section">
        {#if drama?.cover}
          <img src={imageSrc} alt={drama.name} class="detail-cover" />
        {:else}
          <div class="cover-placeholder">
            <svg width="60" height="60" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <rect x="2" y="2" width="20" height="20" rx="2"/>
              <circle cx="8.5" cy="8.5" r="1.5"/>
              <polyline points="21,15 16,10 5,21"/>
            </svg>
          </div>
        {/if}
      </div>

      <div class="info-section">
        <h1 class="drama-title">{drama?.name}</h1>
        
        <div class="meta-info">
          {#if drama?.chapterCount}
            <span class="meta-item">{drama.chapterCount} Episodes</span>
          {/if}
          {#if drama?.playCount}
            <span class="meta-item">{drama.playCount} Views</span>
          {/if}
        </div>

        {#if drama?.introduction}
          <p class="description">{drama.introduction}</p>
        {/if}

        <div class="action-buttons">
          {#if episodes.length > 0}
            <button class="btn btn-primary" on:click={() => playEpisode(episodes[0])}>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                <polygon points="5,3 19,12 5,21"/>
              </svg>
              Play
            </button>
          {/if}
          <button class="btn btn-secondary" on:click={downloadAll}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7,10 12,15 17,10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            Download All
          </button>
        </div>
        

      </div>
    </div>

    <div class="episodes-section">
      <h2 class="section-title">Episodes ({episodes.length})</h2>
      <div class="episodes-grid">
        {#each episodes as episode}
          <div class="episode-card">
            <div class="episode-info">
              <span class="episode-number">{episode.eps}</span>
            </div>
            <div class="episode-actions">
              <button class="btn-icon" on:click={() => playEpisode(episode)} title="Play">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                  <polygon points="5,3 19,12 5,21"/>
                </svg>
              </button>
              <button 
                class="btn-icon" 
                on:click={() => openDownloadDialog(episode)} 
                title="Download"
                disabled={downloadingEps.has(episode.index)}
              >
                {#if downloadingEps.has(episode.index)}
                  <div class="mini-spinner"></div>
                {:else}
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                    <polyline points="7,10 12,15 17,10"/>
                    <line x1="12" y1="15" x2="12" y2="3"/>
                  </svg>
                {/if}
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .detail-container {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .detail-header {
    display: flex;
    gap: 32px;
    margin-bottom: 40px;
  }

  .cover-section {
    flex-shrink: 0;
  }

  .detail-cover {
    width: 240px;
    height: 320px;
    object-fit: cover;
    border-radius: 12px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  }

  .cover-placeholder {
    width: 240px;
    height: 320px;
    background: var(--bg-tertiary);
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
  }

  .info-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .drama-title {
    font-size: 28px;
    font-weight: 700;
    line-height: 1.2;
  }

  .meta-info {
    display: flex;
    gap: 16px;
  }

  .meta-item {
    font-size: 14px;
    color: var(--text-secondary);
  }

  .description {
    font-size: 14px;
    line-height: 1.6;
    color: var(--text-secondary);
    max-width: 600px;
  }

  .action-buttons {
    display: flex;
    gap: 12px;
    margin-top: auto;
  }

  .section-title {
    font-size: 18px;
    font-weight: 600;
    margin-bottom: 16px;
  }

  .episodes-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 12px;
  }

  .episode-card {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border-radius: 8px;
    transition: background 0.15s ease;
  }

  .episode-card:hover {
    background: var(--bg-tertiary);
  }

  .episode-number {
    font-size: 14px;
    font-weight: 500;
  }

  .episode-actions {
    display: flex;
    gap: 8px;
  }

  .mini-spinner {
    width: 18px;
    height: 18px;
    border: 2px solid var(--bg-tertiary);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  .loading-state,
  .error-state {
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

  .download-options {
    margin-top: 16px;
    padding: 12px 16px;
    background: var(--bg-tertiary);
    border-radius: 10px;
  }

  .option-toggle {
    display: flex;
    align-items: center;
    gap: 12px;
    cursor: pointer;
  }

  .option-toggle input {
    display: none;
  }

  .toggle-slider {
    position: relative;
    width: 44px;
    height: 24px;
    background: var(--bg-hover);
    border-radius: 12px;
    transition: background 0.2s ease;
  }

  .toggle-slider::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    width: 20px;
    height: 20px;
    background: #fff;
    border-radius: 50%;
    transition: transform 0.2s ease;
  }

  .option-toggle input:checked + .toggle-slider {
    background: var(--accent);
  }

  .option-toggle input:checked + .toggle-slider::after {
    transform: translateX(20px);
  }

  .toggle-label {
    font-size: 13px;
    color: var(--text-secondary);
  }
</style>
