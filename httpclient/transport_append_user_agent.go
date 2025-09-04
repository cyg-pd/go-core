package httpclient

import "net/http"

// appendUserAgentTransport appends a custom string to the User-Agent header.
type appendUserAgentTransport struct {
	Transport http.RoundTripper
	AppendUA  string
}

// RoundTrip implements the http.RoundTripper interface.
func (t *appendUserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Get the existing User-Agent header.
	existingUA := req.Header.Get("User-Agent")

	// If no User-Agent is set, the default "Go-http-client/1.1" will be added by the http package later.
	// So, we prepend our custom string.
	if existingUA == "" {
		req.Header.Set("User-Agent", t.AppendUA)
	} else {
		// If a User-Agent is already set, append to it.
		req.Header.Set("User-Agent", existingUA+" "+t.AppendUA)
	}

	// Use the underlying transport to perform the request.
	return t.Transport.RoundTrip(req)
}
