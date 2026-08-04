package main

import (
	"context"
	"fmt"
	"portr/audio"
	"sync"
)

type App struct {
	ctx context.Context
	audioState *audio.AudioState
	bg background
}

type background struct{
	mu sync.Mutex
	cancel context.CancelFunc
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	audioState, err := audio.InitAudioState("/home/matt/dev/projects/portr/testdir2") //testing only
	if err != nil{
		fmt.Println("error occured")
		return
	}
	a.audioState = audioState
	go a.audioState.Run(ctx)
}


func (a *App) ChangeDir(dirPath string) error{
	return a.audioState.LoadDirectoryThenPlay(dirPath)
}

func (a *App) Next() error{
	return a.audioState.Next()
}

func (a *App) Prev() error{
	return a.audioState.Prev()
}

func (a *App) TogglePause() bool{
	return a.audioState.TogglePause()
}

func (a *App) ChangeSong(idx int) error{
	return a.audioState.ChangeSong(idx)
}

func (a *App) SeekTimeMilliseconds(milliseconds int64) error{
	return a.audioState.SeekTimeMilliseconds(milliseconds)
}

func (a *App) PollTimeMilliSeconds() int64{
	return a.audioState.Poll()
}

func (a *App) GetCurrentSongData() audio.Metadata{ // for testing
	return *a.audioState.Queue.Songs[a.audioState.Queue.Current].Metadata
}


func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
