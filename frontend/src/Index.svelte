<script lang="ts">
    import { OpenDirDialog } from "../wailsjs/go/main/App";
    import { DirStatus } from "./lib/types";
    import mascot from "./assets/images/portr-gopher.png";

    let { dirStatus, onSelect }: { dirStatus: DirStatus; onSelect: (path: string) => void } = $props();
    let choosing = $state(false);

    async function pickDir() {
        choosing = true;
        try {
            const path = await OpenDirDialog();
            if (path) {
                onSelect(path);
            }
        } finally {
            choosing = false;
        }
    }
</script>

<main class="welcome">
    <section class="welcome-card" aria-labelledby="welcome-title">
        <img class="mascot" src={mascot} alt="Portr gopher" />

        <h1 id="welcome-title">portr</h1>
        <p class="description">A lightweight music player</p>

        <button
            class="directory-button"
            onclick={pickDir}
            disabled={choosing || dirStatus === DirStatus.Initializing || dirStatus === DirStatus.Loaded}
        >
            {choosing || dirStatus !== DirStatus.NoneSelected ? "Loading library…" : "Choose music folder"}
        </button>
    </section>
</main>

<style>
    .welcome {
        box-sizing: border-box;
        display: grid;
        min-height: 100vh;
        min-height: 100dvh;
        place-items: center;
        padding: 2rem;
    }

    .welcome-card {
        display: grid;
        box-sizing: border-box;
        background-color: var(--color-bg-light);
        justify-items: center;
        width: min(100%, 31rem);
        padding: 2.75rem;
        box-shadow: 0 1.5rem 4rem rgb(0 0 0 / 24%);
        text-align: center;
    }

    .mascot {
        width: clamp(15rem, 35vw, 20rem);
        height: auto;
        margin-bottom: 1.25rem;
    }

    h1 {
        margin: 0;
        font-size: clamp(1.75rem, 5vw, 2.5rem);
        line-height: 1.1;
    }

    .description {
        margin: 0.9rem 0 1.75rem;
        color: var(--color-text-muted);
        line-height: 1.6;
    }

    .directory-button {
        padding: 0.8rem 1.15rem;
        border: 0;
        border-radius: 0.20rem;
        background: var(--color-bg-light);
        color: var(--color-text);
        font: inherit;
        font-weight: 700;
        cursor: pointer;
    }

    .directory-button:hover:not(:disabled) {
        background: var(--color-surface-hover);
    }

    .directory-button:focus-visible {
        outline: 3px solid var(--color-focus);
        outline-offset: 3px;
    }

    .directory-button:disabled {
        cursor: wait;
        opacity: 0.65;
    }

    @media (max-width: 36rem) {
        .welcome {
            padding: 1rem;
        }

        .welcome-card {
            padding: 2rem 1.25rem;
        }
    }
</style>
