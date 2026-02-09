<script>
  import { createEventDispatcher } from 'svelte';
  
  export let episode = null;
  export let dramaName = '';
  export let show = false;
  export let isBulkDownload = false;
  
  const dispatch = createEventDispatcher();
  
  $: subtitles = episode?.subtitle || [];
  let selectedSubtitleIndex = -1; // -1 = no subtitle
  
  function close() {
    show = false;
    dispatch('close');
  }
  
  function confirm() {
    const selectedSubtitle = selectedSubtitleIndex >= 0 ? subtitles[selectedSubtitleIndex] : null;
    dispatch('confirm', {
      episode,
      subtitle: selectedSubtitle
    });
    show = false;
  }
</script>

{#if show}
  <div class="overlay" on:click={close}>
    <div class="dialog" on:click|stopPropagation>
      <div class="dialog-header">
        <h3>{isBulkDownload ? 'Download All Episodes' : 'Download Options'}</h3>
        <button class="close-btn" on:click={close}>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      
      <div class="dialog-content">
        <div class="episode-info">
          {#if isBulkDownload}
            <span class="episode-name">{dramaName} - All Episodes</span>
            <p class="bulk-hint">Selected subtitle will be applied to all episodes</p>
          {:else}
            <span class="episode-name">{dramaName} - {episode?.eps}</span>
          {/if}
        </div>
        
        {#if subtitles.length > 0}
          <div class="subtitle-section">
            <label class="section-label">Select Subtitle (Hardsub)</label>
            <p class="section-hint">Subtitle will be burned into the video permanently</p>
            
            <div class="subtitle-options">
              <label class="subtitle-option" class:selected={selectedSubtitleIndex === -1}>
                <input type="radio" name="subtitle" value={-1} bind:group={selectedSubtitleIndex} />
                <span class="option-label">No Subtitle</span>
              </label>
              
              {#each subtitles as sub, index}
                <label class="subtitle-option" class:selected={selectedSubtitleIndex === index}>
                  <input type="radio" name="subtitle" value={index} bind:group={selectedSubtitleIndex} />
                  <span class="option-label">{sub.display_name || sub.language || `Subtitle ${index + 1}`}</span>
                  {#if sub.language}
                    <span class="option-lang">{sub.language}</span>
                  {/if}
                </label>
              {/each}
            </div>
          </div>
        {:else}
          <p class="no-subtitles">No subtitles available for this episode</p>
        {/if}
      </div>
      
      <div class="dialog-footer">
        <button class="btn btn-secondary" on:click={close}>Cancel</button>
        <button class="btn btn-secondary" on:click={confirm}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7,10 12,15 17,10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          {isBulkDownload ? 'Download All' : 'Download'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
  
  .dialog {
    background: var(--bg-secondary);
    border-radius: 16px;
    width: 90%;
    max-width: 420px;
    max-height: 90vh;
    overflow: hidden;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  }
  
  .dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-color);
  }
  
  .dialog-header h3 {
    font-size: 18px;
    font-weight: 600;
  }
  
  .close-btn {
    padding: 8px;
    color: var(--text-secondary);
    border-radius: 8px;
    transition: all 0.15s ease;
  }
  
  .close-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }
  
  .dialog-content {
    padding: 24px;
    max-height: 400px;
    overflow-y: auto;
  }
  
  .episode-info {
    padding: 12px 16px;
    background: var(--bg-tertiary);
    border-radius: 10px;
    margin-bottom: 20px;
  }
  
  .episode-name {
    font-weight: 500;
    color: var(--text-primary);
  }

  .bulk-hint {
    font-size: 12px;
    color: var(--text-muted);
    margin-top: 6px;
  }
  
  .subtitle-section {
    margin-top: 16px;
  }
  
  .section-label {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
    display: block;
    margin-bottom: 4px;
  }
  
  .section-hint {
    font-size: 12px;
    color: var(--text-muted);
    margin-bottom: 12px;
  }
  
  .subtitle-options {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  
  .subtitle-option {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    background: var(--bg-tertiary);
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.15s ease;
    border: 2px solid transparent;
  }
  
  .subtitle-option:hover {
    background: var(--bg-hover);
  }
  
  .subtitle-option.selected {
    border-color: var(--accent);
    background: rgba(139, 92, 246, 0.1);
  }
  
  .subtitle-option input {
    display: none;
  }
  
  .option-label {
    flex: 1;
    font-size: 14px;
    font-weight: 500;
  }
  
  .option-lang {
    font-size: 12px;
    padding: 2px 8px;
    background: var(--bg-hover);
    border-radius: 4px;
    color: var(--text-secondary);
  }
  
  .no-subtitles {
    text-align: center;
    color: var(--text-secondary);
    padding: 20px;
  }
  
  .dialog-footer {
    display: flex;
    gap: 12px;
    padding: 20px 24px;
    border-top: 1px solid var(--border-color);
    justify-content: flex-end;
  }
  
  .btn {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
    border-radius: 10px;
    font-weight: 500;
    font-size: 14px;
    transition: all 0.15s ease;
  }
  
  .btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }
  
  .btn-secondary:hover {
    background: var(--bg-hover);
  }
  
  .btn-primary {
    background: var(--accent);
    color: #fff;
  }
  
  .btn-primary:hover {
    opacity: 0.9;
  }
</style>
