package handler_test

import (
	"testing"

	"github.com/akfaiz/go-starter-kit/internal/lang"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Handler Suite")
}

var _ = BeforeSuite(func() {
	lang.Init()
})
