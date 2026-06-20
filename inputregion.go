package main

// #cgo pkg-config: gtk4
// #include <gtk/gtk.h>
import "C"
import (
	"unsafe"

	coreglib "github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func disableInputRegion(win *gtk.Window) {
	widget := (*C.GtkWidget)(unsafe.Pointer(coreglib.InternObject(win).Native()))
	native := C.gtk_widget_get_native(widget)
	if native == nil {
		return
	}
	surface := C.gtk_native_get_surface(native)
	if surface == nil {
		return
	}
	region := C.cairo_region_create()
	C.gdk_surface_set_input_region(surface, region)
	C.cairo_region_destroy(region)
}
