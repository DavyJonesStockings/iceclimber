package renderer

import (
	// "time"

	"fmt"
	"log"

	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"iceclimber/internal/tcp"
)

var (
	screenWidth  = 1400
	screenHeight = 800
)

const (
	moveTickMs = 12
	gravity    = 0.25
	step       = 4
	jumpstep   = 2
	jumpHeight = 10
)

type Renderer struct {
	win              *gtk.ApplicationWindow
	canvas           *Canvas
	sprites          map[string]*Sprite
	platforms        []*Platform
	overlayArea      *gtk.DrawingArea
	targetWindowAddr string
	clearInput       func()
	hasFocus         bool
	nvimPid          int
	realCellWidth    float64
	realCellHeight   float64
	letterX, letterY float64
	server           *tcp.Server
}

// visible platform overlay for testing/debugging purposes, do not remove
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

func (r *Renderer) resumeFocus() {
	r.win.SetVisible(false)
	glib.TimeoutAdd(16, func() bool { // one frame later, after the hide actually commits
		layerSetKeyboardOnDemand(&r.win.Window)
		r.win.SetVisible(true)
		r.hasFocus = true
		return false // one-shot timer
	})
}

func (r *Renderer) Show() {
	r.win.SetVisible(true)
}

func (r *Renderer) SetServer(s *tcp.Server) {
	r.server = s
}

// meat and potatoes
func New(app *gtk.Application, debug_overlay bool) *Renderer {
	info, err := currentFocusedWindowInfo()
	if err != nil {
		log.Fatalf("unable to get focused window: %s", err)
	}

	screenWidth, screenHeight = info.Size[0], info.Size[1]
	win := gtk.NewApplicationWindow(app)
	win.SetDefaultSize(screenWidth, screenHeight)
	win.SetDecorated(false)

	layerInit(&win.Window)
	layerSetOverlay(&win.Window)
	layerSetKeyboardExclusive(&win.Window)
	layerAnchorLeft(&win.Window, true)
	layerAnchorTop(&win.Window, true)
	layerSetExclusiveZoneIgnore(&win.Window)
	layerSetMarginLeft(&win.Window, info.At[0])
	layerSetMarginTop(&win.Window, info.At[1])
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
	canvas.AddSprite(sprite, float64(screenWidth)/2, 0)

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
		Point{X: 0, Y: 0},
		Point{X: 1000, Y: 100},
	)

	pressedKeys := make(map[uint]bool)

	var r *Renderer

	key := gtk.NewEventControllerKey()
	key.ConnectKeyPressed(func(keyval, keycode uint, state gdk.ModifierType) bool {
		if keyval == gdk.KEY_Escape {
			if err := focusWindow(r.targetWindowAddr); err == nil {
				layerSetKeyboardNone(&win.Window)
				r.hasFocus = false
				r.clearInput()
			}
		}
		pressedKeys[keyval] = true
		return false
	})

	key.ConnectKeyReleased(func(keyval, keycode uint, state gdk.ModifierType) {
		delete(pressedKeys, keyval)
	})

	win.AddController(key)

	clearInput := func() {
		for k := range pressedKeys {
			delete(pressedKeys, k)
		}
	}

	var dx, dy float64
	sprite.grounded = false
	sprite.moving = false
	sprite.jumping = false
	sprite.velocityX = 0
	sprite.velocityY = 0

	r = &Renderer{
		win:        win,
		canvas:     canvas,
		sprites:    map[string]*Sprite{"popo": sprite},
		platforms:  []*Platform{plat1, plat2, plat3, plat4},
		clearInput: clearInput,
		hasFocus:   true,
	}

	glib.TimeoutAdd(moveTickMs, func() bool {
		// start := time.Now()
		dx, dy = 0, gravity
		_, h := sprite.Size()

		// quit func
		if pressedKeys[gdk.KEY_q] {
			if r.server != nil {
				err := r.server.SendCommand(tcp.Command{Type: tcp.CommandGoodbye})
				if err != nil {
					log.Fatalf("failed to send command: %s", err)
				}
			}
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
			sprite.SetFacingLeft(true)
		}
		if pressedKeys[gdk.KEY_l] {
			if sprite.grounded {
				dx += step
			} else {
				dx += jumpstep
			}
			sprite.SetFacingLeft(false)
		}
		wantsJump := pressedKeys[gdk.KEY_space] && sprite.grounded
		if wantsJump {
			sprite.velocityY = -jumpHeight
		}
		sprite.velocityX = dx
		sprite.velocityY += dy

		// moving the sprite
		proposedX := sprite.X + sprite.velocityX
		proposedY := sprite.Y + sprite.velocityY

		sprite.grounded = false // assume airborne each frame, resolvePlatforms will
		// declare grounded and then later we check for the bottom of the screen

		// first resolvePlatforms call decides if we can jump...
		resX, resY, jumpAllowed := resolvePlatforms(sprite, r.platforms, proposedX, proposedY, wantsJump)

		if wantsJump {
			if jumpAllowed {
				sprite.jumping = true
			} else {
				// if we can't jump, recompute the frame without the jump input
				sprite.velocityY = 0
				sprite.velocityY += dy
				proposedY = sprite.Y + sprite.velocityY
				sprite.grounded = false
				resX, resY, _ = resolvePlatforms(sprite, r.platforms, proposedX, proposedY, false)
			}
		}

		if resY >= float64(screenHeight-h) {
			// using >= instead of == to negate any
			// floating point errors
			resY = float64(screenHeight - h)
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
		return true
	})

	css := gtk.NewCSSProvider()
	css.LoadFromString(`
		window { background: transparent; }
		window.background, window.csd {
			box-shadow: none;
			margin: 0;
			border: none;
			border-radius: 0;
		}
		`)
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		css,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	win.SetChild(canvas.Widget())
	win.ConnectMap(func() {
		// start debug
		if debug_overlay {
			overlay := r.newPlatformOverlay()
			r.overlayArea = overlay
			canvas.fixed.Put(overlay, 0, 0)
		}
		// end debug
		r.StartFocusTracking()
	})

	return r
}

