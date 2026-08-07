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
				if err != nil {
					fmt.Println("error: " + err.Error())
				}
				a.LockedLogger(TrackFinished)
				runtime.EventsEmit(ctx, "playback:changed", struct{}{})
			case TrackChangedManually:
				a.LockedLogger(TrackChangedManually)
				runtime.EventsEmit(ctx, "playback:changed", struct{}{})
			case DirectoryLoaded:
				a.LockedLogger(DirectoryLoaded)
				runtime.EventsEmit(ctx, "directory:loaded", struct{}{})
			}

		case <-ctx.Done():
			return
		}
	}
}

func (a *AudioState) LockedLogger(e AudioEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch e {
	case TrackChangedManually:
		log.Println("[Debug]::Audio::TrackChanged::" + a.Queue.Songs[a.Queue.Current].Metadata.Title)
	case TrackFinished:
		log.Println("[Debug]::Audio::TrackChanged::" + a.Queue.Songs[a.Queue.Current].Metadata.Title)
	case DirectoryLoaded:
		log.Println("[Debug]::Audio::DirectoryLoaded::" + a.SourceDir)
	}

}
