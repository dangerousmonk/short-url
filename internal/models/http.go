// Package models describes main entities of the the application
package models

// Request represents model for input data in Shorten HTTP handler
type Request struct {
	URL string `json:"url"`
}

// Response represents model for response data in Shorten HTTP handler
type Response struct {
	Result string `json:"result"`
}
