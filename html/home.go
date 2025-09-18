package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type HomePageProps struct {
	PageProps
}

func HomePage(props HomePageProps) Node {
	return Page(props.PageProps,
		H1(Class("text-6xl font-bold"), Text("App 😎")),
	)
}
