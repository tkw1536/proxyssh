package testutils

// SliceContainsString checks if the slice slice contains the string str
func SliceContainsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}