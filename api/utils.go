package api

import "fmt"

// ConvertDuration take a string "hh:mm:ss" and convert it in seconds
func ConvertDuration(t string) (int, error) {
	var h, m, s int

	if _, err := fmt.Sscanf(t, "%d:%d:%d", &h, &m, &s); err != nil {
		return 0, err
	}

	return (h * 3600) + (m * 60) + s, nil
}
