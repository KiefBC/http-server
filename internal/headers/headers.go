package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

// NewHeaders creates and returns a new Headers map.
func NewHeaders() Headers {
	return make(Headers)
}

// Parse processes HTTP header data from the given bytes.
// Returns the number of bytes consumed, whether parsing is done, and any error encountered.
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Look for CRLF line terminator (RFC 9112 Section 2.2)
	crlfIndex := strings.Index(string(data), "\r\n")
	if crlfIndex == -1 {
		// No CRLF found, need more data
		return 0, false, nil
	}

	// Empty line indicates end of header section (RFC 9112 Section 2.2)
	if crlfIndex == 0 {
		return 2, true, nil
	}

	line := string(data[:crlfIndex])

	// Parse header field: field-name ":" field-value (RFC 9112 Section 5)
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 {
		return 0, false, fmt.Errorf("invalid header format: no colon found")
	}

	key := line[:colonIndex]
	value := line[colonIndex+1:]

	// Field names must be valid tokens - no whitespace allowed (RFC 9112 Section 5.1)
	if strings.HasSuffix(key, " ") {
		return 0, false, fmt.Errorf("invalid header format: space before colon")
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if strings.Contains(key, " ") {
		return 0, false, fmt.Errorf("invalid header format: field name contains spaces")
	}

	if !validationToken(key) {
		return 0, false, fmt.Errorf("invalid header key: contains invalid characters")
	}
	h.Set(key, value)

	return crlfIndex + 2, false, nil
}

// Set adds or updates a header field with the given key and value.
// It handles case-insensitivity of header field names and combines multiple values with commas.
func (h Headers) Set(key, value string) {
	// Field names are case-insensitive (RFC 9110 Section 5.1)
	lowerKey := strings.ToLower(key)
	// Combine multiple field lines with same name using comma separation (RFC 9110 Section 5.2)
	if h[lowerKey] != "" {
		h[lowerKey] = h[lowerKey] + ", " + value
	} else {
		h[lowerKey] = value
	}
}

// Replace sets a header field with the given key and value, replacing any existing value.
// Unlike Set, this does not append to existing values but completely replaces them.
func (h Headers) Replace(key, value string) {
	// Field names are case-insensitive (RFC 9110 Section 5.1)
	lowerKey := strings.ToLower(key)
	h[lowerKey] = value
}

// Get retrieves the value of a header field by key.
// Header field names are case-insensitive per RFC 9110 Section 5.1.
func (h Headers) Get(key string) (value string, ok bool) {
	loweredKey := strings.ToLower(key)
	if val, exists := h[loweredKey]; exists {
		return val, true
	}

	return "", false
}

// validationToken checks if the header key contains only valid ASCII characters.
func validationToken(key string) bool {
	// Validate header key contains only ASCII characters (RFC 9110 Section 5.1)
	for _, char := range key {
		if char > 127 {
			return false
		}
	}

	return true
}
