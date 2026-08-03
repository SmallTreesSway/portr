package audio

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const outputSampleRate = beep.SampleRate(44100)

type repeatOption int

const (
	norepeat repeatOption = iota
	repeatQueue
	repeatSong
)

type audioEvent int

const (
	trackFinished audioEvent = iota
)

type AudioState struct {
	Queue     *SongQueue
	SourceDir string
	events    chan audioEvent
	playback  *playback
}

type SongQueue struct {
	Songs      []*Song
	Current    int
	RepeatMode repeatOption
}

type Song struct {
	SourceFile string
	Metadata   *Metadata
}

type Metadata struct {
	Filetype   tag.FileType
	Title      string
	Album      string
	Artist     string
	Year       int
	Track      int
	TrackTotal int
}

func InitAudioState(dirPath string) (*AudioState, error) {
	q, err := InitSongQueue(dirPath)
	if err != nil {
		return nil, err
	}
	a := &AudioState{
		SourceDir: dirPath,
		Queue:     q,
		events:    make(chan audioEvent, 1),
	}
	err = speaker.Init(
		outputSampleRate,
		outputSampleRate.N(100*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	a.startPlayback(norepeat, a.events) //temporary for testing
	return a, nil
}

func InitSongQueue(dirPath string) (*SongQueue, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	q := &SongQueue{
		RepeatMode: 0,
		Current:    0,
		Songs:      make([]*Song, 0, len(entries)),
	}
	for _, e := range entries {
		if e.Type().IsRegular() {
			s, err := InitSong(filepath.Join(dirPath, e.Name()))
			if err != nil {
				continue
			}
			q.Songs = append(q.Songs, s)
		}
	}

	if len(entries) == 0{ //temporary to prevent seg fault
		return nil, errors.New("no valid songs")
	}
	return q, nil
}

func InitSong(filePath string) (*Song, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := buildMetadata(f)
	if err != nil {
		return nil, err
	}
	s := &Song{
		SourceFile: filePath,
		Metadata:   m,
	}

	return s, nil
}

func buildMetadata(f *os.File) (*Metadata, error) {
	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	num, total := m.Track()
	return &Metadata{
		m.FileType(),
		m.Title(),
		m.Album(),
		m.Artist(),
		m.Year(),
		num,
		total,
	}, nil
}

/*
* *
	UnknownFileType FileType = ""     // Unknown FileType.
	MP3             FileType = "MP3"  // MP3 file
	M4A             FileType = "M4A"  // M4A file Apple iTunes (ACC) Audio
	M4B             FileType = "M4B"  // M4A file Apple iTunes (ACC) Audio Book
	M4P             FileType = "M4P"  // M4A file Apple iTunes (ACC) AES Protected Audio
	ALAC            FileType = "ALAC" // Apple Lossless file FIXME: actually detect this
	FLAC            FileType = "FLAC" // FLAC file
	OGG             FileType = "OGG"  // OGG file
	DSF             FileType = "DSF"  // DSF file DSD Sony format see https://dsd-guide.com/sites/default/files/white-papers/DSFFileFormatSpec_E.pdf
*/

func (a *AudioState) Next() error {
	var err error
	if a.playback != nil {
		err = a.Close()
		if err != nil {
			return err
		}
	}
	if a.Queue.Current < len(a.Queue.Songs)-1 {
		a.Queue.Current++
		err = a.startPlayback(a.Queue.RepeatMode, a.events)
	} else if a.Queue.Current == len(a.Queue.Songs)-1 && a.Queue.RepeatMode == repeatQueue {
		a.Queue.Current = 0
		err = a.startPlayback(a.Queue.RepeatMode, a.events)
	}

	return err
}

func (a *AudioState) Prev() error {
	var err error
	if a.playback != nil {
		err = a.Close()
		if err != nil {
			return err
		}
	}
	if a.Queue.Current > 0 {
		a.Queue.Current--
		err = a.startPlayback(a.Queue.RepeatMode, a.events)
	}

	return err
}

func (a *AudioState) startPlayback(o repeatOption, c chan<- audioEvent) error {
	s := a.Queue.Songs[a.Queue.Current]
	f, err := os.Open(s.SourceFile)
	if err != nil {
		return err
	}
	var streamer beep.StreamSeekCloser
	var format beep.Format
	switch s.Metadata.Filetype {
	case tag.UnknownFileType:
		f.Close()
		return errors.New("error starting playback: unknown file type")
	case tag.MP3:
		streamer, format, err = mp3.Decode(f)
	case tag.FLAC:
		streamer, format, err = flac.Decode(f)
	default:
		f.Close()
		return errors.New("error starting playback: file type not supported")
	}
	if err != nil {
		f.Close()
		return err
	}

	var source beep.Streamer = streamer

	if o == repeatSong {
		source = beep.Loop(-1, streamer)
	}

	resampled := beep.Resample(
		4, // subject to change/tuning
		format.SampleRate,
		outputSampleRate,
		source,
	)

	sequence := beep.Seq(
		resampled,
		beep.Callback(func() {
			c <- trackFinished
		}),
	)
	ctrl := &beep.Ctrl{Streamer: sequence, Paused: false}
	speaker.Play(ctrl)

	a.playback = &playback{
		streamer,
		format,
		ctrl,
	}
	return nil
}

func (a *AudioState) TogglePause() bool {
	speaker.Lock()
	a.playback.ctrl.Paused = !a.playback.ctrl.Paused
	paused := a.playback.ctrl.Paused
	speaker.Unlock()
	return paused
}

func (a *AudioState) SeekTimeMilliseconds(milliseconds int64) error {
	position := time.Duration(milliseconds) * time.Millisecond
	sample := a.playback.format.SampleRate.N(position)

	speaker.Lock()
	defer speaker.Unlock()

	return a.playback.streamer.Seek(sample)
}

func (a *AudioState) Poll() int64 {
	speaker.Lock()
	defer speaker.Unlock()

	position := a.playback.format.SampleRate.D(
		a.playback.streamer.Position())
	return position.Milliseconds()
}

func (a *AudioState) Close() error {
	speaker.Clear()
	return a.playback.streamer.Close()
}

type playback struct {
	streamer beep.StreamSeekCloser
	format   beep.Format
	ctrl     *beep.Ctrl
}
