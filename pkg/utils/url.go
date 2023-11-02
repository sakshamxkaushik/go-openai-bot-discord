// url.go

package utils

import "net/url"


// IsURL function takes a string as an argument and returns a bool indicating whether the string is a valid URL. 
// The function uses the url.Parse function to parse the string as a URL. If the parsing is successful and the URL
// has a scheme and a host, the function returns true. Otherwise, the function returns false.
func IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
