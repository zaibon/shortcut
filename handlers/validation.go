package handlers

import "net/url"

// IsValidURL tests if a string is a valid URL
func IsValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
