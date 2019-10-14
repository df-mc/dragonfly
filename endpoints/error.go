package endpoints

const (
	// errPlayerNotFound is returned when a player could not be found by the UUID that was submitted in the
	// form of the request.
	errPlayerNotFound = 0xf0
	// errMalformedUUID is returned when a UUID passed, usually identifying a player, was malformed. The UUID
	// could not be parsed properly.
	errMalformedUUID = 0xf1
)
