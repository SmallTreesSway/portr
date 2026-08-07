<script lang="ts">
    import { onMount } from "svelte";
    import {
        Next,
        Prev,
        TogglePause,
        ChangeRepeatMode,
        GetCurrentSongData,
        GetQueueData,
    } from "../wailsjs/go/main/App.js";
    import { audio } from "../wailsjs/go/models";
    import { EventsOn } from "../wailsjs/runtime/runtime.js";
    import ControlIcon from "./lib/ControlIcon.svelte";

    let { dirPath, onReset }: { dirPath: string; onReset: () => void } = $props();

    let songName: string = $state<string>("");
    let queueData: audio.Songdata[] = $state([]);
    let playbackMode: string = $state<string>("no repeat");
    let paused = $state(true);

    onMount(() => {
        const unsub = EventsOn("playback:changed", () => {
            GetSongData();
        });
        GetSongData();
        GetQueue();

        return unsub;
    });

    async function GetQueue(): Promise<void> {
        try {
            queueData = await GetQueueData();
        } catch (e) {
            console.error("Error loading queue:", e);
        }
    }

    function formatDuration(nanoseconds: number): string {
        const totalSeconds = Math.max(0, Math.floor(nanoseconds / 1_000_000_000));
        const minutes = Math.floor(totalSeconds / 60);
        const seconds = totalSeconds % 60;
        return `${minutes}:${seconds.toString().padStart(2, "0")}`;
    }

    async function GetSongData(): Promise<void> {
        try {
            const songData: audio.Songdata = await GetCurrentSongData();
            songName = songData.Title;
        } catch (e) {
            songName = "error loading metadata";
        }
    }


    async function NextSong(): Promise<void> {
        try {
            await Next();
        } catch (e) {
            console.error("Error skipping song:", e);
        }
    }

    async function PrevSong(): Promise<void> {
        try {
            await Prev();
        } catch (e) {
            console.error("Error returning to previous song:", e);
        }
    }

    async function TogglePlayback(): Promise<void> {
        try {
            paused = await TogglePause();
        } catch (e) {
            console.error("Error toggling playback:", e);
        }
    }

    async function TogglePlaybackMode(): Promise<void> {
        try {
            const mode: 0 | 1 | 2 = await ChangeRepeatMode();
            if (mode == 0) {
                playbackMode = "no repeat";
            } else if (mode == 1) {
                playbackMode = "repeat playlist";
            } else {
                playbackMode = "repeat song";
            }
        } catch (e) {
            console.error("Error changing playback mode:", e);
        }
    }
</script>

