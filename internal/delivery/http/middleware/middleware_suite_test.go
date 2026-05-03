package middleware_test

import (
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/lang"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Middleware Suite")
}

var _ = BeforeSuite(func() {
	lang.Init()
})
