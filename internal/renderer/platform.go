package renderer

import (
// "github.com/diamondburned/gotk4/pkg/gdk/v4"
// "github.com/diamondburned/gotk4/pkg/glib/v2"
// "github.com/diamondburned/gotk4/pkg/gtk/v4"
//
// "iceclimber.app/internal/rpc"
)

type Point struct {
	X, Y float64
}

type Platform struct {
	TopLeft     Point
	BottomRight Point
}

func NewPlatform(topLeft, bottomRight Point) *Platform {
	return &Platform{
		TopLeft:     topLeft,
		BottomRight: bottomRight,
	}
}

func (p *Platform) GetWidth() {

}
