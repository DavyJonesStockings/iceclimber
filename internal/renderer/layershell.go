package renderer

// #cgo pkg-config: gtk4-layer-shell-0
// #include <gtk4-layer-shell/gtk4-layer-shell.h>
import "C"
import (
	"unsafe"

	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func gtkWin(win *gtk.Window) *C.GtkWindow {
	return (*C.GtkWindow)(unsafe.Pointer(coreglib.BaseObject(win).Native()))
}

func layerInit(win *gtk.Window) {
	C.gtk_layer_init_for_window(gtkWin(win))
}

func layerSetOverlay(win *gtk.Window) {
	C.gtk_layer_set_layer(gtkWin(win), C.GTK_LAYER_SHELL_LAYER_OVERLAY)
}

func layerSetKeyboardOnDemand(win *gtk.Window) {
	C.gtk_layer_set_keyboard_mode(gtkWin(win), C.GTK_LAYER_SHELL_KEYBOARD_MODE_ON_DEMAND)
}

func layerSetKeyboardExclusive(win *gtk.Window) {
	C.gtk_layer_set_keyboard_mode(gtkWin(win), C.GTK_LAYER_SHELL_KEYBOARD_MODE_EXCLUSIVE)
}

func layerSetExclusiveZoneIgnore(win *gtk.Window) {
	C.gtk_layer_set_exclusive_zone(gtkWin(win), C.int(-1))
}

func layerSetKeyboardNone(win *gtk.Window) {
	C.gtk_layer_set_keyboard_mode(gtkWin(win), C.GTK_LAYER_SHELL_KEYBOARD_MODE_NONE)
}

func layerAnchorLeft(win *gtk.Window, anchor bool) {
	C.gtk_layer_set_anchor(gtkWin(win), C.GTK_LAYER_SHELL_EDGE_LEFT, C.gboolean(boolToInt(anchor)))
}

func layerAnchorRight(win *gtk.Window, anchor bool) {
	C.gtk_layer_set_anchor(gtkWin(win), C.GTK_LAYER_SHELL_EDGE_RIGHT, C.gboolean(boolToInt(anchor)))
}

func layerAnchorTop(win *gtk.Window, anchor bool) {
	C.gtk_layer_set_anchor(gtkWin(win), C.GTK_LAYER_SHELL_EDGE_TOP, C.gboolean(boolToInt(anchor)))
}

func layerAnchorBottom(win *gtk.Window, anchor bool) {
	C.gtk_layer_set_anchor(gtkWin(win), C.GTK_LAYER_SHELL_EDGE_BOTTOM, C.gboolean(boolToInt(anchor)))
}

func layerSetMarginLeft(win *gtk.Window, margin int) {
	C.gtk_layer_set_margin(gtkWin(win), C.GTK_LAYER_SHELL_EDGE_LEFT, C.int(margin))
}

func layerSetMarginTop(win *gtk.Window, margin int) {
	C.gtk_layer_set_margin(gtkWin(win), C.GTK_LAYER_SHELL_EDGE_TOP, C.int(margin))
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
