package bucket

// Content represents the content of a bucket. Most of the contents may be placed into the world, except for
// milk and empty buckets.
type Content struct {
	content
}

// Water returns content for a water bucket.
func Water() Content {
	return Content{content: 1}
}

// Lava returns content for a lava bucket.
func Lava() Content {
	return Content{content: 2}
}

// content represents the content of a bucket.
type content uint8