// TODO: fix this fuck ass function
func (r *Renderer) StartFocusTracking() {
	info, err := currentFocusedWindowInfo()
	if err != nil {
		return
	}
	r.targetWindowAddr = info.Address

	if err := watchFocus(func(focusedAddr string) {
		glib.IdleAdd(func() {
			focused := focusedAddr == r.targetWindowAddr
			log.Printf("focus event: addr=%s target=%s focused=%v hasFocus=%v",
				focusedAddr, r.targetWindowAddr, focused, r.hasFocus)
			if focused == r.hasFocus {
				return
			}
			if focused {
				r.resumeFocus()
			} else {
				layerSetKeyboardNone(&r.win.Window)
				r.hasFocus = false
				r.clearInput()
			}
		})
	}); err != nil {
		log.Println("watchFocus failed to start", err)
	}
}

// this function determines if jumping will result in the sprite immediately being
// inside a platform (due to differing animation frame sizes)
func canJump(sprite *Sprite, platforms []*Platform) bool {
	oldW, oldH := sprite.Size()
	newW, newH := sprite.SizeForState(StateJump)

	x := sprite.X
	if !sprite.facingLeft {
		x += float64(oldW - newW)
	}
	y := sprite.Y + float64(oldH-newH)

	left, right := x, x+float64(newW)
	top, bottom := y, y+float64(newH)

	for _, p := range platforms {
		if right <= p.TopLeft.X || left >= p.BottomRight.X {
			continue
		}
		if bottom <= p.TopLeft.Y || top >= p.BottomRight.Y {
			continue
		}
		return false
	}
	return true
}

