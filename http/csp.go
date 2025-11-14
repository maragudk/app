package http

import (
	"maragu.dev/httph"
)

func CSP(allowUnsafeInline, allowUnsafeEval bool) func(*httph.ContentSecurityPolicyOptions) {
	return func(opts *httph.ContentSecurityPolicyOptions) {
		opts.ConnectSrc = "'self' https://cdn.usefathom.com"

		opts.ImgSrc = "'self' https://cdn.usefathom.com"

		scriptSrc := "'self' https://cdn.usefathom.com"
		if allowUnsafeInline {
			scriptSrc += " 'unsafe-inline'"
		}
		if allowUnsafeEval {
			scriptSrc += " 'unsafe-eval'"
		}
		opts.ScriptSrc = scriptSrc

		styleSrc := "'self'"
		if allowUnsafeInline {
			styleSrc += " 'unsafe-inline'"
		}
		opts.StyleSrc = styleSrc
	}
}
