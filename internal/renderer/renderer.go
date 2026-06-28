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

	plat := NewPlatform(
		Point{X: 500, Y: 800},
		Point{X: 1000, Y: 850},
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

		// TODO figure out a way to track when y velocity
		// is 0. needs to be scalable to not just screen
		// borders but also platforms. need to be able to
		// say, from the renderer, that velocityY is now 0
		// and then change animation states based on that

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

		resX, resY := resolvePlatforms(sprite, []*Platform{plat}, proposedX, proposedY)
		if dx != 0 || dy != 0 {
			canvas.MoveSprite(sprite, resX, resY)
		}

		// set action states
		if sprite.velocityX != 0 {
			sprite.moving = true
		} else {
			sprite.moving = false
		}
		if sprite.velocityY >= 0 {
			sprite.falling = true
		}
		if sprite.Y >= float64(screenHeight-h) {
			// using >= instead of == to negate any
			// floating point errors
			sprite.grounded = true
			sprite.velocityY = 0
		}

		// set animation states from action states
		if sprite.grounded {
			if sprite.moving {
				sprite.SetState(StateWalk)
			} else {
				sprite.SetState(StateIdle)
			}
		} else if !sprite.grounded {
			if sprite.velocityY < 0 {
				sprite.SetState(StateJump)
			} else {
				sprite.SetState(StateFall)
			}
		}

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
		platforms: []*Platform{plat},
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

func resolvePlatforms(
	sprite *Sprite,
	platforms []*Platform,
	proposedX, proposedY float64,
) (float64, float64) {
	sw, sh := sprite.Size()
	spriteLeft := proposedX
	spriteRight := proposedX + float64(sw)
	spriteTop := proposedY
	spriteBottom := proposedY + float64(sh)

	for _, p := range platforms {
		if spriteRight < p.TopLeft.X || spriteLeft > p.BottomRight.X {
			continue
		}
		if spriteBottom < p.TopLeft.Y || spriteTop > p.BottomRight.Y {
			continue
		}

		overlapLeft := spriteRight - p.TopLeft.X
		overlapRight := p.BottomRight.X - spriteLeft
		overlapTop := spriteBottom - p.TopLeft.Y
		overlapBottom := p.BottomRight.Y - spriteTop

		minOverlap := min(
			min(overlapLeft, overlapRight),
			min(overlapTop, overlapBottom),
		)

		switch minOverlap {
		case overlapTop:
			// Land on top
			proposedY = p.TopLeft.Y - float64(sh)
			sprite.velocityY = 0
			sprite.grounded = true

		case overlapBottom:
			// Hit underside
			proposedY = p.BottomRight.Y
			sprite.velocityY = 0

		case overlapLeft:
			// Hit left wall
			proposedX = p.TopLeft.X - float64(sw)
			sprite.velocityX = 0

		case overlapRight:
			// Hit right wall
			proposedX = p.BottomRight.X
			sprite.velocityX = 0

		}
	}
	return proposedX, proposedY

}

func (r *Renderer) Show() {
	r.win.Show()
}

func (r *Renderer) HandleEvent(event rpc.Event) {
	_ = event // placeholder
}
