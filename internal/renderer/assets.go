package renderer

import (
	"embed"

	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
)

//go:embed popo/*.png
var popoFrameFS embed.FS

func loadPixbufFromEmbed(path string) (*gdkpixbuf.Pixbuf, error) {
	data, err := popoFrameFS.ReadFile(path)
	if err != nil {
		return nil, err
	}

	loader := gdkpixbuf.NewPixbufLoader()

	if err := loader.Write(data); err != nil {
		return nil, err
	}
	if err := loader.Close(); err != nil {
		return nil, err
	}

	return loader.Pixbuf(), nil
}

func popoFramePaths() map[AnimationState][]string {
	return map[AnimationState][]string{
		StateWalk: {
			"popo/popo_walk0000.png",
			"popo/popo_walk0001.png",
			"popo/popo_walk0002.png",
			"popo/popo_walk0003.png",
		},
		StateIdle: {
			"popo/popo_idle0000.png",
		},
		StateJump: {
			"popo/popo_jump.png",
		},
		StateFall: {
			"popo/popo_fall.png",
		},
	}
}
