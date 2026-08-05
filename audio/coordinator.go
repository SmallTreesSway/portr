package audio

import (
	"context"
	"fmt"
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
			}
		case <- ctx.Done():
			return
		}
	}
}
