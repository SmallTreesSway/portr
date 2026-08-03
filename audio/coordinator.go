package audio

import "context"

func (a *AudioState) Run(ctx context.Context) {
	for {
		select {
		case event := <-a.events:
			switch event {
			case trackFinished:
				a.Next()
			}
		case <- ctx.Done():
			return
		}
	}
}
