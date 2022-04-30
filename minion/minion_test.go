package minion_test

import (
	"github.com/mgjules/minion/minion"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Minion", func() {
	Describe("Creating a new Minion", func() {
		Context("with a name and key", func() {
			It("should print a minion's introduction with the specified name and key", func() {
				m, err := minion.New("minion", "key")
				Expect(err).ToNot(HaveOccurred())
				Expect(m.String()).To(Equal("My name is 'minion' and I have a secret key 'key'."))
			})
		})

		Context("with a name but no key", func() {
			It("should print a minion's introduction with the specified name and a random key", func() {
				m, err := minion.New("minion", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(m.String()).To(
					MatchRegexp("My name is 'minion' and I have a secret key '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'."), //nolint:revive
				)
			})
		})

		Context("with no name but a key", func() {
			It("should print a minion's introduction with no name and specified key", func() {
				m, err := minion.New("", "key")
				Expect(err).ToNot(HaveOccurred())
				Expect(m.String()).To(Equal("My name is '' and I have a secret key 'key'."))
			})
		})

		Context("with no name and no key", func() {
			It("should print a minion's introduction with no name but a random key", func() {
				m, err := minion.New("", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(m.String()).To(
					MatchRegexp("My name is '' and I have a secret key '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'."), //nolint:revive
				)
			})
		})
	})
})
