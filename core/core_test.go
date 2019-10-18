package core_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
	"github.com/TankerHQ/sdk-go/v2/helpers"
)

var _ = Describe("functional", func() {

	var (
		alice       helpers.User
		aliceLaptop *helpers.Device
	)

	BeforeEach(func() {
		alice = TestApp.CreateUser()
		aliceLaptop, _ = alice.CreateDevice()
	})

	Context("Bascis", func() {

		It("Starts and stops a session", func() {
			aliceLaptop, err := alice.CreateDevice()
			Expect(err).ToNot(HaveOccurred())
			session, err := aliceLaptop.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(session).ToNot(BeNil())
			Expect(session.GetStatus()).To(Equal(core.TankerStatusReady))
			Expect(session.Stop()).To(Succeed())
			Expect(session.GetStatus()).To(Equal(core.TankerStatusStopped))
		})

		It("Returns a proper error when it fails", func() {
			_, err := core.CreateTanker("", TestApp.Config.URL, "/tmp")
			Expect(err).To(HaveOccurred())
			terror, ok := (err).(core.Error)
			Expect(ok).To(BeTrue())
			Expect(terror.Code()).To(Equal(core.ErrorInternalError))
		})

		It("Starts and stops a session twice", func() {
			aliceSession, err := aliceLaptop.Start()
			Expect(aliceSession.Stop()).To(Succeed())
			Expect(aliceSession.GetStatus()).To(Equal(core.TankerStatusStopped))

			status, err := aliceSession.Start(alice.Identity)
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(core.TankerStatusReady))
		})

		It("Aborts Registration", func() {
			aliceSession, err := aliceLaptop.CreateSession()
			status, err := aliceSession.Start(alice.Identity)
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(core.TankerStatusIdentityRegistrationNeeded))
			Expect(aliceSession.Stop()).To(Succeed())
			Expect(aliceSession.GetStatus()).To(Equal(core.TankerStatusStopped))
		})

		// Note: crashes on Windows
		XIt("Fails when it open the same device twice", func() {
			session, err := aliceLaptop.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(session.GetStatus()).To(Equal(core.TankerStatusReady))
			_, err = aliceLaptop.Start()
			Expect(err).To(HaveOccurred())
		})

		It("Opens a second device with the same user", func() {
			session, err := aliceLaptop.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(session.GetStatus()).To(Equal(core.TankerStatusReady))
			aliceMobile, _ := alice.CreateDevice()
			mobile, err := aliceMobile.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(mobile).ToNot(BeNil())
			Expect(mobile.GetStatus()).To(Equal(core.TankerStatusReady))
			Expect(mobile.Stop()).To(Succeed())
			Expect(session.Stop()).To(Succeed())
		})

	})

	Context("Alice has a session", func() {

		var (
			bob          helpers.User
			bobLaptop    *helpers.Device
			aliceSession *core.Tanker
		)

		BeforeEach(func() {
			bob = TestApp.CreateUser()
			bobLaptop, _ = bob.CreateDevice()

			aliceSession, _ = aliceLaptop.Start()
		})

		AfterEach(func() {
			Expect(aliceSession.Stop()).To(Succeed())
		})

		It("Gets the deviceID", func() {
			ID, err := aliceSession.GetDeviceID()
			Expect(err).ToNot(HaveOccurred())
			Expect(*ID).ToNot(BeEmpty())
		})

		It("Encrypts and Decrypts", func() {
			clearData := helpers.RandomBytes(1024 * 1024 * 3)
			encrypted, err := aliceSession.Encrypt(clearData, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(encrypted).ToNot(HaveLen(0))
			decrypted, err := aliceSession.Decrypt(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(clearData))
		})

		It("Encrypts and shares with bob", func() {
			bobSession, _ := bobLaptop.Start()
			clearData := helpers.RandomBytes(1024 * 1024 * 3)
			encrypted, err := aliceSession.Encrypt(clearData, &core.EncryptOptions{Recipients: []string{bob.PublicIdentity}})
			Expect(err).ToNot(HaveOccurred())
			Expect(encrypted).ToNot(HaveLen(0))
			decrypted, err := bobSession.Decrypt(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(clearData))
		})

		It("Encrypts then shares with bob", func() {
			bobSession, _ := bobLaptop.Start()
			clearData := helpers.RandomBytes(1024 * 1024 * 3)
			encrypted, err := aliceSession.Encrypt(clearData, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(encrypted).ToNot(HaveLen(0))
			resourceId, err := aliceSession.GetResourceId(encrypted)
			Expect(err).ToNot(HaveOccurred())
			err = aliceSession.Share([]string{*resourceId}, []string{bob.PublicIdentity}, nil)
			Expect(err).ToNot(HaveOccurred())
			decrypted, err := bobSession.Decrypt(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(clearData))
		})

		It("Claims the same provisional Identity twice", func() {
			bobEmail := "bob@gmail.com"
			bobProvisional, err := identity.CreateProvisional(TestApp.IdConfig, bobEmail)
			Expect(err).ToNot(HaveOccurred())
			clearData := helpers.RandomBytes(12)
			bobPublicProvisional, err := identity.GetPublicIdentity(*bobProvisional)
			Expect(err).ToNot(HaveOccurred())
			// Trigger the creation of the provisional Identity on the Tanker Server
			_, err = aliceSession.Encrypt(clearData, &core.EncryptOptions{Recipients: []string{*bobPublicProvisional}})
			Expect(err).ToNot(HaveOccurred())
			bobSession, _ := bobLaptop.Start()
			attachResult, err := bobSession.AttachProvisionalIdentity(*bobProvisional)
			Expect(err).ToNot(HaveOccurred())
			Expect(attachResult.Status).To(Equal(core.TankerStatusIdentityVerificationNeeded))
			code, err := TestApp.GetVerificationCode(bobEmail)
			Expect(err).ToNot(HaveOccurred())
			Expect(bobSession.VerifyProvisionalIdentity(core.EmailVerification{bobEmail, *code})).To(Succeed())

			attachResult2, err := bobSession.AttachProvisionalIdentity(*bobProvisional)
			Expect(err).ToNot(HaveOccurred())
			Expect(attachResult2.Status).To(Equal(core.TankerStatusReady))
		})
	})

})
