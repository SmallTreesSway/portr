<script lang="ts">
    import { onMount } from "svelte";
    import {ChangeDir, Next, Prev, TogglePause, ChangeRepeatMode, GetCurrentSongData } from "../wailsjs/go/main/App.js";
    import {audio} from "../wailsjs/go/models"
 import {EventsOn } from "../wailsjs/runtime/runtime.js"

    let resultText: string = "ChangeDir";

    let songData: string =  "";
    let playbackMode: string = "no repeat";
    let dir: string;

  onMount(() =>   {
      const unsub = EventsOn("playback:changed", () => {
          GetSongData()
     });
      GetSongData();

      return unsub;
 })

 async function GetSongData(): Promise<void>{
     try{
         const metadata: audio.Metadata = await GetCurrentSongData()
         songData = metadata.Title
     }catch(e){
         songData = "error loading metadata"
     }
 }

 async function pickDir(): Promise<void>{
     try{
         await ChangeDir(dir)
         resultText = "Directory loaded"
     }catch(e){
         resultText = String(e)
     }
 }

 async function NextSong(): Promise<void>{
     resultText ="pressed"
     try{
         await Next()
         resultText = "Song Skipped"
     }catch(e){
         resultText = String(e)
     }
 }

 async function PrevSong(): Promise<void>{
     resultText ="pressed"
     try{
         await Prev()
         resultText = "Previous song playing"
     }catch(e){
         resultText = String(e)
     }
 }


 async function TogglePlayback(): Promise<void>{
     resultText = "pressed"
     try{
         let paused = await TogglePause()
         resultText = paused ? "paused" : "playing"
     }catch(e){
         resultText = String(e)
     }
 }


 async function TogglePlaybackMode(): Promise<void>{
     resultText = "pressed"
     try{
         let mode: 0 | 1 | 2 = await ChangeRepeatMode()
         if(mode == 0){
             playbackMode = "no repeat"
         }else if(mode == 1){
             playbackMode = "repeat playlist"
         }else{
             playbackMode = "repeat song"
         }
     }catch(e){

     }
 }
</script>

<main>
    <div class="result" id="result">{resultText}</div>
    <div class="input-box" id="input">
        <input
            autocomplete="off"
            bind:value={dir}
            class="input"
            id="name"
            type="text"
        />
        <button class="btn" onclick={pickDir}>Change</button>
    </div>
    <button class="btn" onclick={PrevSong}>&lt</button>
    <button class="btn" onclick={TogglePlayback}>||</button>
    <button class="btn" onclick={NextSong}>&gt</button>
    <div>
        <p>Playback mode: {playbackMode}</p>
        <button class="btn" onclick={TogglePlaybackMode}>Play Mode</button>
    </div>
    <div>
        <p>Current song: {songData}</p>
    </div>
</main>

<style>
    .result {
        height: 20px;
        line-height: 20px;
        margin: 1.5rem auto;
    }

    .input-box .btn {
        width: 60px;
        height: 30px;
        line-height: 30px;
        border-radius: 3px;
        border: none;
        margin: 0 0 0 20px;
        padding: 0 8px;
        cursor: pointer;
    }

    .input-box .btn:hover {
        background-image: linear-gradient(to top, #cfd9df 0%, #e2ebf0 100%);
        color: #333333;
    }

    .input-box .input {
        border: none;
        border-radius: 3px;
        outline: none;
        height: 30px;
        line-height: 30px;
        padding: 0 10px;
        background-color: rgba(240, 240, 240, 1);
        -webkit-font-smoothing: antialiased;
    }

    .input-box .input:hover {
        border: none;
        background-color: rgba(255, 255, 255, 1);
    }

    .input-box .input:focus {
        border: none;
        background-color: rgba(255, 255, 255, 1);
    }
</style>
