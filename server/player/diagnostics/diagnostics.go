package diagnostics

// Diagnostics represents the latest diagnostics data of the client.
type Diagnostics struct {
	// AverageFramesPerSecond is the average amount of frames per second that the client has been
	// running at.
	AverageFramesPerSecond float64
	// AverageServerSimTickTime is the average time that the server spends simulating a single tick
	// in milliseconds.
	AverageServerSimTickTime float64
	// AverageClientSimTickTime is the average time that the client spends simulating a single tick
	// in milliseconds.
	AverageClientSimTickTime float64
	// AverageBeginFrameTime is the average time that the client spends beginning a frame in
	// milliseconds.
	AverageBeginFrameTime float64
	// AverageInputTime is the average time that the client spends processing input in milliseconds.
	AverageInputTime float64
	// AverageRenderTime is the average time that the client spends rendering in milliseconds.
	AverageRenderTime float64
	// AverageEndFrameTime is the average time that the client spends ending a frame in milliseconds.
	AverageEndFrameTime float64
	// AverageRemainderTimePercent is the average percentage of time that the client spends on
	// tasks that are not accounted for.
	AverageRemainderTimePercent float64
	// AverageUnaccountedTimePercent is the average percentage of time that the client spends on
	// unaccounted tasks.
	AverageUnaccountedTimePercent float64
}
