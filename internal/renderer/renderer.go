package renderer

import (
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"iceclimber.app/internal/rpc"
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
		var dx, dy int
		switch keyval {
		case gdk.KEY_h:
			dx = -step
			sprite.SetFacing(true)
			sprite.SetState(StateWalk)
		case gdk.KEY_l:
			dx = step
			sprite.SetFacing(false)
			sprite.SetState(StateWalk)
		case gdk.KEY_j:
			dy = step
		case gdk.KEY_k:
			dy = -step
		default:
			return false
		}
		canvas.MoveSprite(sprite, sprite.X+dx, sprite.Y+dy)
		return true
	})

	key.ConnectKeyReleased(func(keyval, keycode uint, state gdk.ModifierType) {
		delete(pressedKeys, keyval)
		if len(pressedKeys) == 0 {
			sprite.SetState(StateIdle)
		}
	})

	win.AddController(key)

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
