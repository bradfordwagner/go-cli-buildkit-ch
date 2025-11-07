package http_handler_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHttpHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HttpHandler Suite")
}

