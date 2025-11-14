package html

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"maragu.dev/glue/html"
	. "maragu.dev/gomponents"
	ds "maragu.dev/gomponents-datastar"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

var hashOnce sync.Once
var appCSSPath, appJSPath string
var htmxJSPath, idiomorphExtJSPath string
var datastarJSPath string

func Page(props PageProps, body ...Node) Node {
	hashOnce.Do(func() {
		appCSSPath = getHashedPath("public/styles/app.css")
		appJSPath = getHashedPath("public/scripts/app.js")

		htmxJSPath = getHashedPath("public/scripts/htmx.js")
		idiomorphExtJSPath = getHashedPath("public/scripts/idiomorph-ext.js")

		datastarJSPath = getHashedPath("public/scripts/datastar.js")
	})

	return HTML5(HTML5Props{
		Title:       props.Title,
		Description: props.Description,
		Language:    "en",
		Head: []Node{
			Link(Rel("stylesheet"), Href(appCSSPath)),
			Script(Src(htmxJSPath), Defer()),
			Script(Src(idiomorphExtJSPath), Defer()),
			Script(Type("module"), Src(datastarJSPath), Defer()),
			Script(Src(appJSPath), Defer()),
			Script(Src("https://cdn.usefathom.com/script.js"), Data("site", "123"), Defer()),
			Meta(Name("htmx-config"), Content(`{"scrollIntoViewOnBoost":false}`)),
			html.FavIcons("app"),
		},
		Body: []Node{Class("bg-primary-600 text-gray-900 dark:text-white"),
			hx.Ext("morph"),
			Div(Class("min-h-dvh flex flex-col justify-between"),
				header(props),
				Div(Class("grow bg-white dark:bg-gray-800 h-auto"),
					container(true,
						Group(body),
					),
				),
				Div(Class("bg-white dark:bg-gray-800"),
					footer(),
				),
			),
		},
	})
}

func header(_ PageProps) Node {
	return Div(
		container(false),
	)
}

func footer() Node {
	return Div(
		container(false,
			Div(Class("flex items-center justify-center space-x-4 sm:space-x-8 py-2"),
				ds.Init("$counter = 0"), ds.OnInterval("$counter++"),
				a(Href("https://www.maragu.dev"), Text("by maragu "), Span(ds.Text("$counter"))),
			),
		),
	)
}

func a(children ...Node) Node {
	return A(Class("text-primary-600 hover:text-primary-400 dark:text-white dark:hover:text-gray-700"), Group(children))
}

func container(padY bool, children ...Node) Node {
	return Div(
		Classes{
			"max-w-7xl mx-auto h-full": true,
			"px-4 sm:px-6 lg:px-8":     true,
			"py-4 md:py-8":             padY,
		},
		Group(children),
	)
}

func getHashedPath(path string) string {
	externalPath := strings.TrimPrefix(path, "public")
	ext := filepath.Ext(path)
	if ext == "" {
		panic("no extension found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Sprintf("%v.x%v", strings.TrimSuffix(externalPath, ext), ext)
	}

	return fmt.Sprintf("%v.%x%v", strings.TrimSuffix(externalPath, ext), sha256.Sum256(data), ext)
}
