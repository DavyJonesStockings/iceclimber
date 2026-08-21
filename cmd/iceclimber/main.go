package main

import (
	"flag"
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	// "github.com/msgpack-rpc/msgpack-rpc-go/rpc"

	"iceclimber.app/internal/renderer"
	"iceclimber.app/internal/tcp"
)

func main() {
	standalone := flag.Bool("standalone", false, "run without waiting for nvim tcp connect (for debugging)")
	flag.Parse()
	app := gtk.NewApplication("iceclimber.app", gio.ApplicationFlagsNone)

	var r *renderer.Renderer

	app.ConnectActivate(func() {
		r = renderer.New(app)
		r.Show()
	})

	if !*standalone {
		go func() {
			s := tcp.Start(func(event tcp.Event) {
				glib.IdleAdd(func() {
					if r != nil {
						r.HandleEvent(event)
					}
				})
			})
			glib.IdleAdd(func() {
				if r != nil {
					r.SetServer(s)
				}
			})
		}()
	}

	os.Exit(app.Run(os.Args[:1]))
}
