package renderer

import "os"

func popoFramePaths() map[AnimationState][]string {
	home, _ := os.UserHomeDir()
	return map[AnimationState][]string{
		StateWalk: {
			home + "/pictures/krita/popo_walk0000.png",
			home + "/pictures/krita/popo_walk0001.png",
			home + "/pictures/krita/popo_walk0002.png",
			home + "/pictures/krita/popo_walk0003.png",
		},
		StateIdle: {
			home + "/pictures/krita/popo_idle0000.png",
		},
		StateJump: {
			home + "/pictures/krita/popo_idle0000.png",
		},
		StateFall: {
			home + "/pictures/krita/popo_idle0000.png",
		},
	}
}
