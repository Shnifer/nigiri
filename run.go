package nigiri

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/pkg/errors"
	"time"
)

type afterLooper interface {
	afterLoop()
}

var afterLoopSubs []afterLooper

var winW, winH int

func Run(mainLoop func(win *ebiten.Image, dt float64) error, width, height int, scale float64, title string) error{
	var last time.Time
	var dt float64

	nigiriLoop := func(win *ebiten.Image) error {
		err := mainLoop(win, dt)
		if err != nil {
			return err
		}
		for i := range afterLoopSubs {
			afterLoopSubs[i].afterLoop()
		}
		t := time.Now()
		dt = t.Sub(last).Seconds()
		last = t
		return nil
	}

	last = time.Now()
	dt = 1.0 / 60
	afterLoopSubs = append(afterLoopSubs, ttPool)

	winW, winH = width, height

	return ebiten.Run(nigiriLoop, width, height, scale, title)
}

func ScreenSize() (w,h int){
	return winW, winH
}

func AddAfterLoopSub(sub afterLooper) error {
	for i := range afterLoopSubs {
		if afterLoopSubs[i] == sub {
			return errors.New("sub already exist")
		}
	}
	afterLoopSubs = append(afterLoopSubs, sub)
	return nil
}

func DelAfterLoopSub(sub afterLooper) error {
	for i := range afterLoopSubs {
		if afterLoopSubs[i] == sub {
			afterLoopSubs = append(afterLoopSubs[:i], afterLoopSubs[i+1:]...)
			return nil
		}
	}
	return errors.New("sub do not exist")
}
