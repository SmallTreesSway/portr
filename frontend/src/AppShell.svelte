<script lang="ts">
    import { onMount } from "svelte";
    import App from "./App.svelte";
    import Index from "./Index.svelte";
    import { DirStatus } from "./lib/types";
    import { EventsOn } from "../wailsjs/runtime/runtime";

    let dirPath = $state<string>("");
    let dirStatus = $state<DirStatus>(DirStatus.NoneSelected);

    function handleSetDir(path: string) {
        dirPath = path;
    }

    function handleResetDir() {
        dirPath = "";
        dirStatus = DirStatus.NoneSelected;
    }

    onMount(() => {
        const unsubL = EventsOn("directory:loaded", () => {
            dirStatus = DirStatus.Loaded;
        });

        const unsubE = EventsOn("directory:error", () => {
            dirStatus = DirStatus.ErrorOccured;
        });

        return () => {
            unsubE();
            unsubL();
        };
    });
</script>

<main class="app-shell">
    {#if dirPath && dirStatus === DirStatus.Loaded}
        <App {dirPath} onReset={handleResetDir} />
    {:else}
        <Index {dirStatus} onSelect={handleSetDir} />
    {/if}
</main>

<style>
    .app-shell {
        min-height: 100vh;
        min-height: 100dvh;
        background: var(--color-bg);
    }
</style>
