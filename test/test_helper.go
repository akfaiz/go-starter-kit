package test

import "github.com/gavv/httpexpect/v2"

// ProblemJSON is a global constant for Content-Type application/problem+json.
// It is used to verify that error responses follow RFC 7807.
var ProblemJSON = httpexpect.ContentOpts{
	MediaType: "application/problem+json",
}
