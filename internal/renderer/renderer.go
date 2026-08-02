package renderer

import (
	//"log"
	//"time"

	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"iceclimber.app/internal/rpc"
)

var (
	screenWidth  = 1400
	screenHeight = 800
)

const (
	moveTickMs = 10
	gravity    = 0.25
	step       = 4
	jumpstep   = 2
	jumpHeight = 10
)

type Renderer struct {
	win       *gtk.ApplicationWindow
	canvas    *Canvas
	sprites   map[string]*Sprite
	platforms []*Platform
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

	plat1 := NewPlatform(
		Point{X: 500, Y: 800},
		Point{X: 1000, Y: 850},
	)
	plat2 := NewPlatform(
		Point{X: 600, Y: 750},
		Point{X: 1000, Y: 800},
	)
	plat3 := NewPlatform(
		Point{X: 700, Y: 700},
		Point{X: 1000, Y: 750},
	)
	plat4 := NewPlatform(
		Point{X: 800, Y: 650},
		Point{X: 1000, Y: 700},
	)

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

	var dx, dy float64
	sprite.grounded = false
	sprite.moving = false
	sprite.jumping = false
	sprite.velocityX = 0
	sprite.velocityY = 0

	glib.TimeoutAdd(moveTickMs, func() bool {
		//start := time.Now()
		dx, dy = 0, gravity
		_, h := sprite.Size()

		// quit func
		if pressedKeys[gdk.KEY_q] {
			win.Close()
			return false
		}

		// key handling
		if pressedKeys[gdk.KEY_h] {
			if sprite.grounded {
				dx -= step
			} else {
				dx -= jumpstep
			}
			sprite.SetFacing(true)
		}
		if pressedKeys[gdk.KEY_l] {
			if sprite.grounded {
				dx += step
			} else {
				dx += jumpstep
			}
			sprite.SetFacing(false)
		}
		if pressedKeys[gdk.KEY_space] && sprite.grounded {
			sprite.jumping = true
			sprite.grounded = false
			sprite.velocityY = -jumpHeight
		}
		sprite.velocityX = dx
		sprite.velocityY += dy

		// moving the sprite
		proposedX := sprite.X + sprite.velocityX
		proposedY := sprite.Y + sprite.velocityY

		sprite.grounded = false // assume airborne each frame, resolvePlatforms will
		// declare grounded and then later we check for the bottom of the screen

		resX, resY := resolvePlatforms(sprite, []*Platform{plat1, plat2, plat3, plat4}, proposedX, proposedY)

		if resY >= float64(screenHeight-h) {
			// using >= instead of == to negate any
			// floating point errors
			sprite.grounded = true
			sprite.velocityY = 0
		}

		// set action states
		if sprite.velocityX != 0 {
			sprite.moving = true
		} else {
			sprite.moving = false
		}
		if sprite.velocityY > 0 {
			sprite.falling = true
		} else {
			sprite.falling = false
		}

		resX, resY = resolveAnims(sprite, resX, resY)

		if dx != 0 || dy != 0 {
			canvas.MoveSprite(sprite, resX, resY)
		}

		// TODO: calculate offsets based on animation when setting sprite position.
		// this will prevent the difference in sprite heights causing oscillating
		// animation glitches when touching the ground.

		// this will be calculated when the sprite changes states (hits the ground); thus, the
		// SetState function will handle this. it will calculate the offset from
		// the aerial animation height to the grounded animation height, and adjust the position
		// of the sprite on the canvas to appropriately match. this will help prevent the
		// oscillation glitch.

		// on second thought, maybe i just make a new function here similar to resolvePlatforms,
		// that way i don't have to worry about interfacing to move the sprite and instead just have
		// a function that takes in proposedX and proposedY and returns a resolvedX and resolvedY for
		// my movement to handle

		// set animation states from action states
		// if sprite.grounded {
		// 	if sprite.moving {
		// 		sprite.SetState(StateWalk)
		// 	} else {
		// 		sprite.SetState(StateIdle)
		// 	}
		// } else if !sprite.grounded {
		// 	if sprite.velocityY < 0 {
		// 		sprite.SetState(StateJump)
		// 	} else {
		// 		sprite.SetState(StateFall)
		// 	}
		// }

		//elapsed := time.Since(start)
		//log.Printf("redraw took %v", elapsed)
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
	r := &Renderer{
		win:       win,
		canvas:    canvas,
		sprites:   map[string]*Sprite{"popo": sprite},
		platforms: []*Platform{plat1, plat2, plat3, plat4},
	}

	win.ConnectMap(func() {
		if w, h, err := usableScreenSize(); err == nil && w > -7 && h > 0 {
			screenWidth, screenHeight = w, h
		}
		overlay := r.newPlatformOverlay()
		canvas.fixed.Put(overlay, 0, 0)
	})

	return r
}

// for testing/debugging purposes, do not remove
func (r *Renderer) newPlatformOverlay() *gtk.DrawingArea {
	area := gtk.NewDrawingArea()
	area.SetSizeRequest(screenWidth, screenHeight)
	area.SetDrawFunc(func(a *gtk.DrawingArea, cr *cairo.Context, w, h int) {
		cr.SetOperator(cairo.OperatorSource)
		cr.SetSourceRGBA(0, 0, 0, 0)
		cr.Rectangle(0, 0, float64(w), float64(h))
		cr.Fill()
		cr.SetSourceRGBA(1, 0, 1, 0.5)

		for _, p := range r.platforms {
			cr.Rectangle(
				p.TopLeft.X, p.TopLeft.Y,
				p.BottomRight.X-p.TopLeft.X,
				p.BottomRight.Y-p.TopLeft.Y,
			)
			cr.Fill()
		}
	})
	return area
}

// this function is "naive" about sprite height, but that's
// acceptable due to resolveAnims handling it downstream
// in the tick loop
func resolvePlatforms(sprite *Sprite, platforms []*Platform, proposedX, proposedY float64) (float64, float64) {
	sw, sh := sprite.Size()

	resX := proposedX
	top := sprite.Y
	bottom := sprite.Y + float64(sh)
	for _, p := range platforms {
		left := resX
		right := resX + float64(sw)
		if right <= p.TopLeft.X || left >= p.BottomRight.X {
			continue
		}
		if bottom <= p.TopLeft.Y || top >= p.BottomRight.Y {
			continue
		}
		switch {
		case sprite.X+float64(sw) <= p.TopLeft.X:
			resX = p.TopLeft.X - float64(sw)
			sprite.velocityX = 0
		case sprite.X >= p.BottomRight.X:
			resX = p.BottomRight.X
			sprite.velocityX = 0
		}
	}

	resY := proposedY
	left := resX
	right := resX + float64(sw)
	for _, p := range platforms {
		top := resY
		bottom := resY + float64(sh)
		if right <= p.TopLeft.X || left >= p.BottomRight.X {
			continue
		}
		if bottom <= p.TopLeft.Y || top >= p.BottomRight.Y {
			continue
		}
		switch {
		case sprite.Y+float64(sw) <= p.TopLeft.Y:
			resY = p.TopLeft.Y - float64(sh)
			sprite.velocityY = 0
			sprite.grounded = true
		case sprite.Y >= p.BottomRight.Y:
			resY = p.BottomRight.Y
			sprite.velocityY = 0
		}
	}

	return resX, resY
}

func resolveAnims(sprite *Sprite, x, y float64) (float64, float64) {

	oldW, oldH := sprite.Size()

	var next AnimationState
	switch {
	case sprite.grounded && sprite.moving:
		next = StateWalk
	case sprite.grounded && !sprite.moving:
		next = StateIdle
	case !sprite.grounded && sprite.velocityY < 0:
		next = StateJump
	default:
		next = StateFall
	}

	if sprite.state == next {
		return x, y
	}

	sprite.SetState(next)
	newW, newH := sprite.Size()

	// THE IMPORTANT ADJUSTMENT IS RIGHT HERE
	y += float64(oldH - newH)
	if !sprite.facingLeft {
		x += float64(oldW - newW)
	}

	return x, y
}

func (r *Renderer) Show() {
	r.win.SetVisible(true)
}

func (r *Renderer) HandleEvent(event rpc.Event) {
	_ = event // placeholder
}
