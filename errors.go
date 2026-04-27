package nendesktop

import "fmt"

// APIError is returned for non-2xx responses from the Nen Desktop API.
// Callers can inspect StatusCode and Body for details.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Body)
}
