package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a title into a URL-friendly slug, e.g. "Go Conf 2024!" -> "go-conf-2024".
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugNonAlnum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// UniqueSlugSuffix appends a short numeric suffix, used when a base slug collides.
func UniqueSlugSuffix(base string, n int) string {
	return fmt.Sprintf("%s-%d", base, n)
}
