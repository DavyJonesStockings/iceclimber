package renderer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

func hyprSocketPath(socketName string) (string, error) {
	sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if sig == "" {
		return "", fmt.Errorf("HYPRLAND_INSTANCE_SIGNATURE not set — not running under Hyprland")
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		return "", fmt.Errorf("XDG_RUNTIME_DIR not set")
	}
	return fmt.Sprintf("%s/hypr/%s/%s", runtimeDir, sig, socketName), nil
}

type hyprActiveWindow struct {
	Address string `json:"address"`
	At      [2]int `json:"at"`
	Size    [2]int `json:"size"`
}

// currentFocusedWindowAddress queries whatever window currently has focus,
// meant to be called once at game-start to capture "this is the target."
func currentFocusedWindowInfo() (hyprActiveWindow, error) {
	out, err := exec.Command("hyprctl", "activewindow", "-j").Output()
	if err != nil {
		return hyprActiveWindow{}, err
	}
	var w hyprActiveWindow
	if err := json.Unmarshal(out, &w); err != nil {
		return hyprActiveWindow{}, err
	}
	return w, nil
}

// watchFocus connects to Hyprland's event socket and invokes cb with the
// newly-focused window's address every time focus changes (empty string if
// focus goes to nothing, e.g. the desktop). Runs on its own goroutine;
// cb is NOT called on the GTK main thread — wrap GTK calls in glib.IdleAdd.
func watchFocus(cb func(windowAddress string)) error {
	path, err := hyprSocketPath(".socket2.sock")
	if err != nil {
		return err
	}

	conn, err := net.Dial("unix", path)
	if err != nil {
		return fmt.Errorf("connect to hyprland event socket: %w", err)
	}

	go func() {
		defer conn.Close()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			name, data, found := strings.Cut(scanner.Text(), ">>")
			if !found {
				continue
			}
			if name == "activewindowv2" {
				cb(data)
			}
		}
	}()

	return nil
}
