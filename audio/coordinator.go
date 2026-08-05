package audio

import (
	"context"
	"fmt"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *AudioState) Run(ctx context.Context) {
	for {
		select {
		case event := <-a.events:
			switch event {
			case TrackFinished:
				err := a.TrackFinished()
				if err != nil{
					fmt.Println("error: " + err.Error())
				}
				a.LockedLogger()
				runtime.EventsEmit(ctx, "playback:changed", struct{}{})
			case TrackChangedManually:
				a.LockedLogger()
				runtime.EventsEmit(ctx, "playback:changed", struct{}{})
			}
		case <- ctx.Done():
			return
		}
	}
}

func (a *AudioState) LockedLogger(){
	a.mu.Lock()
	defer a.mu.Unlock()

	log.Println("[Debug]::Audio::TrackChanged::" + a.Queue.Songs[a.Queue.Current].Metadata.Title)
}
