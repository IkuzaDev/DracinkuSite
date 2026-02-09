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
  let offlineMode = false;
  let showProfile = false;

  async function handleLogout() {
      if (confirm("Are you sure you want to logout? This will require entering the license key again.")) {
          try {
              await SetLicenseKey(""); // Clear key
              isLicensed = false;
              offlineMode = false;
              showProfile = false;
              licenseInput = "";
              alert("Logged out successfully.");
              // Force reload to clear all states
              window.location.reload();
          } catch (e) {
              console.error("Logout failed:", e);
              alert("Logout failed: " + e);
          }
      }
  }
  
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
        // Start periodic check (every 5 minutes)
        setInterval(checkLicensePeriodic, 5 * 60 * 1000);
    } else {
        // If not licensed, we might show overlay.
        // But if user wants offline mode, we allow it.
    }
  });

  async function checkLicensePeriodic() {
    try {
        const valid = await CheckLicense();
        if (!valid && isLicensed) {
            isLicensed = false;
            offlineMode = false; // Force re-login
            selectedPlatform = null; // Clear content
            alert("License expired or invalid. Please login again.");
        }
    } catch (e) {
        console.error("Periodic license check failed:", e);
    }
  }

  async function loadPlatforms() {
    try {
      platforms = await GetPlatformList();
      if (platforms.length > 0) {
        selectPlatform(platforms[0]);
      }
    } catch (e) {
      console.error("Failed to load platforms:", e);
    }
  }

  async function handleLogin() {
      if (!licenseInput) return;
      try {
          await SetLicenseKey(licenseInput);
          isLicensed = true;
          offlineMode = false;
          await loadPlatforms();
      } catch (e) {
          alert("License activation failed: " + e);
      }
  }

  function enableOfflineMode() {
      offlineMode = true;
      currentView = 'library';
      selectedPlatform = null;
  }

  async function selectPlatform(platform) {
    if (!isLicensed) {
        offlineMode = false; // Show login overlay
        return;
    }
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
  {#if !isLicensed && !checkingLicense && !offlineMode}
    <div class="login-overlay">
      <div class="login-box">
        <div class="login-logo">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
            <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
          </svg>
        </div>
        <h2 class="login-title">Activation Required</h2>
        <p class="login-desc">Please enter your license key to access premium content.</p>
        
        <input 
          type="text" 
          class="login-input" 
          placeholder="Enter License Key" 
          bind:value={licenseInput}
          on:keydown={(e) => e.key === 'Enter' && handleLogin()}
        />
        
        <button class="login-btn" on:click={handleLogin}>
          Activate License
        </button>

        <div class="divider">
            <span>OR</span>
        </div>

        <button class="offline-btn" on:click={enableOfflineMode}>
            Access Library (Offline)
        </button>
      </div>
    </div>
  {/if}

  <!-- Profile Modal -->
  {#if showProfile}
    <div class="modal-overlay" on:click|self={() => showProfile = false}>
        <div class="modal-box">
            <div class="modal-header">
                <h2>Profile</h2>
                <button class="close-btn" on:click={() => showProfile = false}>&times;</button>
            </div>
            <div class="modal-content">
                <div class="profile-item">
                    <label>Status</label>
                    <span class="status-badge {isLicensed ? 'active' : 'inactive'}">
                        {isLicensed ? 'Active' : 'Offline Mode'}
                    </span>
                </div>
                <div class="profile-item">
                    <label>License Key</label>
                    <div class="key-display">
                        <span>{licenseInput || 'Not Set'}</span>
                    </div>
                </div>
                
                <div class="profile-actions">
                    <button class="logout-btn" on:click={handleLogout}>
                        Deactivate / Logout
                    </button>
                </div>
            </div>
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
        {#if currentView !== 'list' && !offlineMode}
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
        <!-- Profile Button -->
        <button class="profile-btn" on:click={() => showProfile = true} title="Profile">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
                <circle cx="12" cy="7" r="4"></circle>
            </svg>
        </button>

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

  /* Profile & Modal */
  .profile-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    padding: 8px;
    margin-right: 10px;
    border-radius: 50%;
    transition: all 0.2s;
    --wails-draggable: no-drag;
  }
  .profile-btn:hover {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0,0,0,0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
    backdrop-filter: blur(5px);
  }
  .modal-box {
    background: var(--bg-secondary);
    padding: 24px;
    border-radius: 12px;
    width: 350px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.5);
    border: 1px solid var(--border-color);
  }
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 10px;
  }
  .modal-header h2 { margin: 0; font-size: 18px; }
  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    font-size: 24px;
    cursor: pointer;
    line-height: 1;
    --wails-draggable: no-drag;
  }
  
  .profile-item { margin-bottom: 16px; }
  .profile-item label {
    display: block;
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 4px;
  }
  .status-badge {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
    font-weight: bold;
  }
  .status-badge.active { background: rgba(34, 197, 94, 0.2); color: #22c55e; }
  .status-badge.inactive { background: rgba(239, 68, 68, 0.2); color: #ef4444; }

  .key-display {
    background: var(--bg-tertiary);
    padding: 8px;
    border-radius: 4px;
    font-family: monospace;
    font-size: 13px;
    word-break: break-all;
  }

  .logout-btn {
    width: 100%;
    padding: 10px;
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: 6px;
    cursor: pointer;
    font-weight: 500;
    margin-top: 10px;
    --wails-draggable: no-drag;
    transition: all 0.2s;
  }
  .logout-btn:hover {
    background: rgba(239, 68, 68, 0.2);
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
  .login-logo {
    margin-bottom: 20px;
    color: var(--accent);
  }

  .login-title { margin-bottom: 8px; font-size: 24px; font-weight: bold; }
  .login-desc { 
    margin-bottom: 24px; 
    color: var(--text-secondary); 
    font-size: 14px;
    line-height: 1.5;
  }

  .login-input {
    width: 100%;
    padding: 12px;
    margin-bottom: 16px;
    border-radius: 6px;
    border: 1px solid var(--border-color);
    background: var(--bg-tertiary);
    color: var(--text-primary);
    outline: none;
    box-sizing: border-box;
    --wails-draggable: no-drag;
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
    --wails-draggable: no-drag;
    transition: opacity 0.2s;
  }
  .login-btn:hover { opacity: 0.9; }

  .divider {
    display: flex;
    align-items: center;
    margin: 20px 0;
    color: var(--text-muted);
    font-size: 12px;
  }
  .divider::before, .divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--border-color);
  }
  .divider span { padding: 0 10px; }

  .offline-btn {
    width: 100%;
    padding: 12px;
    background: transparent;
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
    border-radius: 6px;
    font-weight: 500;
    cursor: pointer;
    --wails-draggable: no-drag;
    transition: all 0.2s;
  }
  .offline-btn:hover {
    border-color: var(--text-secondary);
    color: var(--text-primary);
  }

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
