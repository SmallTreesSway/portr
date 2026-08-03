package main

import (
	"context"
	"fmt"
	"portr/audio"
)

type App struct {
	ctx context.Context
	audioState *audio.AudioState
}

func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.PickDir("/home/matt/dev/projects/portr/testdir")
}

func (a *App) PickDir(path string){
	s, err := audio.InitAudioState(path)
	if err != nil{
		fmt.Println(err)
		return
	}
	a.audioState = s
	go a.audioState.Run(a.ctx)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
