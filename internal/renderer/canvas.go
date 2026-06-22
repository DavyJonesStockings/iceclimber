package renderer

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"golang.org/x/sys/unix"
)

const fpsMsInterval = 83

type Canvas struct {
	sprites []*Sprite
	fixed   *gtk.Fixed
} // this is the full screen surface that we draw onto

type Cell struct {
	width  uint16
	height uint16
}

func NewCanvas() *Canvas {
	return &Canvas{fixed: gtk.NewFixed()}
}

func Clamp(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

func (c *Canvas) AddSprite(s *Sprite, x, y int) {
	w, h := s.Size()
	x = Clamp(x, 0, screenWidth-w)
	y = Clamp(y, 0, screenHeight-h)

	c.sprites = append(c.sprites, s)
	s.X, s.Y = x, y
	c.fixed.Put(s.Widget(), float64(x), float64(y))
}

func (c *Canvas) MoveSprite(s *Sprite, x, y int) {
	w, h := s.Size()
	x = Clamp(x, 0, screenWidth-w)
	y = Clamp(y, 0, screenHeight-h)

	s.X, s.Y = x, y
	c.fixed.Move(s.Widget(), float64(x), float64(y))
}

func findTTY(pid int) (string, error) {
	fdDir := "/proc/" + strconv.Itoa(pid) + "/fd"
	fds, err := ioutil.ReadDir(fdDir)
	if err != nil {
		panic(err)
	}

	for _, fd := range fds {
		link := filepath.Join(fdDir, fd.Name())
		target, err := os.Readlink(link)
		if err != nil {
			continue
			fmt.Println("failure in canvas.go findtty")
		}
		if filepath.Dir(target) == "/dev/pts" {
			return target, nil
		}
	}
	return "", fmt.Errorf("no pts device found")
}

func GetTerminalCell(pid int) *Cell {
	tty, err := findTTY(pid)
	if err != nil {
		panic(err)
	}
	fd, err := unix.Open(tty, unix.O_RDONLY|unix.O_NOCTTY, 0)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		panic(err)
	}

	// TODO: fix this. these values are the total number
	// of rows and columns, it should be the actual pixel
	// values for one cell.
	return &Cell{
		width:  ws.Col,
		height: ws.Row,
	}
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
