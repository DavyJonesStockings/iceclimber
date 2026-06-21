package renderer

import (
	"github.com/diamondburned/gotk4/pkg/cairo"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const scale = 4

type AnimationState string

const (
	StateIdle AnimationState = "idle"
	StateWalk AnimationState = "walk"
)

type Sprite struct {
	animations   map[AnimationState][]*gdkpixbuf.Pixbuf
	state        AnimationState
	currentFrame int
	facingLeft   bool
	X, Y         int
	area         *gtk.DrawingArea
}

func NewSprite(animations map[AnimationState][]string) (*Sprite, error) {
	loaded := make(map[AnimationState][]*gdkpixbuf.Pixbuf)
	for state, paths := range animations {
		frames := make([]*gdkpixbuf.Pixbuf, len(paths))
		for i, path := range paths {
			pb, err := gdkpixbuf.NewPixbufFromFile(path)
			if err != nil {
				return nil, err
			}
			frames[i] = pb.ScaleSimple(pb.Width()*scale, pb.Height()*scale, gdkpixbuf.InterpNearest)
		}
		loaded[state] = frames
	}
	return &Sprite{animations: loaded, state: StateIdle}, nil
}

func (s *Sprite) frames() []*gdkpixbuf.Pixbuf {
	return s.animations[s.state]
}

func (s *Sprite) SetState(state AnimationState) {
	if state == s.state {
		return
	}
	s.state = state
	s.currentFrame = 0
}

func (s *Sprite) SetFacing(left bool) { s.facingLeft = left }

func (s *Sprite) Tick() {
	frames := s.frames()
	if len(frames) > 0 {
		s.currentFrame = (s.currentFrame + 1) % len(frames)
	}
	if s.area != nil {
		s.area.QueueDraw()
	}
}

func (s *Sprite) Size() (int, int) {
	first := s.animations[StateIdle][0]
	return first.Width(), first.Height()
}

func (s *Sprite) draw(a *gtk.DrawingArea, cr *cairo.Context, width, height int) {
	cr.SetOperator(cairo.OperatorSource)
	cr.SetSourceRGBA(0, 0, 0, 0)
	cr.Paint()
	cr.SetOperator(cairo.OperatorOver)

	frames := s.frames()
	if len(frames) == 0 {
		return
	}
	frame := frames[s.currentFrame]

	if s.facingLeft {
		gdk.CairoSetSourcePixbuf(cr, frame, 0, 0)
		cr.Paint()
	} else {
		cr.Save()
		cr.Translate(float64(frame.Width()), 0)
		cr.Scale(-1, 1)
		gdk.CairoSetSourcePixbuf(cr, frame, 0, 0)
		cr.Paint()
		cr.Restore()
	}
}

func (s *Sprite) Widget() *gtk.DrawingArea {
	if s.area == nil {
		first := s.animations[StateIdle][0]
		s.area = gtk.NewDrawingArea()
		s.area.SetSizeRequest(first.Width(), first.Height())
		s.area.SetDrawFunc(s.draw)
	}
	return s.area
}
