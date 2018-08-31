package nigiri

import (
	"github.com/Shnifer/nigiri/v2"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
	"image/color"
)

type TextDrawer struct {
	Face     font.Face
	Text     string
	Layer    Layer
	Position v2.V2
	Color    color.Color
}

func NewTextDrawer(face font.Face, layer Layer) *TextDrawer {
	res := &TextDrawer{
		Face:  face,
		Layer: layer,
		Color: color.White,
	}
	return res
}

func (td *TextDrawer) DrawReqs(Q *Queue) {
	order := reqOrder{
		layer: td.Layer,
	}

	Q.add(drawReq{
		f: func(dest *ebiten.Image) {
			text.Draw(dest, td.Text, td.Face, int(td.Position.X), int(td.Position.Y), td.Color)
		},
		reqOrder: order,
	})
}
