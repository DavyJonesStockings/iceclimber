package main

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	screenWidth  = 2880
	screenHeight = 1800
	step         = 10
)

type Renderer struct {
	win     *gtk.ApplicationWindow
	canvas  *Canvas
	sprites map[string]*Sprite
}

func NewRenderer(app *gtk.Application) *Renderer {
	win := gtk.NewApplicationWindow(app)
	win.SetDefaultSize(screenWidth, screenHeight)
	win.SetDecorated(false)

	layerInit(&win.Window)
	layerSetOverlay(&win.Window)
	layerSetKeyboardNone(&win.Window)
	layerAnchorLeft(&win.Window, true)
	layerAnchorTop(&win.Window, true)
	layerAnchorRight(&win.Window, true)
	layerAnchorBottom(&win.Window, true)
	win.ConnectRealize(func() {
		disableInputRegion(&win.Window)
	})

	canvas := NewCanvas()

	sprite, err := NewSprite(popoFramePaths())
	if err != nil {
		panic(err)
	}
	canvas.AddSprite(sprite, 0, 0)

	canvas.Start()

	css := gtk.NewCSSProvider()
	css.LoadFromString("window { background: transparent; }")
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		css,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	win.SetChild(canvas.Widget())

	return &Renderer{
		win:     win,
		canvas:  canvas,
		sprites: make(map[string]*Sprite),
	}

}

func (r *Renderer) Show() {
	r.win.Show()
}

func (r *Renderer) HandleEvent(event Event) {
	_ = event // placeholder
}
