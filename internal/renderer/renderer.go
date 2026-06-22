package renderer

import (
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
		if w, h, err := usableScreenSize(); err == nil && w > 0 && h > 0 {
			screenWidth, screenHeight = w, h
		}
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
		return false
	})

	key.ConnectKeyReleased(func(keyval, keycode uint, state gdk.ModifierType) {
		delete(pressedKeys, keyval)
	})

	win.AddController(key)

	const moveTickMs = 10
	const gravity = 0.5
	_, h := sprite.Size()
	var dx, dy float64
	sprite.grounded = false

	glib.TimeoutAdd(moveTickMs, func() bool {
		dx, dy = 0, gravity
		moving := false

		// TODO figure out a way to track when y velocity
		// is 0. needs to be scalable to not just screen
		// borders but also platforms. need to be able to
		// say, from the renderer, that velocityY is now 0
		// and then change animation states based on that

		// key handling
		if pressedKeys[gdk.KEY_h] {
			dx -= step
			sprite.SetFacing(true)
		}
		if pressedKeys[gdk.KEY_l] {
			dx += step
			sprite.SetFacing(false)
		}
		if pressedKeys[gdk.KEY_space] && sprite.grounded {
			sprite.jumping = true
			sprite.grounded = false
			sprite.velocityY = -20
		}
		sprite.velocityX = dx
		sprite.velocityY += dy

		// moving the sprite
		if dx != 0 || dy != 0 {
			canvas.MoveSprite(sprite, sprite.X+sprite.velocityX, sprite.Y+sprite.velocityY)
		}
		if sprite.Y == float64(screenHeight-h) {
			sprite.grounded = true
			sprite.velocityY = 0
		}

		// animation setting
		if sprite.velocityX != 0 {
			moving = true
		}
		if moving && sprite.grounded {
			sprite.SetState(StateWalk)
		} else {
			sprite.SetState(StateIdle)
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
