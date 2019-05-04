package commons

import (
	"fmt"
	"strings"
)

// ExtractPathElement extracts the nthElement (zero-based) element of the
// resource path and returns it without surrounding path delimiters ("/").
// An error is returnes if the operation fails.
// For resource "/foo/bar/qux" and nThElement 1, "bar" is returned.
func ExtractPathElement(resource string, nthElement int) (string, error) {
	paths := strings.Split(resource, "/") // ["", "foo", "bar", "qux"]
	paths = paths[1:]                     // cut off starting element (always empty)
	if len(paths) <= nthElement {
		return "", fmt.Errorf("resource path '%s' too short to extract %d elements",
			resource, nthElement)
	}
	return paths[nthElement], nil
}
