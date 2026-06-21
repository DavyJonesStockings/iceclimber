package renderer

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type hyprMonitor struct {
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Scale    float64 `json:"scale"`
	Reserved [4]int  `json:"reserved"` // top, bottom, left, right
	Focused  bool    `json:"focused"`
}

func usableScreenSize() (int, int, error) {
	out, err := exec.Command("hyprctl", "monitors", "-j").Output()
	if err != nil {
		return 0, 0, err
	}

	var monitors []hyprMonitor
	if err := json.Unmarshal(out, &monitors); err != nil {
		return 0, 0, err
	}

	for _, m := range monitors {
		if !m.Focused {
			continue
		}
		w := int(float64(m.Width)/m.Scale) - m.Reserved[2] - m.Reserved[3]
		h := int(float64(m.Height)/m.Scale) - m.Reserved[0] - m.Reserved[1]
		return w, h, nil
	}

	return 0, 0, fmt.Errorf("no focused monitor found")
}
