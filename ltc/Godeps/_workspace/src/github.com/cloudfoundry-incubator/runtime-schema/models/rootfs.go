package models

import "net/url"

const PreloadedRootFSScheme = "preloaded"

func PreloadedRootFS(stack string) string {
	return (&url.URL{
		Scheme: PreloadedRootFSScheme,
		Opaque: stack,
	}).String()
}
