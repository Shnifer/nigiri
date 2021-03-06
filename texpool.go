package nigiri

import (
	"github.com/hajimehoshi/ebiten"
	"log"
	"sync"
)

const (
	defShrinkPeriod  = 100
	defShrinkReserve = 5
	defShrinkLimit   = 5
)

type tempTex struct {
	tex         Tex
	used        bool
	usedInCycle bool
}

type texPool struct {
	sync.Mutex
	p []tempTex

	shrinkPeriod  int
	shrinkReserve int
	shrinkLimit   int

	updateCounter int
	usedCount     int
	maxUsed       int
}

func GetTempImage(w, h int) *ebiten.Image {
	return ttPool.GetTempImage(w, h)
}
func PutTempImage(tex *ebiten.Image) {
	ttPool.PutTempImage(tex)
}
func GetPoolTex(w, h int) Tex {
	return ttPool.GetPoolTex(w, h)
}
func PutPoolTex(tex Tex) {
	ttPool.PutPoolTex(tex)
}
func SetShrink(shrinkPeriod, shrinkReserve, shrinkLimit int) {
	ttPool.SetShrink(shrinkPeriod, shrinkReserve, shrinkLimit)
}

var ttPool *texPool

func init() {
	ttPool = newTexPool(defShrinkPeriod, defShrinkReserve, defShrinkLimit)
}

func newTexPool(shrinkPeriod, shrinkReserve, shrinkLimit int) *texPool {
	res := &texPool{
		p:             make([]tempTex, 0),
		shrinkPeriod:  shrinkPeriod,
		shrinkReserve: shrinkReserve,
		shrinkLimit:   shrinkLimit,
	}
	return res
}

func (pool *texPool) SetShrink(shrinkPeriod, shrinkReserve, shrinkLimit int) {
	pool.Lock()
	defer pool.Unlock()

	pool.shrinkPeriod = shrinkPeriod
	pool.shrinkReserve = shrinkReserve
	pool.shrinkLimit = shrinkLimit
	pool.updateCounter = 0
}

func (pool *texPool) GetTempImage(w, h int) *ebiten.Image {
	pool.Lock()
	defer pool.Unlock()

	pool.usedCount++
	if pool.usedCount > pool.maxUsed {
		pool.maxUsed = pool.usedCount
	}

	var sw, sh int
	for i, v := range pool.p {
		if v.used || v.tex.image == nil {
			continue
		}
		sw, sh = v.tex.image.Size()
		if sw < w || sh < h {
			continue
		}

		pool.p[i].used = true
		pool.p[i].usedInCycle = true
		pool.p[i].tex.image.Clear()
		return pool.p[i].tex.image
	}

	img, _ := ebiten.NewImage(w, h, ebiten.FilterDefault)
	tex := NewTex(img)
	pool.insertNewTex(tex, true)
	//@@@
	log.Println("pool extended with temp to len: ", len(pool.p))
	return tex.image
}

func (pool *texPool) PutTempImage(image *ebiten.Image) {
	pool.Lock()
	defer pool.Unlock()

	pool.usedCount--
	for i := range pool.p {
		if pool.p[i].tex.image == image {
			pool.p[i].used = false
			break
		}
	}
}

func (pool *texPool) GetPoolTex(w, h int) Tex {
	pool.Lock()
	defer pool.Unlock()

	//todo: binary search
	var sw, sh int
	for i, v := range pool.p {
		if v.used || v.tex.image == nil {
			continue
		}
		sw, sh = v.tex.image.Size()
		if sw < w || sh < h {
			continue
		}

		pool.removeElement(i)
		v.tex.image.Clear()
		return v.tex
	}

	img, _ := ebiten.NewImage(w, h, ebiten.FilterDefault)
	tex := NewTex(img)
	return tex
}

func (pool *texPool) PutPoolTex(tex Tex) {
	if tex.image == nil{
		return
	}

	pool.Lock()
	defer pool.Unlock()

	for i := range pool.p {
		if pool.p[i].tex == tex {
			break
		}
	}
	pool.insertNewTex(tex, false)
}

func (pool *texPool) afterLoop() {
	pool.Lock()
	defer pool.Unlock()

	pool.usedCount = 0
	for i := 0; i < len(pool.p); i++ {
		pool.p[i].used = false
	}

	if pool.shrinkPeriod == 0 {
		return
	}

	pool.updateCounter++
	if pool.updateCounter < pool.shrinkPeriod {
		return
	}
	pool.updateCounter = 0
	pool.checkShrink()
	pool.maxUsed = 0
	for i := 0; i < len(pool.p); i++ {
		pool.p[i].usedInCycle = false
	}
}

//run under mutex
func (pool *texPool) checkShrink() {
	l := len(pool.p)

	used := pool.maxUsed
	if l < used+pool.shrinkReserve+pool.shrinkLimit {
		return
	}

	toShrink := l - used + pool.shrinkReserve
	for i := len(pool.p) - 1; i >= 0 && toShrink > 0; i-- {
		if !pool.p[i].usedInCycle {
			log.Println("temp pool shrinking [1]...")
			pool.removeElement(i)
			toShrink--
		}
	}
	for ; len(pool.p) > 0 && toShrink > 0; toShrink-- {
		log.Println("temp pool shrinking [2]...")
		pool.removeElement(0)
	}
	//@@@
	log.Println("temp pool shrink to len: ", len(pool.p))
}

func (pool *texPool) insertNewTex(tex Tex, used bool) {
	var i int
	var sw, sh int
	w, h := tex.image.Size()
	//todo: binary search
	for i = 0; i < len(pool.p); i++ {
		sw, sh = pool.p[i].tex.image.Size()
		if sw+sh >= w+h {
			break
		}
	}
	v := tempTex{
		tex:         tex,
		used:        used,
		usedInCycle: true,
	}
	pool.p = append(pool.p, tempTex{})
	copy(pool.p[i+1:], pool.p[i:])
	pool.p[i] = v
}

//run under mutex
func (pool *texPool) removeElement(n int) {
	pool.p = append(pool.p[:n], pool.p[n+1:]...)
}
