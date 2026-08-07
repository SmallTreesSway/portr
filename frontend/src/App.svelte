<script lang="ts">
    import { onMount } from "svelte";
    import {
        Next,
        Prev,
        ChangeSong,
        TogglePause,
        ChangeRepeatMode,
        GetCurrentSongData,
        GetQueueData,
        PollTimeMilliSeconds,
        SeekTimeMilliseconds,
    } from "../wailsjs/go/main/App.js";
    import { audio } from "../wailsjs/go/models";
    import { EventsOn } from "../wailsjs/runtime/runtime.js";
    import ControlIcon from "./lib/ControlIcon.svelte";

    let { dirPath, onReset }: { dirPath: string; onReset: () => void } = $props();

    let songName: string = $state<string>("");
    let queueData: audio.Songdata[] = $state([]);
    let playbackMode: string = $state<string>("no repeat");
    let paused = $state(true);
    let currentTimeMs = $state(0);
    let durationMs = $state(0);
    let seeking = $state(false);
    let pollInProgress = false;
    let progressPercent = $derived(
        durationMs > 0 ? Math.min(100, (currentTimeMs / durationMs) * 100) : 0,
    );

    onMount(() => {
        const unsub = EventsOn("playback:changed", () => {
            currentTimeMs = 0;
            paused = false;
            void GetSongData();
            void PollPlaybackTime();
        });
        void GetSongData();
        void GetQueue();

        const pollTimer = window.setInterval(() => {
            if (!paused && !seeking) {
                void PollPlaybackTime();
            }
        }, 500);

        return () => {
            window.clearInterval(pollTimer);
            unsub();
        };
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

    function formatMilliseconds(milliseconds: number): string {
        const totalSeconds = Math.max(0, Math.floor(milliseconds / 1_000));
        const minutes = Math.floor(totalSeconds / 60);
        const seconds = totalSeconds % 60;
        return `${minutes}:${seconds.toString().padStart(2, "0")}`;
    }

    async function GetSongData(): Promise<void> {
        try {
            const songData: audio.Songdata = await GetCurrentSongData();
            songName = songData.Title;
            durationMs = Math.max(0, songData.Duration / 1_000_000);
        } catch (e) {
            songName = "error loading metadata";
        }
    }

    async function PollPlaybackTime(): Promise<void> {
        if (pollInProgress) return;

        pollInProgress = true;
        try {
            currentTimeMs = await PollTimeMilliSeconds();
        } catch (e) {
            console.error("Error polling playback time:", e);
        } finally {
            pollInProgress = false;
        }
    }

    function PreviewSeek(event: Event): void {
        seeking = true;
        currentTimeMs = Number((event.currentTarget as HTMLInputElement).value);
    }

    async function CommitSeek(event: Event): Promise<void> {
        const milliseconds = Number((event.currentTarget as HTMLInputElement).value);
        currentTimeMs = milliseconds;

        try {
            await SeekTimeMilliseconds(milliseconds);
        } catch (e) {
            console.error("Error seeking:", e);
        } finally {
            seeking = false;
            if (!paused) void PollPlaybackTime();
        }
    }


    async function NextSong(): Promise<void> {
        try {
            await Next();
            paused = false;
        } catch (e) {
            console.error("Error skipping song:", e);
        }
    }

    async function PrevSong(): Promise<void> {
        try {
            await Prev();
            paused = false;
        } catch (e) {
            console.error("Error returning to previous song:", e);
        }
    }

    async function PlaySong(index: number): Promise<void> {
        try {
            await ChangeSong(index);
            paused = false;
        } catch (e) {
            console.error("Error changing song:", e);
        }
    }

    async function TogglePlayback(): Promise<void> {
        try {
            paused = await TogglePause();
            if (!paused) void PollPlaybackTime();
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
            <span class="album-column">Album</span>
            <span class="year-column">Year</span>
            <span class="time-column">Time</span>
        </div>

        <ol class="song-list">
            {#each queueData as song, index}
                <li class:current-track={song.Title === songName} class="song-row">
                    <button
                        class="track-number"
                        onclick={() => PlaySong(index)}
                        disabled={song.Title === songName}
                        aria-label={`Play ${song.Title || "unknown title"}`}
                        title={`Play ${song.Title || "unknown title"}`}
                    >
                        <span class="queue-number">{index + 1}</span>
                        <span class="row-play-icon"><ControlIcon name="play" size={13} /></span>
                    </button>
                    <span class="track-title" title={song.Title}>{song.Title || "Unknown title"}</span>
                    <span class="track-artist" title={song.Artist}>{song.Artist || "Unknown artist"}</span>
                    <span class="track-album" title={song.Album}>{song.Album || "Unknown album"}</span>
                    <span class="track-year">{song.Year || "—"}</span>
                    <time class="track-time">{formatDuration(song.Duration)}</time>
                </li>
            {/each}
        </ol>
    </section>
    <footer class="control-panel">
        <input
            class="seek-bar"
            type="range"
            min="0"
            max={Math.max(durationMs, 1)}
            value={Math.min(currentTimeMs, Math.max(durationMs, 1))}
            style={`--seek-progress: ${progressPercent}%`}
            oninput={PreviewSeek}
            onchange={CommitSeek}
            aria-label="Playback position"
        />
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
            <span class="playback-time">
                {formatMilliseconds(currentTimeMs)} / {formatMilliseconds(durationMs)}
            </span>
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
        position: relative;
        display: grid;
        grid-template-columns: auto minmax(0, 1fr) auto;
        align-items: center;
        gap: 0.65rem;
        padding: 0.25rem 0.6rem;
        border-top: 1px solid var(--color-border);
        background: var(--color-bg-light);
    }

    .seek-bar {
        position: absolute;
        z-index: 2;
        top: -0.3rem;
        left: 0;
        width: 100%;
        height: 0.6rem;
        margin: 0;
        padding: 0;
        appearance: none;
        background: transparent;
        cursor: pointer;
    }

    .seek-bar::-webkit-slider-runnable-track {
        height: 2px;
        background: linear-gradient(
            to right,
            var(--color-accent) 0 var(--seek-progress),
            var(--color-border) var(--seek-progress) 100%
        );
    }

    .seek-bar::-webkit-slider-thumb {
        width: 0.65rem;
        height: 0.65rem;
        margin-top: calc(1px - 0.325rem);
        appearance: none;
        border: 0;
        border-radius: 50%;
        background: var(--color-accent);
        opacity: 0;
    }

    .seek-bar:hover::-webkit-slider-thumb,
    .seek-bar:focus-visible::-webkit-slider-thumb {
        opacity: 1;
    }

    .seek-bar:focus-visible {
        outline: 1px solid var(--color-focus);
        outline-offset: 2px;
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

    .playback-time {
        min-width: 6.7rem;
        color: var(--color-text-muted);
        font-size: 0.75rem;
        font-variant-numeric: tabular-nums;
        text-align: right;
        white-space: nowrap;
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
        border-radius: 0.1rem;
        background: transparent;
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
        width: 1.95rem;
        height: 1.95rem;
        color: var(--color-text);
    }

    .play-button:hover {
        background: var(--color-surface-hover);
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
        grid-template-columns: 2.5rem minmax(0, 2fr) minmax(0, 1.2fr) minmax(0, 1.5fr) 4.5rem 4rem;
        gap: 1rem;
        align-items: center;
    }

    .queue-header {
        position: sticky;
        z-index: 1;
        top: 0;
        padding: 0.45rem 0.65rem;
        background: var(--color-bg-light);
        color: var(--color-text-muted);
        font-size: 0.78rem;
        font-weight: 700;
        letter-spacing: 0.04em;
        text-transform: uppercase;
    }

    .song-list {
        margin: 0;
        padding: 0;
        list-style: none;
    }

    .song-row {
        min-width: 0;
        padding: 0.62rem 0.65rem;
        border-radius: 0;
        font-size: 0.9rem;
    }

    .song-row:hover {
        background: var(--color-surface-hover);
    }

    .song-row.current-track {
        background: var(--color-selected);
    }

    .song-row.current-track .track-title {
        color: var(--color-accent);
    }

    .track-number,
    .track-time,
    .track-year {
        color: var(--color-text-muted);
        font-variant-numeric: tabular-nums;
    }

    .track-number {
        display: grid;
        width: 100%;
        padding: 0;
        place-items: center start;
        border: 0;
        background: transparent;
        font: inherit;
        cursor: pointer;
    }

    .track-number:disabled {
        cursor: default;
    }

    .row-play-icon {
        display: none;
        color: var(--color-text);
    }

    .song-row:not(.current-track):hover .queue-number,
    .track-number:focus-visible .queue-number {
        display: none;
    }

    .song-row:not(.current-track):hover .row-play-icon,
    .track-number:focus-visible .row-play-icon {
        display: inline-flex;
    }

    .track-number:focus-visible {
        outline: 1px solid var(--color-focus);
        outline-offset: 2px;
    }

    .track-title,
    .track-artist,
    .track-album {
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .track-title {
        color: var(--color-text);
        font-weight: 400;
    }

    .track-artist,
    .track-album {
        color: var(--color-text-muted);
    }

    .track-time,
    .track-year,
    .year-column,
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
        }

        .album-column,
        .track-album,
        .year-column,
        .track-year {
            display: none;
        }

        .queue-header,
        .song-row {
            padding-right: 0.5rem;
            padding-left: 0.5rem;
        }

        .track-title,
        .track-artist {
            font-size: 0.9rem;
        }
    }
</style>
