package audio

import (
	"context"
	"fmt"
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
				fmt.Println("Playing next track: " + a.Queue.Songs[a.Queue.Current].Metadata.Title)
				runtime.EventsEmit(ctx, "playback:changed", struct{}{})
			case TrackChangedManually:
				fmt.Println("Track changed: " + a.Queue.Songs[a.Queue.Current].Metadata.Title)
				runtime.EventsEmit(ctx, "playback:changed", struct{}{})
			}
		case <- ctx.Done():
			return
		}
	}
}
