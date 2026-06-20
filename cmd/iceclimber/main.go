package main

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	// "github.com/msgpack-rpc/msgpack-rpc-go/rpc"

	"iceclimber.app/internal/renderer"
	"iceclimber.app/internal/rpc"
)

func main() {
	app := gtk.NewApplication("iceclimber.app", gio.ApplicationFlagsNone)

	var r *renderer.Renderer

	app.ConnectActivate(func() {
		r = renderer.New(app)
		r.Show()
	})

	go rpc.Start(func(event rpc.Event) {
		glib.IdleAdd(func() {
			if r != nil {
				r.HandleEvent(event)
			}
		})
	})

	os.Exit(app.Run(os.Args))
}
