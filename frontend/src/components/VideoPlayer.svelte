<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';

  export let episode = null;
  export let episodes = [];
  export let dramaName = '';

  const dispatch = createEventDispatcher();

  let subtitleBlobURL = null;
  
  // Reactive statement untuk load subtitle
  $: if (subtitlesEnabled && selectedSubtitle && !isLocalFile) {
    loadSubtitleURL();
  }
  
  async function loadSubtitleURL() {
    if (selectedSubtitle?.subtitle) {
      subtitleBlobURL = await loadSubtitle(selectedSubtitle.subtitle);
    }
  }
  
  async function loadSubtitle(subtitleURL) {
    try {
      const response = await fetch(subtitleURL);
      const blob = await response.blob();
      return URL.createObjectURL(blob);
    } catch (error) {
      console.error('Error loading subtitle:', error);
      return null;
    }
  }
  
  let videoElement;
  let isPlaying = false;
  let currentTime = 0;
  let duration = 0;
  let volume = 1;
  let isFullscreen = false;
  let showControls = true;
  let showEpisodeList = false;
  let showSubtitleMenu = false;
  let selectedSubtitle = null;
  let subtitlesEnabled = true;
  let controlsTimeout;

  // Check if video is a local file
  $: isLocalFile = episode?.videoPath && (
    episode.videoPath.startsWith('/') ||
    episode.videoPath.startsWith('file://')
  );

  // Get proper video source URL
  $: videoSrc = (() => {
    const path = episode?.videoPath;
    if (!path) return '';
    
    // Already a remote URL
    if (path.startsWith('http://') || path.startsWith('https://')) {
      return path;
    }
    
    // Local file path - use /localfile/ endpoint
    if (path.match(/^[A-Za-z]:/)) {
      // Windows path - convert backslashes to forward slashes
      const cleanPath = path.replace(/\\/g, '/');
      return '/localfile/' + cleanPath;
    }
    
    // Unix path
    if (path.startsWith('/')) {
      return '/localfile' + path;
    }
    
    // file:// URL - convert to /localfile/
    if (path.startsWith('file:///')) {
      return '/localfile/' + path.substring(8);
    }
    if (path.startsWith('file://')) {
      return '/localfile' + path.substring(7);
    }
    
    return path;
  })();

  // Get available subtitles from current episode
  $: subtitles = episode?.subtitle || [];
  
  // Auto-select first subtitle when episode changes
  $: if (subtitles.length > 0 && !selectedSubtitle) {
    selectedSubtitle = subtitles[0];
  }

  onMount(() => {
    document.addEventListener('fullscreenchange', handleFullscreenChange);
    document.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    document.removeEventListener('fullscreenchange', handleFullscreenChange);
    document.removeEventListener('keydown', handleKeydown);
    clearTimeout(controlsTimeout);
  });

  function handleKeydown(e) {
    if (e.key === 'Escape') {
      if (showSubtitleMenu) {
        showSubtitleMenu = false;
      } else if (isFullscreen) {
        document.exitFullscreen();
      } else {
        closePlayer();
      }
    }
    // Toggle subtitles with 'C' key
    if (e.key === 'c' || e.key === 'C') {
      subtitlesEnabled = !subtitlesEnabled;
    }
  }

  function handleFullscreenChange() {
    isFullscreen = !!document.fullscreenElement;
  }

  function closePlayer() {
    dispatch('close');
  }

  function togglePlay() {
    if (videoElement.paused) {
      videoElement.play();
    } else {
      videoElement.pause();
    }
  }

  function handleTimeUpdate() {
    currentTime = videoElement.currentTime;
  }

  function handleLoadedMetadata() {
    duration = videoElement.duration;
  }

  function handlePlay() {
    isPlaying = true;
  }

  function handlePause() {
    isPlaying = false;
  }

  function seek(e) {
    const rect = e.currentTarget.getBoundingClientRect();
    const percent = (e.clientX - rect.left) / rect.width;
    videoElement.currentTime = percent * duration;
  }

  function formatTime(seconds) {
    if (isNaN(seconds)) return '0:00';
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }

  function toggleFullscreen() {
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      document.querySelector('.player-container').requestFullscreen();
    }
  }

  function toggleEpisodeList() {
    showEpisodeList = !showEpisodeList;
    showSubtitleMenu = false;
  }

  function toggleSubtitleMenu() {
    showSubtitleMenu = !showSubtitleMenu;
    showEpisodeList = false;
  }

  function selectSubtitle(sub) {
    selectedSubtitle = sub;
    subtitlesEnabled = true;
    showSubtitleMenu = false;
  }

  function disableSubtitles() {
    subtitlesEnabled = false;
    showSubtitleMenu = false;
  }

  function selectEpisode(ep) {
    selectedSubtitle = null; // Reset subtitle selection for new episode
    dispatch('select', ep);
  }

  function handleMouseMove() {
    showControls = true;
    clearTimeout(controlsTimeout);
    controlsTimeout = setTimeout(() => {
      if (isPlaying && !showEpisodeList && !showSubtitleMenu) showControls = false;
    }, 3000);
  }
