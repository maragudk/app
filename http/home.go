package http

import (
	"log/slog"

	"app/html"

	. "maragu.dev/gomponents"
)

func Home(r *Router, log *slog.Logger) {
	r.Get("/", func(props html.PageProps) (Node, error) {
		return html.HomePage(html.HomePageProps{PageProps: props}), nil
	})
}
