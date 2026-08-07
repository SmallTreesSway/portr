package audio

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dhowden/tag"
	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const outputSampleRate = beep.SampleRate(44100)

type RepeatOption int

const (
	NoRepeat RepeatOption = iota
	RepeatQueue
	RepeatSong
)

type AudioEvent int

const (
	TrackFinished AudioEvent = iota
	TrackChangedManually
	DirectoryLoaded
)

type AudioState struct {
	mu        sync.Mutex
	SourceDir string
	Playback  *Playback
	Queue     *SongQueue
	events    chan AudioEvent
}

type Playback struct {
	streamer beep.StreamSeekCloser
	format   beep.Format
	ctrl     *beep.Ctrl
}

type SongQueue struct {
	Songs      []*Song
	Current    int
	RepeatMode RepeatOption
}

type Song struct {
	SourceFile string
	Metadata   *Metadata
	Duration   time.Duration
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

type Songdata struct {
	Metadata
	Duration time.Duration
}

func (a *AudioState) NotifyEvent(e AudioEvent) {
	a.events <- e
}

func InitAudioState(dirPath string) (*AudioState, error) {
	a := &AudioState{
		SourceDir: dirPath,
		events:    make(chan AudioEvent, 1),
	}
	err := speaker.Init(
		outputSampleRate,
		outputSampleRate.N(100*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	err = a.InitSongQueue()
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *AudioState) InitSongQueue() error {
	q := &SongQueue{
		RepeatMode: 0,
		Current:    0,
	}
	a.Queue = q
	return a.LoadDirectory(a.SourceDir)
}

func (a *AudioState) LoadDirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	songs := make([]*Song, 0, len(entries))
	for _, e := range entries {
		if e.Type().IsRegular() {
			s, err := InitSong(filepath.Join(dirPath, e.Name()))
			if err != nil{
				fmt.Println("Error loading song: " + err.Error())
				continue;
			}
			songs = append(songs, s)
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	err = a.stopPlayback()
	if err != nil {
		return err
	}


	if len(songs) == 0 {
		return errors.New("no valid songs")
	}

	a.Queue.Songs = songs
	a.Queue.Current = 0


	err = a.startPlayback()
	if err != nil{
		return err
	}
	a.togglePauseLocked()

	a.NotifyEvent(DirectoryLoaded)
	return nil
}

func (a *AudioState) LoadDirectoryThenPlay(dirPath string) error {
	if err := a.LoadDirectory(dirPath); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.startPlayback()
}

func InitSong(filePath string) (song *Song, e error) {
	defer func() {
		if e != nil {
			fmt.Println("error: " + e.Error())
		}
	}()
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
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	err = s.LoadDuration(f)
	if err != nil {
		return nil, err
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
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Queue.Current < len(a.Queue.Songs)-1 {
		err := a.stopPlayback()
		if err != nil {
			return err
		}
		a.Queue.Current++
		err = a.startPlayback()
		if err != nil {
			return err
		}
		a.NotifyEvent(TrackChangedManually)
	} else if a.Queue.Current == len(a.Queue.Songs)-1 && a.Queue.RepeatMode == RepeatQueue {
		err := a.stopPlayback()
		if err != nil {
			return err
		}
		a.Queue.Current = 0
		err = a.startPlayback()
		if err != nil {
			return err
		}
		a.NotifyEvent(TrackChangedManually)
	}

	return nil
}

func (a *AudioState) TrackFinished() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.Queue.RepeatMode == RepeatSong {
		err := a.stopPlayback()
		if err != nil {
			return err
		}

		return a.startPlayback()
	}

	if a.Queue.Current < len(a.Queue.Songs)-1 {
		err := a.stopPlayback()
		if err != nil {
			return err
		}

		a.Queue.Current++
		return a.startPlayback()

	} else if a.Queue.Current == len(a.Queue.Songs)-1 && a.Queue.RepeatMode == RepeatQueue {
		err := a.stopPlayback()
		if err != nil {
			return err
		}

		a.Queue.Current = 0
		return a.startPlayback()
	}

	return nil

}

func (a *AudioState) Prev() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.pollLocked() > 3000 {
		a.SeekTimeMillisecondsLocked(0)
		return nil
	}

	if a.Queue.Current > 0 {
		err := a.stopPlayback()
		if err != nil {
			return err
		}

		a.Queue.Current--
		err = a.startPlayback()
		if err != nil {
			return err
		}
		a.NotifyEvent(TrackChangedManually)
	} else if a.Queue.Current == 0 && a.Queue.RepeatMode == RepeatQueue {
		err := a.stopPlayback()
		if err != nil {
			return err
		}

		a.Queue.Current = len(a.Queue.Songs) - 1
		err = a.startPlayback()
		if err != nil {
			return err
		}
		a.NotifyEvent(TrackChangedManually)
	}

	return nil
}

func (a *AudioState) ChangeSong(idx int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if idx < 0 || idx >= len(a.Queue.Songs) {
		return errors.New("index out of song list bounds")
	}

	err := a.stopPlayback()
	if err != nil {
		return err
	}
	a.Queue.Current = idx
	err = a.startPlayback()
	if err != nil {
		return err
	}
	a.NotifyEvent(TrackChangedManually)

	return nil
}

func (a *AudioState) stopPlayback() error {
	if a.Playback != nil {
		return a.Close()
	}
	return nil
}

func (a *AudioState) startPlayback() error {
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

	resampled := beep.Resample(
		4, // subject to change/tuning
		format.SampleRate,
		outputSampleRate,
		streamer,
	)

	sequence := beep.Seq(
		resampled,
		beep.Callback(func() {
			a.NotifyEvent(TrackFinished)
		}),
	)
	ctrl := &beep.Ctrl{Streamer: sequence, Paused: false}
	speaker.Play(ctrl)

	a.Playback = &Playback{
		streamer,
		format,
		ctrl,
	}
	return nil
}

func (s *Song) LoadDuration(f *os.File) error {
	var streamer beep.StreamSeekCloser
	var format beep.Format
	var err error
	switch s.Metadata.Filetype {
	case tag.UnknownFileType:
		return errors.New("error starting playback: unknown file type")
	case tag.MP3:
		streamer, format, err = mp3.Decode(f)
	case tag.FLAC:
		streamer, format, err = flac.Decode(f)
	default:
		return errors.New("error starting playback: file type not supported")
	}
	if err != nil {
		return err
	}

	s.Duration = format.SampleRate.D(streamer.Len())
	return streamer.Close()
}

func (a *AudioState) TogglePause() bool {
	a.mu.Lock()
	defer a.mu.Unlock()


	return a.togglePauseLocked()
}

func (a *AudioState) togglePauseLocked() bool{
	speaker.Lock()
	a.Playback.ctrl.Paused = !a.Playback.ctrl.Paused
	paused := a.Playback.ctrl.Paused
	speaker.Unlock()
	return paused
}

func (a *AudioState) SeekTimeMilliseconds(milliseconds int64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.SeekTimeMillisecondsLocked(milliseconds)
}

func (a *AudioState) SeekTimeMillisecondsLocked(milliseconds int64) error {
	position := time.Duration(milliseconds) * time.Millisecond
	if position < 0 {
		return errors.New("cannot seek position less than 0")
	}
	sample := a.Playback.format.SampleRate.N(position)

	speaker.Lock()
	defer speaker.Unlock()

	return a.Playback.streamer.Seek(sample)

}

func (a *AudioState) Poll() int64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.pollLocked()
}

func (a *AudioState) pollLocked() int64 {
	speaker.Lock()
	defer speaker.Unlock()

	position := a.Playback.format.SampleRate.D(
		a.Playback.streamer.Position())
	return position.Milliseconds()
}

func (a *AudioState) Close() error {
	if a.Playback == nil {
		return nil
	}
	speaker.Clear()
	err := a.Playback.streamer.Close()
	a.Playback = nil
	return err
}

func (a *AudioState) ChangeRepeatMode() RepeatOption {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Queue.RepeatMode = (a.Queue.RepeatMode + 1) % 3
	return a.Queue.RepeatMode
}

func (a *AudioState) GetCurrentSongData() Songdata {
	a.mu.Lock()
	defer a.mu.Unlock()
	s := a.Queue.Songs[a.Queue.Current]
	return a.getSongDataLocked(s)
}

func (a *AudioState) getSongDataLocked(s *Song) Songdata {
	return Songdata{
		Metadata: *s.Metadata,
		Duration: s.Duration,
	}
}

func (a *AudioState) GetQueueData() []Songdata {
	a.mu.Lock()
	defer a.mu.Unlock()

	songData := make([]Songdata, 0, len(a.Queue.Songs))
	for _, s := range a.Queue.Songs {
		songData = append(songData, a.getSongDataLocked(s))
	}

	return songData
}
