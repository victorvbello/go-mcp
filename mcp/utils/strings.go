package utils

import "regexp"

func IsAlphanumeric(str string) bool {
	// Matches strings that contain only alphanumeric characters from start to end.
	// The ^ and $ anchors ensure the entire string must match the pattern.
	var alphanumericRegex = regexp.MustCompile("^[a-zA-Z0-9]*$")
	return alphanumericRegex.MatchString(str)
}