// this function is "naive" about sprite height, but that's
// acceptable due to resolveAnims handling it downstream
// in the tick loop
func resolvePlatforms(sprite *Sprite, platforms []*Platform, proposedX, proposedY float64, jumping bool) (float64, float64, bool) {
	sw, sh := sprite.Size()

	moveDirX := sprite.velocityX
	moveDirY := sprite.velocityY

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

		// push out toward whichever edge requires the smaller correction
		leftPen := right - p.TopLeft.X
		rightPen := p.BottomRight.X - left
		switch {
		case leftPen < rightPen:
			resX = p.TopLeft.X - float64(sw)
		case rightPen < leftPen:
			resX = p.BottomRight.X
		case moveDirX >= 0:
			// tie breaker fallback to direction of travel
			resX = p.TopLeft.X - float64(sw)
		default:
			resX = p.BottomRight.X
		}
		sprite.velocityX = 0
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

		topPen := bottom - p.TopLeft.Y
		bottomPen := p.BottomRight.Y - top
		switch {
		case topPen < bottomPen:
			resY = p.TopLeft.Y - float64(sh)
			sprite.grounded = true
		case bottomPen < topPen:
			resY = p.BottomRight.Y
		case moveDirY >= 0:
			resY = p.TopLeft.Y - float64(sh)
			sprite.grounded = true
		default:
			resY = p.BottomRight.Y
		}
		sprite.velocityY = 0
	}

	jumpAllowed := true
	if jumping {
		jumpAllowed = canJump(sprite, platforms)
	}

	return resX, resY, jumpAllowed
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

func (r *Renderer) handleStateEvent(event tcp.Event) {
	if event.WinWidth == 0 || event.WinHeight == 0 || event.ScreenCols == 0 || event.ScreenRows == 0 {
		return
	}

	var cellWidth, cellHeight, letterX, letterY float64

	if r.realCellWidth > 0 && r.realCellHeight > 0 {
		cellWidth, cellHeight = r.realCellWidth, r.realCellHeight
		usedWidth := cellWidth * float64(event.ScreenCols)
		usedHeight := cellHeight * float64(event.ScreenRows)
		letterX = (float64(screenWidth) - usedWidth) / 2
		letterY = (float64(screenHeight) - usedHeight) / 2
	} else {
		cellWidth = float64(screenWidth) / float64(event.ScreenCols)
		cellHeight = float64(screenHeight) / float64(event.ScreenRows)
	}

	fmt.Printf("iceclimber: screenWH=%dx%d cols/rows=%d/%d cellWH=%.2f/%.2f letterXY=%.2f/%.2f\n",
		screenWidth, screenHeight, event.ScreenCols, event.ScreenRows, cellWidth, cellHeight, letterX, letterY) // temp debu

	gutterLeft := event.Config.GutterLeft
	usableCols := event.WinWidth - gutterLeft
	xOffset := float64(gutterLeft)*cellWidth + letterX

	platforms := make([]*Platform, 0, len(event.Lines))
	for i, line := range event.Lines {
		if line.Width == 0 {
			continue
		}

		widthCells := line.Width
		if widthCells > usableCols {
			widthCells = usableCols
		}

		rowTop := float64(i)*cellHeight + letterY
		rowBottom := rowTop + cellHeight

		platforms = append(platforms, NewPlatform(
			Point{X: xOffset, Y: rowTop},
			Point{X: xOffset + float64(widthCells)*cellWidth, Y: rowBottom},
		))
	}

	r.platforms = platforms
	if r.overlayArea != nil {
		r.overlayArea.QueueDraw()
	}

}

func (r *Renderer) handleHelloEvent(event tcp.Event) {
	r.nvimPid = event.Pid
	fmt.Println("iceclimber: hello received, pid =", r.nvimPid)

	cell := GetTerminalCell(r.nvimPid)
	if cell == nil {
		log.Println("iceclimber: terminal did not report pixel geometry; falling back uncorrected math")
		return
	}

	fmt.Println("iceclimber: real cell size from ioctl:", cell.width, cell.height)

	r.realCellWidth = cell.width
	r.realCellHeight = cell.height
}

func (r *Renderer) HandleEvent(event tcp.Event) {
	// get ready for this function to become huge...
	switch event.Type {
	case tcp.EventTypeHello:
		r.handleHelloEvent(event)
	case tcp.EventTypeState:
		r.handleStateEvent(event)
	case "resume_focus":
		r.resumeFocus()
	default:
		fmt.Println("iceclimber: unhandled event type:", event.Type)
	}
}
