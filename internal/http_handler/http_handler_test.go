package http_handler_test

import (
	"bkch/internal/args"
	"bkch/internal/cache"
	"bkch/internal/http_handler"
	"context"
	"fmt"
	"net/http/httptest"

	"github.com/bradfordwagner/go-util/bwutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HttpHandler", func() {
	var (
		ctx        context.Context
		cancel     context.CancelFunc
		serverArgs args.ServerArgs
		testCache  *bwutil.Lockable[*cache.Cache]
		handler    *http_handler.Handler
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		serverArgs = args.ServerArgs{
			Port:                8080,
			Kubeconfig:          "/path/to/kubeconfig",
			KubernetesNamespace: "test-namespace",
			SelectorLabel:       "app=test",
		}

		// Create a test cache with some pods
		c := cache.NewCache()
		c.DnsFormatApiGateway = "api-gateway-pod-%d.example.com"
		c.DnsFormatInCluster = "pod-%d.test-namespace.svc.cluster.local"
		c.Replicas = 3

		// Add some available pods
		c.Pods["pod-0"] = &cache.Pod{IsAvailable: true, Index: 0}
		c.Pods["pod-1"] = &cache.Pod{IsAvailable: true, Index: 1}
		c.Pods["pod-2"] = &cache.Pod{IsAvailable: true, Index: 2}

		testCache = bwutil.NewLockableWithValue(c)
		handler = http_handler.NewHandler(ctx, cancel, serverArgs, testCache)
	})

	AfterEach(func() {
		cancel()
	})

	Describe("NewHandler", func() {
		It("creates a new handler with valid arguments", func() {
			Expect(handler).NotTo(BeNil())
		})
	})

	Describe("handler with API Gateway mode", func() {
		It("returns a consistent hash result for a given hash", func() {
			w := httptest.NewRecorder()

			// We need to manually call the handler method through reflection or
			// by using the actual HTTP routing. For simplicity, we'll test via
			// the HTTP server routes which we'll test separately.
			// For now, let's test that we can get consistent hash results directly
			c := testCache.Get()
			result, err := c.ConsistentHash(cache.HashModeAPIGateway, "test-hash", w)
			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring("api-gateway-pod-"))
			Expect(result).To(ContainSubstring(".example.com"))
		})

		It("returns the same result for the same hash", func() {
			w1 := httptest.NewRecorder()
			w2 := httptest.NewRecorder()
			c := testCache.Get()

			result1, err1 := c.ConsistentHash(cache.HashModeAPIGateway, "consistent-hash", w1)
			result2, err2 := c.ConsistentHash(cache.HashModeAPIGateway, "consistent-hash", w2)

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(result1).To(Equal(result2))
		})

		It("handles error when no pods are available", func() {
			// Create cache with no available pods
			emptyCache := cache.NewCache()
			emptyCache.DnsFormatApiGateway = "api-gateway-pod-%d.example.com"
			emptyCache.Pods["pod-0"] = &cache.Pod{IsAvailable: false, Index: 0}

			emptyCacheLockable := bwutil.NewLockableWithValue(emptyCache)
			emptyHandler := http_handler.NewHandler(ctx, cancel, serverArgs, emptyCacheLockable)

			Expect(emptyHandler).NotTo(BeNil())

			w := httptest.NewRecorder()
			c := emptyCacheLockable.Get()
			result, err := c.ConsistentHash(cache.HashModeAPIGateway, "test-hash", w)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("no available pods"))
			Expect(result).To(BeEmpty())
		})
	})

	Describe("handler with In-Cluster mode", func() {
		It("returns a consistent hash result for a given hash", func() {
			w := httptest.NewRecorder()
			c := testCache.Get()

			result, err := c.ConsistentHash(cache.HashModeInCluster, "test-hash", w)

			Expect(err).To(BeNil())
			Expect(result).To(ContainSubstring("pod-"))
			Expect(result).To(ContainSubstring(".test-namespace.svc.cluster.local"))
		})

		It("returns the same result for the same hash", func() {
			w1 := httptest.NewRecorder()
			w2 := httptest.NewRecorder()
			c := testCache.Get()

			result1, err1 := c.ConsistentHash(cache.HashModeInCluster, "consistent-hash", w1)
			result2, err2 := c.ConsistentHash(cache.HashModeInCluster, "consistent-hash", w2)

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			Expect(result1).To(Equal(result2))
		})
	})

	Describe("hash distribution", func() {
		It("distributes different hashes across multiple pods", func() {
			w := httptest.NewRecorder()
			c := testCache.Get()
			results := make(map[string]bool)

			// Try multiple hashes to see if we get different pod assignments
			for i := 0; i < 100; i++ {
				hashInput := fmt.Sprintf("test-hash-%d", i)
				result, err := c.ConsistentHash(cache.HashModeAPIGateway, hashInput, w)
				Expect(err).To(BeNil())
				results[result] = true
			}

			// We should get at least 2 different pod assignments
			// (statistically very likely with 100 different hashes and 3 pods)
			Expect(len(results)).To(BeNumerically(">=", 2))
		})
	})

	Describe("context cancellation", func() {
		It("allows cancellation through the cancel function", func() {
			Expect(ctx.Err()).To(BeNil())
			cancel()
			Eventually(ctx.Done()).Should(BeClosed())
		})
	})

	Describe("cache updates", func() {
		It("reflects changes when pods are added", func() {
			c := testCache.Get()
			initialPodCount := len(c.Pods)

			// Add a new pod
			c.Pods["pod-3"] = &cache.Pod{IsAvailable: true, Index: 3}

			Expect(len(c.Pods)).To(Equal(initialPodCount + 1))
		})

		It("reflects changes when pods are marked unavailable", func() {
			c := testCache.Get()

			// Mark all pods as unavailable
			for _, pod := range c.Pods {
				pod.IsAvailable = false
			}

			w := httptest.NewRecorder()
			result, err := c.ConsistentHash(cache.HashModeAPIGateway, "test-hash", w)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("no available pods"))
			Expect(result).To(BeEmpty())
		})

		It("handles cache with single available pod", func() {
			// Create cache with only one available pod
			singleCache := cache.NewCache()
			singleCache.DnsFormatApiGateway = "api-gateway-pod-%d.example.com"
			singleCache.Pods["pod-0"] = &cache.Pod{IsAvailable: true, Index: 0}

			singleCacheLockable := bwutil.NewLockableWithValue(singleCache)
			w1 := httptest.NewRecorder()
			w2 := httptest.NewRecorder()

			c := singleCacheLockable.Get()
			result1, err1 := c.ConsistentHash(cache.HashModeAPIGateway, "hash1", w1)
			result2, err2 := c.ConsistentHash(cache.HashModeAPIGateway, "hash2", w2)

			Expect(err1).To(BeNil())
			Expect(err2).To(BeNil())
			// With only one pod, all hashes should go to the same pod
			Expect(result1).To(Equal(result2))
			Expect(result1).To(Equal("api-gateway-pod-0.example.com"))
		})
	})

	Describe("DNS format strings", func() {
		It("uses correct format for API Gateway mode", func() {
			c := testCache.Get()
			w := httptest.NewRecorder()

			result, err := c.ConsistentHash(cache.HashModeAPIGateway, "test", w)

			Expect(err).To(BeNil())
			Expect(result).To(MatchRegexp(`api-gateway-pod-\d+\.example\.com`))
		})

		It("uses correct format for In-Cluster mode", func() {
			c := testCache.Get()
			w := httptest.NewRecorder()

			result, err := c.ConsistentHash(cache.HashModeInCluster, "test", w)

			Expect(err).To(BeNil())
			Expect(result).To(MatchRegexp(`pod-\d+\.test-namespace\.svc\.cluster\.local`))
		})
	})
})

