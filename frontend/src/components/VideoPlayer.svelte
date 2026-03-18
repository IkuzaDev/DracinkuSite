<script>
  import { onMount, onDestroy, createEventDispatcher } from 'svelte';

  export let episode = null;
  export let episodes = [];
  export let dramaName = '';

  const dispatch = createEventDispatcher();

  let subtitleBlobURL = null;
  
  // Generate CSS for subtitle based on settings
  $: subtitleCSS = `
    video::cue {
      background: ${subtitleSettings.backgroundColor};
      color: ${subtitleSettings.color};
      font-size: ${subtitleSettings.fontSize}em;
      font-weight: ${subtitleSettings.fontWeight};
      padding: 8px 16px;
      border-radius: 6px;
      font-family: Arial, sans-serif;
      line-height: 1.4;
      ${subtitleSettings.textShadow ? `text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.9);` : ''}
      ${subtitleSettings.edgeStyle === 'dropshadow' ? 'text-shadow: 2px 2px 2px black;' : ''}
      ${subtitleSettings.edgeStyle === 'raised' ? 'text-shadow: 1px 1px 1px black, -1px -1px 1px white;' : ''}
      ${subtitleSettings.edgeStyle === 'depressed' ? 'text-shadow: 1px 1px 2px black, -1px -1px 2px white;' : ''}
      ${subtitleSettings.edgeStyle === 'uniform' ? 'text-shadow: 0 0 2px black, 0 0 2px black;' : ''}
    }
    
    video::cue(:past) {
      opacity: 0.8;
    }
    
    video::cue(:future) {
      opacity: 0.6;
    }
  `;

  // Update subtitle position using JavaScript
  $: if (videoElement && videoElement.textTracks.length > 0) {
    updateSubtitlePosition();
  }

  function updateSubtitlePosition() {
    if (!videoElement) return;
    
    const track = videoElement.textTracks[0];
    if (!track || !track.cues) return;
    
    console.log('[VideoPlayer] Updating subtitle position to:', subtitleSettings.position);
    
    for (let i = 0; i < track.cues.length; i++) {
      const cue = track.cues[i];
      // Convert position percentage to line number
      // 0% (top) = line -3, 50% (middle) = line 0, 100% (bottom) = line 15
      const lineValue = Math.round(((subtitleSettings.position - 50) / 50) * 15);
      cue.line = lineValue;
      cue.position = 50; // Center horizontally
      cue.align = 'middle';
      
      console.log(`[VideoPlayer] Cue ${i} line set to:`, lineValue);
    }
  }

  // Handle position slider change
  function handlePositionChange() {
    updateSubtitlePosition();
  }

  // Reactive statement untuk load subtitle
  $: if (subtitlesEnabled && selectedSubtitle && selectedSubtitle.subtitle) {
    loadSubtitleURL();
  }
  
  // Inject dynamic CSS for subtitles
  $: if (subtitleCSS) {
    // Remove existing style element if any
    const existingStyle = document.getElementById('dynamic-subtitle-css');
    if (existingStyle) {
      existingStyle.remove();
    }
    
    // Create and append new style element
    const styleElement = document.createElement('style');
    styleElement.id = 'dynamic-subtitle-css';
    styleElement.textContent = subtitleCSS;
    document.head.appendChild(styleElement);
  }

  // Debug subtitle track when video element is ready
  $: if (videoElement && subtitleBlobURL) {
    console.log('[VideoPlayer] Video element ready, checking text tracks...');
    console.log('[VideoPlayer] Text tracks:', videoElement.textTracks);
    console.log('[VideoPlayer] Number of tracks:', videoElement.textTracks.length);
    
    for (let i = 0; i < videoElement.textTracks.length; i++) {
      const track = videoElement.textTracks[i];
      console.log(`[VideoPlayer] Track ${i}:`, {
        kind: track.kind,
        label: track.label,
        language: track.language,
        mode: track.mode,
        cues: track.cues?.length || 0
      });
    }
    
    // Enable the first subtitle track
    if (videoElement.textTracks.length > 0) {
      const subtitleTrack = videoElement.textTracks[0];
      subtitleTrack.mode = 'showing';
      console.log('[VideoPlayer] Subtitle track enabled:', subtitleTrack.label);
      
      // Add cue change listener for position updates
      subtitleTrack.oncuechange = () => {
        updateSubtitlePosition();
      };
    }
  }
  
  async function loadSubtitleURL() {
    if (selectedSubtitle?.subtitle) {
      // Clean up previous blob URL
      if (subtitleBlobURL) {
        URL.revokeObjectURL(subtitleBlobURL);
      }
      subtitleBlobURL = await loadSubtitle(selectedSubtitle.subtitle);
    }
  }
  
  // Convert SRT to WebVTT format
  function convertSRTToWebVTT(srtText) {
    try {
      console.log('[VideoPlayer] Converting SRT to WebVTT...');
      
      // Add WebVTT header
      let vttText = 'WEBVTT\n\n';
      
      // Split by double newlines to get subtitle blocks
      const blocks = srtText.trim().split(/\n\s*\n/);
      
      blocks.forEach(block => {
        const lines = block.trim().split('\n');
        if (lines.length >= 3) {
          // First line: index
          const index = lines[0].trim();
          
          // Second line: timestamp (SRT format: 00:00:22,055 --> 00:00:25,356)
          const timestamp = lines[1].trim();
          const vttTimestamp = timestamp.replace(/,/g, '.'); // Replace comma with dot for WebVTT
          
          // Remaining lines: subtitle text
          const text = lines.slice(2).join('\n').trim();
          
          if (timestamp.includes('-->')) {
            vttText += `${index}\n${vttTimestamp}\n${text}\n\n`;
          }
        }
      });
      
      console.log('[VideoPlayer] SRT to WebVTT conversion completed');
      console.log('[VideoPlayer] Total subtitle blocks:', blocks.length);
      return vttText;
    } catch (error) {
      console.error('[VideoPlayer] Error converting SRT to WebVTT:', error);
      return srtText; // Return original if conversion fails
    }
  }

  async function loadSubtitle(subtitleURL) {
    try {
      console.log('[VideoPlayer] Loading subtitle:', subtitleURL);
      
      // Handle both local and remote subtitle files
      let response;
      if (subtitleURL.startsWith('http://') || subtitleURL.startsWith('https://')) {
        // Remote subtitle
        console.log('[VideoPlayer] Loading remote subtitle');
        response = await fetch(subtitleURL);
      } else {
        // Local subtitle file - use /localfile/ endpoint
        console.log('[VideoPlayer] Loading local subtitle');
        let cleanPath = subtitleURL.replace(/\\/g, '/');
        
        // Ensure path starts with / for /localfile/ endpoint
        if (!cleanPath.startsWith('/')) {
          cleanPath = '/' + cleanPath;
        }
        
        // For Windows paths that start with drive letter, make sure they're properly formatted
        if (cleanPath.match(/^\/[A-Za-z]:/)) {
          // Path like /C:/path/to/file.srt - this is correct
          response = await fetch('/localfile' + cleanPath);
        } else if (cleanPath.startsWith('//')) {
          // Network path or double slash, convert to single
          cleanPath = cleanPath.substring(1);
          response = await fetch('/localfile' + cleanPath);
        } else {
          // Unix-style path
          response = await fetch('/localfile' + cleanPath);
        }
      }
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const subtitleText = await response.text();
      console.log('[VideoPlayer] Subtitle text length:', subtitleText.length);
      console.log('[VideoPlayer] Subtitle preview:', subtitleText.substring(0, 200));
      
      // Convert SRT to WebVTT format
      const vttText = convertSRTToWebVTT(subtitleText);
      console.log('[VideoPlayer] WebVTT text preview:', vttText.substring(0, 500));
      
      // Create blob with WebVTT MIME type
      const blob = new Blob([vttText], { type: 'text/vtt;charset=utf-8' });
      const blobURL = URL.createObjectURL(blob);
      console.log('[VideoPlayer] Subtitle loaded successfully, blob URL:', blobURL);
      console.log('[VideoPlayer] Subtitle MIME type: text/vtt');
      return blobURL;
    } catch (error) {
      console.error('[VideoPlayer] Error loading subtitle:', error);
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
  
  // Subtitle settings
  let subtitleSettings = {
    fontSize: 1.2,
    color: '#ffffff',
    backgroundColor: 'rgba(0, 0, 0, 0.8)',
    fontWeight: 'bold',
    position: 85, // Position percentage (0-100)
    textShadow: true,
    edgeStyle: 'none'
  };

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
    
    // Clean up subtitle blob URL
    if (subtitleBlobURL) {
      URL.revokeObjectURL(subtitleBlobURL);
    }
    
    // Clean up dynamic CSS
    const dynamicStyle = document.getElementById('dynamic-subtitle-css');
    if (dynamicStyle) {
      dynamicStyle.remove();
    }
  });

  // Handle video element events for subtitle debugging
  function handleVideoLoadStart() {
    console.log('[VideoPlayer] Video load start event');
  }

  function handleVideoLoadedData() {
    console.log('[VideoPlayer] Video loaded data event');
    if (videoElement) {
      console.log('[VideoPlayer] Video ready state:', videoElement.readyState);
      console.log('[VideoPlayer] Video text tracks after load:', videoElement.textTracks.length);
    }
  }

  function handleVideoCanPlay() {
    console.log('[VideoPlayer] Video can play event');
    // Force enable subtitles when video can play
    if (videoElement && videoElement.textTracks.length > 0 && subtitlesEnabled) {
      const subtitleTrack = videoElement.textTracks[0];
      subtitleTrack.mode = 'showing';
      console.log('[VideoPlayer] Forced subtitle track enable on can play');
    }
  }

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
      if (subtitlesEnabled) {
        disableSubtitles();
      } else if (selectedSubtitle) {
        selectSubtitle(selectedSubtitle);
      } else {
        subtitlesEnabled = !subtitlesEnabled;
      }
      console.log('[VideoPlayer] Subtitles toggled:', subtitlesEnabled);
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
    
    // Force enable subtitle track after selection
    setTimeout(() => {
      if (videoElement && videoElement.textTracks.length > 0) {
        const subtitleTrack = videoElement.textTracks[0];
        subtitleTrack.mode = 'showing';
        console.log('[VideoPlayer] Subtitle track enabled after selection:', subtitleTrack.label);
      }
    }, 100);
  }

  function disableSubtitles() {
    subtitlesEnabled = false;
    showSubtitleMenu = false;
    
    // Disable all subtitle tracks
    if (videoElement) {
      for (let i = 0; i < videoElement.textTracks.length; i++) {
        const track = videoElement.textTracks[i];
        if (track.kind === 'subtitles') {
          track.mode = 'disabled';
        }
      }
      console.log('[VideoPlayer] All subtitle tracks disabled');
    }
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
      on:loadstart={handleVideoLoadStart}
      on:loadeddata={handleVideoLoadedData}
      on:canplay={handleVideoCanPlay}
      autoplay
    >
      {#if subtitlesEnabled && selectedSubtitle && selectedSubtitle.subtitle && subtitleBlobURL}
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
        <h4>Subtitle Settings</h4>
        <button class="panel-close-btn" on:click={toggleSubtitleMenu}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      <div class="settings-panel-body">
        <!-- Subtitle On/Off -->
        <div class="settings-section">
          <h5>Subtitle</h5>
          <div class="subtitle-options">
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

        <!-- Subtitle Style Settings -->
        {#if subtitlesEnabled}
          <div class="settings-section">
            <h5>Style</h5>
            
            <!-- Font Size -->
            <div class="setting-row">
              <label>Size</label>
              <div class="size-controls">
                <button class="size-btn" on:click={() => subtitleSettings.fontSize = Math.max(0.8, subtitleSettings.fontSize - 0.1)}>-</button>
                <span class="size-value">{Math.round(subtitleSettings.fontSize * 10) / 10}em</span>
                <button class="size-btn" on:click={() => subtitleSettings.fontSize = Math.min(2.0, subtitleSettings.fontSize + 0.1)}>+</button>
              </div>
            </div>

            <!-- Color -->
            <div class="setting-row">
              <label>Color</label>
              <div class="color-controls">
                <input type="color" bind:value={subtitleSettings.color} class="color-picker">
                <div class="preset-colors">
                  {#each ['#ffffff', '#ffff00', '#00ff00', '#00ffff', '#ff00ff', '#ff8800'] as color}
                    <button 
                      class="preset-color" 
                      class:active={subtitleSettings.color === color}
                      style="background-color: {color}"
                      on:click={() => subtitleSettings.color = color}
                    ></button>
                  {/each}
                </div>
              </div>
            </div>

            <!-- Background Color -->
            <div class="setting-row">
              <label>Background</label>
              <div class="bg-controls">
                <button 
                  class="bg-option" 
                  class:active={subtitleSettings.backgroundColor === 'rgba(0, 0, 0, 0.8)'}
                  on:click={() => subtitleSettings.backgroundColor = 'rgba(0, 0, 0, 0.8)'}
                >Dark</button>
                <button 
                  class="bg-option" 
                  class:active={subtitleSettings.backgroundColor === 'rgba(0, 0, 0, 0)'}
                  on:click={() => subtitleSettings.backgroundColor = 'rgba(0, 0, 0, 0)'}
                >None</button>
                <button 
                  class="bg-option" 
                  class:active={subtitleSettings.backgroundColor === 'rgba(0, 0, 0, 0.5)'}
                  on:click={() => subtitleSettings.backgroundColor = 'rgba(0, 0, 0, 0.5)'}
                >Semi</button>
              </div>
            </div>

            <!-- Font Weight -->
            <div class="setting-row">
              <label>Weight</label>
              <div class="weight-controls">
                <button 
                  class="weight-option" 
                  class:active={subtitleSettings.fontWeight === 'normal'}
                  on:click={() => subtitleSettings.fontWeight = 'normal'}
                >Normal</button>
                <button 
                  class="weight-option" 
                  class:active={subtitleSettings.fontWeight === 'bold'}
                  on:click={() => subtitleSettings.fontWeight = 'bold'}
                >Bold</button>
              </div>
            </div>

            <!-- Position -->
            <div class="setting-row">
              <label>Position</label>
              <div class="position-slider-container">
                <input 
                  type="range" 
                  min="0" 
                  max="100" 
                  bind:value={subtitleSettings.position}
                  on:input={handlePositionChange}
                  class="position-slider"
                />
                <div class="position-labels">
                  <span>Top</span>
                  <span>Middle</span>
                  <span>Bottom</span>
                </div>
                <div class="position-value">{subtitleSettings.position}%</div>
              </div>
            </div>

            <!-- Text Shadow -->
            <div class="setting-row">
              <label>Shadow</label>
              <div class="shadow-controls">
                <button 
                  class="shadow-option" 
                  class:active={!subtitleSettings.textShadow}
                  on:click={() => subtitleSettings.textShadow = false}
                >None</button>
                <button 
                  class="shadow-option" 
                  class:active={subtitleSettings.textShadow}
                  on:click={() => subtitleSettings.textShadow = true}
                >On</button>
              </div>
            </div>
          </div>
        {/if}
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
    background: #000;
    outline: none;
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
    width: 300px;
    background: rgba(0, 0, 0, 0.95);
    border-radius: 12px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    z-index: 20;
    overflow: hidden;
    max-height: 80vh;
    overflow-y: auto;
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
  }

  .settings-section {
    margin-bottom: 16px;
  }

  .settings-section h5 {
    font-size: 12px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.7);
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin: 8px 12px 4px;
  }

  .subtitle-options {
    display: flex;
    flex-direction: column;
    gap: 4px;
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

  .setting-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    gap: 12px;
  }

  .setting-row label {
    font-size: 12px;
    color: rgba(255, 255, 255, 0.8);
    min-width: 60px;
  }

  /* Size Controls */
  .size-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .size-btn {
    width: 24px;
    height: 24px;
    border: none;
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.15s ease;
  }

  .size-btn:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .size-value {
    font-size: 11px;
    color: #fff;
    min-width: 40px;
    text-align: center;
  }

  /* Color Controls */
  .color-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .color-picker {
    width: 32px;
    height: 24px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 4px;
    cursor: pointer;
    background: transparent;
  }

  .preset-colors {
    display: flex;
    gap: 4px;
  }

  .preset-color {
    width: 20px;
    height: 20px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 3px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .preset-color:hover {
    transform: scale(1.1);
  }

  .preset-color.active {
    border-color: #fff;
    box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.3);
  }

  /* Background Controls */
  .bg-controls {
    display: flex;
    gap: 4px;
  }

  .bg-option {
    padding: 4px 8px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .bg-option:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .bg-option.active {
    background: rgba(255, 255, 255, 0.3);
    border-color: #fff;
  }

  /* Weight Controls */
  .weight-controls {
    display: flex;
    gap: 4px;
  }

  .weight-option {
    padding: 4px 12px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .weight-option:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .weight-option.active {
    background: rgba(255, 255, 255, 0.3);
    border-color: #fff;
  }

  /* Position Controls */
  .position-slider-container {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .position-slider {
    width: 100%;
    height: 4px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.2);
    outline: none;
    cursor: pointer;
    -webkit-appearance: none;
    appearance: none;
  }

  .position-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #fff;
    cursor: pointer;
    border: 2px solid rgba(255, 255, 255, 0.3);
    transition: all 0.15s ease;
  }

  .position-slider::-webkit-slider-thumb:hover {
    transform: scale(1.1);
    border-color: rgba(255, 255, 255, 0.5);
  }

  .position-slider::-moz-range-thumb {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: #fff;
    cursor: pointer;
    border: 2px solid rgba(255, 255, 255, 0.3);
    transition: all 0.15s ease;
  }

  .position-slider::-moz-range-thumb:hover {
    transform: scale(1.1);
    border-color: rgba(255, 255, 255, 0.5);
  }

  .position-labels {
    display: flex;
    justify-content: space-between;
    font-size: 10px;
    color: rgba(255, 255, 255, 0.6);
  }

  .position-value {
    font-size: 11px;
    color: #fff;
    text-align: center;
    font-weight: 500;
  }

  .position-controls {
    display: flex;
    gap: 4px;
  }

  .position-option {
    padding: 4px 8px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .position-option:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .position-option.active {
    background: rgba(255, 255, 255, 0.3);
    border-color: #fff;
  }

  /* Shadow Controls */
  .shadow-controls {
    display: flex;
    gap: 4px;
  }

  .shadow-option {
    padding: 4px 12px;
    border: 1px solid rgba(255, 255, 255, 0.3);
    background: rgba(255, 255, 255, 0.1);
    color: #fff;
    border-radius: 4px;
    font-size: 11px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .shadow-option:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .shadow-option.active {
    background: rgba(255, 255, 255, 0.3);
    border-color: #fff;
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