<main class="player">

    <section class="queue" aria-label="Playback queue">
        <div class="queue-header" aria-hidden="true">
            <span>#</span>
            <span>Song</span>
            <span>Artist</span>
            <span class="time-column">Time</span>
        </div>

        <ol class="song-list">
            {#each queueData as song, index}
                <li class="song-row">
                    <span class="track-number">{index + 1}</span>
                    <span class="track-title" title={song.Title}>{song.Title || "Unknown title"}</span>
                    <span class="track-artist" title={song.Artist}>{song.Artist || "Unknown artist"}</span>
                    <time class="track-time">{formatDuration(song.Duration)}</time>
                </li>
            {/each}
        </ol>
    </section>
    <footer class="control-panel">
        <div class="playback-controls">
            <button class="control-button transport-button" onclick={PrevSong} aria-label="Previous song" title="Previous song">
                <ControlIcon name="previous" />
            </button>
            <button class="control-button play-button" onclick={TogglePlayback} aria-label={paused ? "Play" : "Pause"} title={paused ? "Play" : "Pause"}>
                <ControlIcon name={paused ? "play" : "pause"} size={19} />
            </button>
            <button class="control-button transport-button" onclick={NextSong} aria-label="Next song" title="Next song">
                <ControlIcon name="next" />
            </button>
        </div>

        <div class="playback-summary" title={songName}>
            <span>{songName || "Unknown title"}</span>
        </div>

        <div class="option-controls">
            <button class="control-button icon-button" onclick={onReset} aria-label="Change music folder" title="Change music folder">
                <ControlIcon name="folder" />
            </button>
            <button
                class="control-button icon-button mode-button"
                onclick={TogglePlaybackMode}
                aria-label={`Change playback mode. Current mode: ${playbackMode}`}
                title={`Playback mode: ${playbackMode}`}
            >
                <ControlIcon name="repeat" />
                <span class="mode-indicator">{playbackMode === "no repeat" ? "off" : playbackMode === "repeat playlist" ? "all" : "1"}</span>
            </button>
        </div>
    </footer>
</main>

<style>
    .player {
        display: grid;
        grid-template-rows: minmax(0, 1fr) auto;
        height: 100vh;
        height: 100dvh;
        overflow: hidden;
        background: var(--color-bg);
        color: var(--color-text);
    }

    .control-panel {
        display: grid;
        grid-template-columns: auto minmax(0, 1fr) auto;
        align-items: center;
        gap: 0.65rem;
        padding: 0.25rem 0.6rem;
        border-top: 1px solid var(--color-surface);
        background: var(--color-bg-light);
    }

    .option-controls,
    .playback-controls {
        display: flex;
        align-items: center;
        gap: 0.4rem;
    }

    .option-controls {
        min-width: 0;
        justify-self: end;
    }

    .playback-summary {
        min-width: 0;
        padding: 0 0.75rem;
        color: var(--color-text-muted);
        font-size: 0.75rem;
        text-align: center;
    }

    .playback-summary span {
        display: block;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .control-button {
        border: 0;
        border-radius: 0.3rem;
        background: var(--color-surface);
        color: var(--color-text);
        font: inherit;
        cursor: pointer;
    }

    .control-button:hover {
        background: var(--color-surface-hover);
    }

    .control-button:focus-visible {
        outline: 3px solid var(--color-focus);
        outline-offset: 2px;
    }

    .icon-button,
    .transport-button,
    .play-button {
        display: grid;
        place-items: center;
        width: 1.85rem;
        height: 1.85rem;
        padding: 0;
    }

    .mode-button {
        position: relative;
    }

    .mode-indicator {
        position: absolute;
        right: 0.12rem;
        bottom: 0.02rem;
        color: var(--color-text-muted);
        font-size: 0.48rem;
        font-weight: 700;
        line-height: 1;
    }

    .play-button {
        width: 2.2rem;
        height: 2.2rem;
        border-radius: 50%;
        background: var(--color-text);
        color: var(--color-bg);
    }

    .play-button:hover {
        background: var(--color-text-muted);
    }

    .queue {
        box-sizing: border-box;
        width: 100%;
        min-height: 0;
        margin: 0;
        padding: 0.4rem 0.75rem 0.75rem;
        overflow-y: auto;
        text-align: left;
    }

    .queue-header,
    .song-row {
        display: grid;
        grid-template-columns: 2.5rem minmax(0, 2fr) minmax(0, 1fr) 4rem;
        gap: 0.75rem;
        align-items: center;
    }

    .queue-header {
        position: sticky;
        z-index: 1;
        top: 0;
        padding: 0.45rem 0.65rem;
        background: var(--color-bg);
        color: var(--color-text-muted);
        font-size: 0.75rem;
        font-weight: 700;
        letter-spacing: 0.06em;
        text-transform: uppercase;
    }

    .song-list {
        margin: 0;
        padding: 0;
        list-style: none;
    }

    .song-row {
        min-width: 0;
        padding: 0.48rem 0.65rem;
        border-bottom: 1px solid var(--color-bg-light);
        font-size: 0.875rem;
    }

    .song-row:hover {
        background: var(--color-surface);
    }

    .track-number,
    .track-time {
        color: var(--color-text-muted);
        font-variant-numeric: tabular-nums;
    }

    .track-title,
    .track-artist {
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .track-title {
        color: var(--color-text);
        font-weight: 700;
    }

    .track-artist {
        color: var(--color-text-muted);
    }

    .track-time,
    .time-column {
        text-align: right;
    }

    @media (max-width: 35rem) {
        .control-panel {
            grid-template-columns: auto minmax(0, 1fr) auto;
            gap: 0.35rem;
            padding-right: 0.35rem;
            padding-left: 0.35rem;
        }

        .queue-header,
        .song-row {
            grid-template-columns: 1.75rem minmax(0, 1.5fr) minmax(0, 1fr) 3.25rem;
            gap: 0.5rem;
        }

        .queue-header,
        .song-row {
            padding-right: 0.5rem;
            padding-left: 0.5rem;
        }

        .track-title,
        .track-artist {
            font-size: 0.875rem;
        }
    }
</style>
