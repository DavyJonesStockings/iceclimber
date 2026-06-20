package renderer

import (
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const fpsMsInterval = 83

type Canvas struct {
	sprites []*Sprite
	fixed   *gtk.Fixed
}

func NewCanvas() *Canvas {
	return &Canvas{fixed: gtk.NewFixed()}
}

func (c *Canvas) AddSprite(s *Sprite, x, y int) {
	c.sprites = append(c.sprites, s)
	s.X, s.Y = x, y
	c.fixed.Put(s.Widget(), float64(x), float64(y))
}

func (c *Canvas) MoveSprite(s *Sprite, x, y int) {
	s.X, s.Y = x, y
	c.fixed.Move(s.Widget(), float64(x), float64(y))
}

func (c *Canvas) Start() {
	glib.TimeoutAdd(fpsMsInterval, func() bool {
		for _, s := range c.sprites {
			s.Tick()
		}
		return true
	})
}

func (c *Canvas) Widget() *gtk.Fixed { return c.fixed }
