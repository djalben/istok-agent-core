package http

import "errors"

// Sentinel-ошибки HTTP transport (err113).
var (
	ErrRailwayHTTPError    = errors.New("railway HTTP error")
	ErrRailwayGraphQLError = errors.New("railway GraphQL error")
	ErrRailwayEmptyProject = errors.New("railway empty project id")
	ErrInternalPanic       = errors.New("internal panic")
	ErrMarshalRequest      = errors.New("marshal request failed")

	// Preview server-side build.
	ErrPreviewFetch = errors.New("preview dependency fetch failed")
)
