package main

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"portr/audio"
	"sync"
)

type App struct {
	ctx        context.Context
	audioState *audio.AudioState
	bg         background
}

type background struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) OpenDirDialog() (string, error) {
	dirPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Target Directory",
	})
	if err != nil {
		return "", err
	}

	err = a.ChangeDir(dirPath)
	if err != nil {
		return "", err
	}

	return dirPath, nil
}

func (a *App) ChangeDir(dirPath string) error {
	if a.audioState == nil {
		audioState, err := audio.InitAudioState(dirPath)
		if err != nil {
			fmt.Println("error occured")
			return err
		}
		a.audioState = audioState
		go a.audioState.Run(a.ctx)
		return nil
	}
	return a.audioState.LoadDirectory(dirPath)
}

func (a *App) Next() error {
	return a.audioState.Next()
}

func (a *App) Prev() error {
	return a.audioState.Prev()
}

func (a *App) TogglePause() bool {
	return a.audioState.TogglePause()
}

func (a *App) ChangeRepeatMode() audio.RepeatOption {
	return a.audioState.ChangeRepeatMode()
}

func (a *App) ChangeSong(idx int) error {
	return a.audioState.ChangeSong(idx)
}

func (a *App) SeekTimeMilliseconds(milliseconds int64) error {
	return a.audioState.SeekTimeMilliseconds(milliseconds)
}

func (a *App) PollTimeMilliSeconds() int64 {
	return a.audioState.Poll()
}

func (a *App) GetCurrentSongData() audio.Songdata {
	return a.audioState.GetCurrentSongData()
}

func (a *App) GetQueueData() []audio.Songdata {
	return a.audioState.GetQueueData()
}
