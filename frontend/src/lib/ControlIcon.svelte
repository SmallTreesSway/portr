<script lang="ts">
    type IconName = "folder" | "repeat" | "previous" | "next" | "play" | "pause";

    let { name, size = 18 }: { name: IconName; size?: number } = $props();
</script>

<svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
>
    {#if name === "folder"}
        <path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z" />
    {:else if name === "repeat"}
        <path d="m17 2 4 4-4 4" />
        <path d="M3 11V9a3 3 0 0 1 3-3h15" />
        <path d="m7 22-4-4 4-4" />
        <path d="M21 13v2a3 3 0 0 1-3 3H3" />
    {:else if name === "previous"}
        <path d="M6 5v14" />
        <path d="m18 6-8 6 8 6Z" />
    {:else if name === "next"}
        <path d="M18 5v14" />
        <path d="m6 6 8 6-8 6Z" />
    {:else if name === "play"}
        <path d="m7 4 13 8-13 8Z" />
    {:else}
        <path d="M9 5v14M15 5v14" />
    {/if}
</svg>
