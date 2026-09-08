package handlers

import "strconv"

// parseUint parses a route param into a uint, returning 0 on failure
// (callers rely on the DB foreign-key/lookup to reject a 0 id).
func parseUint(v any) uint {
	s, ok := v.(string)
	if !ok {
		return 0
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return uint(n)
}
