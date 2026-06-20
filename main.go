package main

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	// "github.com/msgpack-rpc/msgpack-rpc-go/rpc"
)

func main() {
	app := gtk.NewApplication("iceclimber.app", gio.ApplicationFlagsNone)

	var renderer *Renderer

	app.ConnectActivate(func() {
		renderer = NewRenderer(app)
		renderer.Show()
	})

	go listenIPC(func(event Event) {
		// Marshal back onto the GTK thread — never call renderer methods directly from here.
		glib.IdleAdd(func() {
			renderer.HandleEvent(event)
		})
	})

	os.Exit(app.Run(os.Args))
}
