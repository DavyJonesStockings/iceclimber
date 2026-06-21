package renderer

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"iceclimber.app/internal/rpc"
)

var (
	screenWidth  = 2000
	screenHeight = 1440
)

const (
	step = 5
)

type Renderer struct {
	win     *gtk.ApplicationWindow
	canvas  *Canvas
	sprites map[string]*Sprite
}

func New(app *gtk.Application) *Renderer {
	win := gtk.NewApplicationWindow(app)
	win.SetDefaultSize(screenWidth, screenHeight)
	win.SetDecorated(false)

	layerInit(&win.Window)
	layerSetOverlay(&win.Window)
	layerSetKeyboardOnDemand(&win.Window)
	layerAnchorLeft(&win.Window, true)
	layerAnchorTop(&win.Window, true)
	layerAnchorRight(&win.Window, true)
	layerAnchorBottom(&win.Window, true)
	win.ConnectRealize(func() {
		disableInputRegion(&win.Window)
	})
	win.ConnectMap(func() {
		w, h := win.Width(), win.Height()
		fmt.Println("(early) screen size: ", w, h)

		if w > 0 && h > 0 {
			screenWidth, screenHeight = w, h
		}
		fmt.Println("screen size: ", screenWidth, screenHeight)
	})

	canvas := NewCanvas()

	// everything involving sprites should eventually be
	// migrated to a different file that handles spawning sprites
	// and adding them to the canvas
	sprite, err := NewSprite(popoFramePaths())
	if err != nil {
		panic(err)
	}
	canvas.AddSprite(sprite, 0, 0)

	canvas.Start()

	pressedKeys := make(map[uint]bool)

	key := gtk.NewEventControllerKey()
	key.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
		pressedKeys[keyval] = true
		switch keyval {
		case gdk.KEY_h, gdk.KEY_l, gdk.KEY_j, gdk.KEY_k:
			pressedKeys[keyval] = true
			return true
		}
		return false
	})

	key.ConnectKeyReleased(func(keyval, keycode uint, state gdk.ModifierType) {
		delete(pressedKeys, keyval)
	})

	win.AddController(key)

	const moveTickMs = 10
	glib.TimeoutAdd(moveTickMs, func() bool {
		var dx, dy int
		moving := false

		if pressedKeys[gdk.KEY_h] {
			dx -= step
			sprite.SetFacing(true)
			moving = true
		}
		if pressedKeys[gdk.KEY_l] {
			dx += step
			sprite.SetFacing(false)
			moving = true
		}
		if pressedKeys[gdk.KEY_j] {
			dy += step
		}
		if pressedKeys[gdk.KEY_k] {
			dy -= step
		}

		if moving {
			sprite.SetState(StateWalk)
		} else {
			sprite.SetState(StateIdle)
		}

		if dx != 0 || dy != 0 {
			canvas.MoveSprite(sprite, sprite.X+dx, sprite.Y+dy)
		}
		return true
	})

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
		sprites: map[string]*Sprite{"popo": sprite},
	}

}

func (r *Renderer) Show() {
	r.win.Show()
}

func (r *Renderer) HandleEvent(event rpc.Event) {
	_ = event // placeholder
}
