package item

// Camera represents an item that behaves as a camera, allowing the user to take pictures.
type Camera interface {
	// CameraInfo returns the camera information of the item.
	CameraInfo() CameraInfo
}

// CameraInfo is a struct returned by items that implement Camera. It contains the information required for
// the client to handle the item as a camera.
type CameraInfo struct {
	// BlackBarsDuration is the duration in seconds the black bars are shown when taking a picture.
	BlackBarsDuration float64
	// BlackBarsScreenRatio is the ratio of the screen covered by the black bars.
	BlackBarsScreenRatio float64
	// PictureDuration is the duration in seconds the picture is shown.
	PictureDuration float64
	// ShutterDuration is the duration in seconds the shutter is shown.
	ShutterDuration float64
	// ShutterScreenRatio is the ratio of the screen covered by the shutter.
	ShutterScreenRatio float64
	// SlideAwayDuration is the duration in seconds the picture slides away.
	SlideAwayDuration float64
	// UseDuration is the duration in ticks the item takes to be fully used.
	UseDuration int
}
