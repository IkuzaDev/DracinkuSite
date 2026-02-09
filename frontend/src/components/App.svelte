<script>
  import { onMount } from 'svelte';
  import Sidebar from './components/Sidebar.svelte';
  import DramaList from './components/DramaList.svelte';
  import DramaDetail from './components/DramaDetail.svelte';
  import VideoPlayer from './components/VideoPlayer.svelte';
  import DownloadQueue from './components/DownloadQueue.svelte';
  import Library from './components/Library.svelte';
  import { GetPlatformList, GetTabList, SearchSeries, CheckLicense, SetLicenseKey } from '../wailsjs/go/main/App';
  import { EventsOn } from '../wailsjs/runtime/runtime';

  let platforms = [];
  let selectedPlatform = null;
  let tabs = [];
  let selectedTab = null;
  let searchKeyword = '';
  let currentView = 'list';
  let selectedDrama = null;
  let selectedEpisode = null;
  let currentEpisodes = [];
  let showDownloads = false;
  let offlineItem = null; 

  let ffmpegStatus = '';
  let ffmpegProgress = 0;

  // License & Language
  let isLicensed = false;
  let licenseInput = '';
  let checkingLicense = true;
  
  let languages = [
    { code: 'id', name: 'Indonesia' },
    { code: 'en', name: 'English' },
    { code: 'es', name: 'Español' },
    { code: 'vi', name: 'Tiếng Việt' },
    { code: 'th', name: 'ไทย' },
    { code: 'zhtw', name: '繁體中文' },
    { code: 'zhcn', name: '简体中文' },
    { code: 'ko', name: '한국어' },
    { code: 'ja', name: '日本語' }
  ];
  let selectedLanguage = 'id';

  onMount(async () => {
    // Listen for FFmpeg events
    EventsOn('ffmpeg_status', (status) => {
      ffmpegStatus = status;
      if (status === 'ready') {
        setTimeout(() => ffmpegStatus = '', 3000);
      }
    });

    EventsOn('ffmpeg_progress', (progress) => {
      ffmpegProgress = progress;
    });

    // Check license first
    try {
        isLicensed = await CheckLicense();
    } catch (e) {
        console.error("License check failed:", e);
    }
    checkingLicense = false;

    if (isLicensed) {
        await loadPlatforms();
    }
  });

  async function loadPlatforms() {
    platforms = await GetPlatformList();
    if (platforms.length > 0) {
      selectPlatform(platforms[0]);
    }
  }

  async function handleLogin() {
      if (!licenseInput) return;
      try {
          await SetLicenseKey(licenseInput);
          isLicensed = true;
          await loadPlatforms();
      } catch (e) {
          alert("License activation failed: " + e);
      }
  }

  async function selectPlatform(platform) {
    selectedPlatform = platform;
    selectedTab = null;
    currentView = 'list';
    try {
      // Pass language to GetTabList
      tabs = await GetTabList(platform.id, selectedLanguage);
      if (tabs && tabs.length > 0) {
        selectedTab = tabs[0];
      }
    } catch (e) {
      console.error('Failed to get tabs:', e);
      tabs = [];
    }
  }

  function handleLanguageChange() {
      if (selectedPlatform) {
          selectPlatform(selectedPlatform);
      }
  }

  async function handleSearch() {
    if (!searchKeyword.trim() || !selectedPlatform) return;
    currentView = 'search';
  }

  function openDrama(drama) {
    selectedDrama = drama;
    currentView = 'detail';
  }

  function playEpisode(episode, episodes = []) {
    offlineItem = null;
    selectedEpisode = episode;
    if (episodes.length > 0) {
      currentEpisodes = episodes;
    }
    currentView = 'player';
  }

  function playOffline(data) {
    const { item, episodes } = data;
    offlineItem = item;
    
    // Convert library items to episode format for VideoPlayer
    currentEpisodes = (episodes || []).map(ep => ({
      eps: ep.episodeName,
      index: ep.episodeIndex,
      videoPath: ep.filePath
    }));
    
    selectedEpisode = {
      eps: item.episodeName,
      index: item.episodeIndex,
      videoPath: item.filePath
    };
    selectedDrama = { name: item.dramaName };
    currentView = 'player';
  }

  function closePlayer() {
    // If playing from offline library, go back to library
    if (offlineItem) {
      currentView = 'library';
    } else {
      currentView = 'detail';
    }
    selectedEpisode = null;
    offlineItem = null;
  }
  
  function selectEpisodeFromPlayer(episode) {
      // Just update selected episode to play new one
      selectedEpisode = episode;
  }
  
  function showLibrary() {
    currentView = 'library';
  }
  
  function goBack() {
    if (currentView === 'detail' || currentView === 'search') {
      currentView = 'list';
      searchKeyword = '';
    } else if (currentView === 'player') {
      closePlayer();
    } else if (currentView === 'library') {
      currentView = 'list';
    }
  }
