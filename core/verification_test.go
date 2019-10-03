package core_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/TankerHQ/sdk-go/v2/core"
)

var _ = Describe("functional", func() {
	It("Generates a verification key", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, err := aliceLaptop.CreateSession()
		Expect(err).ToNot(HaveOccurred())
		Expect(session.Start(alice.Identity)).To(Equal(core.TankerStatusIdentityRegistrationNeeded))
		Expect(session.GenerateVerificationKey()).ToNot(BeNil())
	})

	It("Gets verification methods", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, _ := aliceLaptop.Start()
		methods, err := session.GetVerificationMethods()
		Expect(err).ToNot(HaveOccurred())
		Expect(methods).To(HaveLen(1))
		Expect(methods).To(ConsistOf(core.VerificationMethod{Type: core.VerificationMethodTypePassphrase}))
	})

	It("Set verification method", func() {
		alice := TestApp.CreateUser()
		aliceEmail := "alice@domain.io"
		code, err := TestApp.GetVerificationCode(aliceEmail)
		Expect(err).ToNot(HaveOccurred())

		aliceLaptop, _ := alice.CreateDevice()
		session, _ := aliceLaptop.Start()
		Expect(session.SetVerificationMethod(
			core.EmailVerification{Email: aliceEmail, VerificationCode: *code},
		)).To(Succeed())
		methods, err := session.GetVerificationMethods()
		Expect(err).ToNot(HaveOccurred())
		Expect(methods).To(HaveLen(2))
		Expect(methods).To(ConsistOf(
			core.VerificationMethod{Type: core.VerificationMethodTypePassphrase},
			core.VerificationMethod{Type: core.VerificationMethodTypeEmail, Email: &aliceEmail},
		))
	})

})
