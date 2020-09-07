package testutils

// SliceContainsString returns true if haystack contains needle, and false otherwise.
func SliceContainsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
