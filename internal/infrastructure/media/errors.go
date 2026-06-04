package media

import "errors"

// Sentinel-ошибки media (err113).
var (
	ErrUnsupportedImageModel    = errors.New("unsupported image model")
	ErrImageDimensionsExceeded  = errors.New("image dimensions exceed model limits")
	ErrVideoGenerationNotImpl   = errors.New("video generation not implemented yet")
	ErrVideoStatusNotImpl       = errors.New("video status check not implemented yet")
	ErrUnsupportedMediaType     = errors.New("unsupported media type")
	ErrExpectedThreeVideoVars   = errors.New("expected 3 video variants")
	ErrReplicateTokenNotSet     = errors.New("replicate API token not set")
	ErrReplicateHTTPError       = errors.New("replicate HTTP error")
	ErrReplicatePredictionError = errors.New("replicate prediction error")
	ErrReplicatePollTimeout     = errors.New("replicate poll timeout")
	ErrReplicatePollStatus      = errors.New("replicate poll status error")
)
