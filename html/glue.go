package html

import (
	"maragu.dev/glue/html"

	. "maragu.dev/gomponents"
)

type PageProps = html.PageProps

type PageFunc = html.PageFunc

func ErrorPage() Node {
	return html.ErrorPage(Page)
}

func NotFoundPage() Node {
	return html.NotFoundPage(Page)
}