</script>

<div class="player-container" on:mousemove={handleMouseMove}>
  <div class="video-wrapper">
    <video
      bind:this={videoElement}
      src={videoSrc}
      on:timeupdate={handleTimeUpdate}
      on:loadedmetadata={handleLoadedMetadata}
      on:play={handlePlay}
      on:pause={handlePause}
      on:click={togglePlay}
      on:ended={selectEpisode(episodes[episode.index + 1])}
      autoplay
    >
      {#if subtitlesEnabled && selectedSubtitle && !isLocalFile && subtitleBlobURL}
        <track 
          kind="subtitles" 
          src={subtitleBlobURL} 
          srclang={selectedSubtitle.language || 'id'} 
          label={selectedSubtitle.display_name || selectedSubtitle.language || 'Subtitle'}
          default
        />
      {/if}
    </video>

    <div class="controls" class:hidden={!showControls}>
      <div class="controls-top">
        <button class="close-btn" on:click={closePlayer} title="Close (Esc)">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
        <h3 class="video-title">{dramaName} - {episode?.eps}</h3>
        <div class="top-buttons">
          {#if subtitles.length > 0}
            <button class="top-btn" class:active={subtitlesEnabled} on:click={toggleSubtitleMenu} title="Subtitles (C)">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="4" width="20" height="16" rx="2"/>
                <path d="M7 15h4M13 15h4M7 11h2M11 11h6"/>
              </svg>
            </button>
          {/if}
          <button class="top-btn" on:click={toggleEpisodeList} title="Episodes">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
          </button>
        </div>
      </div>

      <div class="controls-bottom">
        <div class="progress-bar" on:click={seek} on:keydown={(e) => e.key === 'Enter' && seek(e)} role="slider" tabindex="0" aria-valuenow={currentTime} aria-valuemin={0} aria-valuemax={duration}>
          <div class="progress-track">
            <div class="progress-fill" style="width: {duration ? (currentTime / duration) * 100 : 0}%"></div>
          </div>
        </div>

        <div class="controls-row">
          <div class="controls-left">
            <button class="control-btn" on:click={togglePlay}>
              {#if isPlaying}
                <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                  <rect x="6" y="4" width="4" height="16"/>
                  <rect x="14" y="4" width="4" height="16"/>
                </svg>
              {:else}
                <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
                  <polygon points="5,3 19,12 5,21"/>
                </svg>
              {/if}
            </button>

            <span class="time-display">
              {formatTime(currentTime)} / {formatTime(duration)}
            </span>
          </div>

          <div class="controls-right">
            <button class="control-btn" on:click={toggleFullscreen}>
              {#if isFullscreen}
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M8 3v3a2 2 0 0 1-2 2H3m18 0h-3a2 2 0 0 1-2-2V3m0 18v-3a2 2 0 0 1 2-2h3M3 16h3a2 2 0 0 1 2 2v3"/>
                </svg>
              {:else}
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/>
                </svg>
              {/if}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Subtitle Menu -->
  {#if showSubtitleMenu}
    <div class="settings-panel">
      <div class="settings-panel-header">
        <h4>Subtitles</h4>
        <button class="panel-close-btn" on:click={toggleSubtitleMenu}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="settings-panel-body">
        <button 
          class="subtitle-option" 
          class:active={!subtitlesEnabled}
          on:click={disableSubtitles}
        >
          <span>Off</span>
          {#if !subtitlesEnabled}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <polyline points="20,6 9,17 4,12" stroke="currentColor" stroke-width="2" fill="none"/>
            </svg>
          {/if}
        </button>
        {#each subtitles as sub}
          <button 
            class="subtitle-option" 
            class:active={subtitlesEnabled && selectedSubtitle?.subtitle === sub.subtitle}
            on:click={() => selectSubtitle(sub)}
          >
            <span>{sub.display_name || sub.language || 'Unknown'}</span>
            {#if subtitlesEnabled && selectedSubtitle?.subtitle === sub.subtitle}
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                <polyline points="20,6 9,17 4,12" stroke="currentColor" stroke-width="2" fill="none"/>
              </svg>
            {/if}
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Floating Episode List -->
  {#if showEpisodeList}
    <div class="episode-panel">
      <div class="episode-panel-header">
        <h4>Episodes</h4>
        <button class="panel-close-btn" on:click={toggleEpisodeList}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="episode-panel-list">
        {#each episodes as ep}
          <button 
            class="episode-item" 
            class:active={episode && episode.index === ep.index}
            on:click={() => selectEpisode(ep)}
          >
            <span class="ep-number">{ep.eps}</span>
            {#if episode && episode.index === ep.index}
              <span class="now-playing">Now Playing</span>
            {/if}
          </button>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .player-container {
    position: fixed;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #000;
    z-index: 100;
  }

  .video-wrapper {
    position: relative;
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  video {
    width: 100%;
    height: 100%;
    object-fit: contain;
    cursor: pointer;
    background: #000;              /* Background hitam */
    outline: none;                 /* Hilangkan outline saat focus */
  }

  video::cue {
    background: rgba(0, 0, 0, 0); /* Background semi-transparan */
    color: #fffd9aff;
    font-size: 1em;
    padding: 6px 12px;              /* Padding lebih besar */
    border-radius: 4px;
    font-family: Arial, sans-serif;
    line-height: 1.5;
    position: 66%;
  }
  video::cue(:past) {
    opacity: 0.7;
  }
  video::cue(:future) {
    opacity: 0.5;
  }

  .controls {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    background: linear-gradient(to bottom, rgba(0,0,0,0.7) 0%, transparent 30%, transparent 70%, rgba(0,0,0,0.9) 100%);
    opacity: 1;
    transition: opacity 0.3s ease;
    pointer-events: none;
  }

  .controls.hidden {
    opacity: 0;
  }

  .controls-top,
  .controls-bottom {
    pointer-events: auto;
  }

  .controls-top {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px 20px;
  }

  .close-btn {
    padding: 8px;
    color: #fff;
    border-radius: 8px;
    transition: all 0.15s ease;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .video-title {
    flex: 1;
    font-size: 16px;
    font-weight: 500;
    color: #fff;
  }

  .top-buttons {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .top-btn {
    padding: 8px;
    color: #fff;
    border-radius: 8px;
    transition: all 0.15s ease;
  }

  .top-btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .top-btn.active {
    color: var(--accent, #6366f1);
  }

  .controls-bottom {
    padding: 16px 24px 24px;
  }

  .progress-bar {
    width: 100%;
    height: 20px;
    display: flex;
    align-items: center;
    cursor: pointer;
    margin-bottom: 12px;
  }

  .progress-track {
    width: 100%;
    height: 4px;
    background: rgba(255, 255, 255, 0.3);
    border-radius: 2px;
    overflow: hidden;
  }

  .progress-track:hover {
    height: 6px;
  }

  .progress-fill {
    height: 100%;
    background: #fff;
    border-radius: 2px;
    transition: width 0.1s linear;
  }

  .controls-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .controls-left,
  .controls-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .control-btn {
    padding: 8px;
    color: #fff;
    transition: transform 0.15s ease;
  }

  .control-btn:hover {
    transform: scale(1.1);
  }

  .time-display {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.8);
    font-variant-numeric: tabular-nums;
  }

  /* Settings Panel */
  .settings-panel {
    position: absolute;
    right: 20px;
    bottom: 100px;
    width: 250px;
    background: rgba(0, 0, 0, 0.95);
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    z-index: 20;
    overflow: hidden;
  }

  .settings-panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .settings-panel-header h4 {
    font-size: 14px;
    font-weight: 600;
    color: #fff;
  }

  .settings-panel-body {
    padding: 8px;
    max-height: 300px;
    overflow-y: auto;
  }

  .subtitle-option {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 12px;
    border-radius: 6px;
    color: rgba(255, 255, 255, 0.8);
    font-size: 13px;
    text-align: left;
    transition: all 0.15s ease;
  }

  .subtitle-option:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
  }

  .subtitle-option.active {
    background: rgba(255, 255, 255, 0.15);
    color: #fff;
  }

  /* Episode Panel */
  .episode-panel {
    position: absolute;
    right: 0;
    top: 0;
    bottom: 0;
    width: 300px;
    background: rgba(0, 0, 0, 0.95);
    border-left: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    flex-direction: column;
    z-index: 10;
  }

  .episode-panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .episode-panel-header h4 {
    font-size: 16px;
    font-weight: 600;
    color: #fff;
  }

  .panel-close-btn {
    padding: 4px;
    color: #fff;
    border-radius: 6px;
    transition: background 0.15s ease;
  }

  .panel-close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .episode-panel-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
  }

  .episode-item {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-radius: 8px;
    color: rgba(255, 255, 255, 0.8);
    transition: all 0.15s ease;
    text-align: left;
  }

  .episode-item:hover {
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
  }

  .episode-item.active {
    background: rgba(255, 255, 255, 0.15);
    color: #fff;
  }

  .ep-number {
    font-size: 14px;
    font-weight: 500;
  }

  .now-playing {
    font-size: 11px;
    padding: 2px 8px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 4px;
    color: #fff;
  }
</style>