</script>

<main class="app-container">
  {#if !isLicensed && !checkingLicense}
      <div class="login-overlay">
          <div class="login-box">
              <h2 class="login-title">Activate Dracinku</h2>
              <p style="margin-bottom: 20px; color: var(--text-secondary);">Please enter your license key to continue.</p>
              <input type="text" class="login-input" placeholder="Enter License Key" bind:value={licenseInput} />
              <button class="login-btn" on:click={handleLogin}>Activate License</button>
          </div>
      </div>
  {/if}

  <Sidebar 
    {platforms} 
    {selectedPlatform} 
    on:select={(e) => selectPlatform(e.detail)}
    on:toggleDownloads={() => showDownloads = !showDownloads}
    on:showLibrary={showLibrary}
  />
  
  <div class="main-content">
    <header class="header">
      <div class="header-left">
        {#if currentView !== 'list'}
          <button class="btn-icon" on:click={goBack}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M19 12H5M12 19l-7-7 7-7"/>
            </svg>
          </button>
        {/if}
        <h1 class="page-title">
          {#if currentView === 'list'}
            {selectedPlatform?.name + ' || Dracinku ' || 'Dracinku.site'}
          {:else if currentView === 'search'}
            Search Results
          {:else if currentView === 'detail'}
            {selectedDrama?.name || 'Drama Detail'}
          {:else if currentView === 'player'}
            {selectedEpisode?.eps || 'Player'}
          {/if}
        </h1>
      </div> 
      
      <div class="header-right">
        <!-- Language Switcher -->
        <select class="lang-select" bind:value={selectedLanguage} on:change={handleLanguageChange}>
            {#each languages as lang}
                <option value={lang.code}>{lang.name}</option>
            {/each}
        </select>

        <div class="search-box">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <path d="M21 21l-4.35-4.35"/>
          </svg>
          <input 
            type="text" 
            placeholder="Search drama..." 
            bind:value={searchKeyword}
            on:keydown={(e) => e.key === 'Enter' && handleSearch()}
          />
        </div>
      </div>
    </header>

    {#if currentView === 'list'}
      <DramaList 
        platform={selectedPlatform?.id} 
        language={selectedLanguage}
        on:select={(e) => openDrama(e.detail)}
      />
    {:else if currentView === 'search'}
      <DramaList 
        platform={selectedPlatform?.id}
        searchKeyword={searchKeyword}
        language={selectedLanguage}
        on:select={(e) => openDrama(e.detail)}
      />
    {:else if currentView === 'detail'}
      <DramaDetail 
        platform={selectedPlatform?.id}
        drama={selectedDrama}
        language={selectedLanguage}
        on:play={(e) => playEpisode(e.detail.episode, e.detail.episodes)}
      />
    {:else if currentView === 'player'}
      <VideoPlayer 
        episode={selectedEpisode}
        episodes={currentEpisodes}
        dramaName={selectedDrama?.name}
        on:close={closePlayer}
        on:select={(e) => selectEpisodeFromPlayer(e.detail)}
      />
    {:else if currentView === 'library'}
      <Library on:playOffline={(e) => playOffline(e.detail)} />
    {/if}
  </div>

  {#if ffmpegStatus && ffmpegStatus !== 'ready'}
    <div class="ffmpeg-status-bar">
      <div class="status-content">
        {#if ffmpegStatus === 'checking'}
          <span>Checking Required Tools...</span>
        {:else if ffmpegStatus === 'downloading'}
          <span>Downloading Required Tools [{Math.round(ffmpegProgress)}%]</span>
          <div class="progress-track">
            <div class="progress-fill" style="width: {ffmpegProgress}%"></div>
          </div>
        {:else if ffmpegStatus === 'error'}
          <span class="error">Failed to download Required Tools. Please install manually.</span>
        {/if}
      </div>
    </div>
  {:else if ffmpegStatus === 'ready'}
    <div class="ffmpeg-status-bar success">
      <span>Required Tools is ready!</span>
    </div>
  {/if}

  {#if showDownloads}
    <DownloadQueue on:close={() => showDownloads = false} />
  {/if}
</main>

<style>
  .app-container {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 24px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .page-title {
    font-size: 20px;
    font-weight: 600;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .search-box {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 16px;
    background: var(--bg-tertiary);
    border-radius: 10px;
    border: 1px solid var(--border-color);
    width: 300px;
  }

  .search-box input {
    flex: 1;
    font-size: 14px;
    background: transparent;
    border: none;
    color: var(--text-primary);
    outline: none;
  }

  .search-box input::placeholder {
    color: var(--text-muted);
  }

  .tabs-container {
    display: flex;
    gap: 8px;
    padding: 16px 24px;
    overflow-x: auto;
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border-color);
  }

  .tab-btn {
    padding: 8px 16px;
    border-radius: 20px;
    font-size: 13px;
    font-weight: 500;
    color: var(--text-secondary);
    background: var(--bg-tertiary);
    white-space: nowrap;
    transition: all 0.2s ease;
  }

  .tab-btn:hover {
    background: var(--bg-hover);
  }

  .tab-btn.active {
    background: var(--accent);
    color: var(--bg-primary);
  }

  .ffmpeg-status-bar {
    position: fixed;
    bottom: 20px;
    left: 50%;
    transform: translateX(-50%);
    background: var(--bg-secondary);
    padding: 12px 24px;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
    z-index: 1000;
    border: 1px solid var(--border-color);
    min-width: 300px;
  }

  .ffmpeg-status-bar.success {
    background: var(--success);
    color: #fff;
    border: none;
  }

  .status-content {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .progress-track {
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

  .error {
    color: var(--error);
  }

  /* Login Overlay */
  .login-overlay {
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: var(--bg-primary);
    z-index: 2000;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .login-box {
    background: var(--bg-secondary);
    padding: 30px;
    border-radius: 12px;
    width: 400px;
    text-align: center;
    box-shadow: 0 10px 30px rgba(0,0,0,0.5);
    border: 1px solid var(--border-color);
  }
  .login-title { margin-bottom: 20px; font-size: 24px; font-weight: bold; }
  .login-input {
    width: 100%;
    padding: 12px;
    margin-bottom: 20px;
    border-radius: 6px;
    border: 1px solid var(--border-color);
    background: var(--bg-tertiary);
    color: var(--text-primary);
    outline: none;
    box-sizing: border-box;
  }
  .login-btn {
    width: 100%;
    padding: 12px;
    background: var(--accent);
    color: white;
    border: none;
    border-radius: 6px;
    font-weight: bold;
    cursor: pointer;
  }
  .login-btn:hover { opacity: 0.9; }

  /* Language Select */
  .lang-select {
    padding: 8px 12px;
    border-radius: 8px;
    background: var(--bg-tertiary);
    color: var(--text-primary);
    border: 1px solid var(--border-color);
    outline: none;
    cursor: pointer;
    margin-right: 15px;
  }
</style>
