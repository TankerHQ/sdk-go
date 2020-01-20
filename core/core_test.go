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
			defer session.Stop() // nolint: errcheck
			Expect(err).ToNot(HaveOccurred())
			Expect(session).ToNot(BeNil())
			Expect(session.GetStatus()).To(Equal(core.StatusReady))
			Expect(session.Stop()).To(Succeed())
			Expect(session.GetStatus()).To(Equal(core.StatusStopped))
		})

		It("Returns a proper error when it fails", func() {
			_, err := core.NewTanker(core.TankerOptions{"", "/tmp", &TestApp.Config.URL})
			Expect(err).To(HaveOccurred())
			terror, ok := (err).(core.Error)
			Expect(ok).To(BeTrue())
			Expect(terror.Code()).To(Equal(core.ErrorInternalError))
		})

		It("Starts and stops a session twice", func() {
			aliceSession, _ := aliceLaptop.Start()
			defer aliceSession.Stop() // nolint: errcheck
			Expect(aliceSession.Stop()).To(Succeed())
			Expect(aliceSession.GetStatus()).To(Equal(core.StatusStopped))

			status, err := aliceSession.Start(alice.Identity)
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(core.StatusReady))
		})

		It("Aborts Registration", func() {
			aliceSession, _ := aliceLaptop.CreateSession()
			status, err := aliceSession.Start(alice.Identity)
			defer aliceSession.Stop() // nolint: errcheck
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(core.StatusIdentityRegistrationNeeded))
			Expect(aliceSession.Stop()).To(Succeed())
			Expect(aliceSession.GetStatus()).To(Equal(core.StatusStopped))
		})

		It("Fails when it open the same device twice", func() {
			session, err := aliceLaptop.Start()
			defer session.Stop() // nolint: errcheck
			Expect(err).ToNot(HaveOccurred())
			Expect(session.GetStatus()).To(Equal(core.StatusReady))
			_, err = aliceLaptop.Start()
			Expect(err).To(HaveOccurred())
		})

		It("Opens a second device with the same user", func() {
			session, err := aliceLaptop.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(session.GetStatus()).To(Equal(core.StatusReady))
			aliceMobile, _ := alice.CreateDevice()
			mobile, err := aliceMobile.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(mobile).ToNot(BeNil())
			Expect(mobile.GetStatus()).To(Equal(core.StatusReady))
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
			defer bobSession.Stop() // nolint: errcheck
			clearData := helpers.RandomBytes(1024 * 1024 * 3)
			encrypted, err := aliceSession.Encrypt(clearData, &core.EncryptOptions{Recipients: []string{bob.PublicIdentity}})
			Expect(err).ToNot(HaveOccurred())
			Expect(encrypted).ToNot(HaveLen(0))
			decrypted, err := bobSession.Decrypt(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(clearData))
		})

		It("Encrypts an empty array", func() {
			encrypted, err := aliceSession.Encrypt([]byte{}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(encrypted).ToNot(HaveLen(0))
		})

		It("Fails to encrypt nil", func() {
			_, err := aliceSession.Encrypt(nil, nil)
			Expect(err).To(HaveOccurred())
		})

		It("Fails to decrypts a too short buffer", func() {
			_, err := aliceSession.Decrypt(nil)
			Expect(err).To(HaveOccurred())
			_, err = aliceSession.Decrypt([]byte{3, 1})
			Expect(err).To(HaveOccurred())
		})

		It("Encrypts then shares with bob", func() {
			bobSession, _ := bobLaptop.Start()
			defer bobSession.Stop() // nolint: errcheck
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
			defer bobSession.Stop() // nolint: errCheck
			attachResult, err := bobSession.AttachProvisionalIdentity(*bobProvisional)
			Expect(err).ToNot(HaveOccurred())
			Expect(attachResult.Status).To(Equal(core.StatusIdentityVerificationNeeded))
			code, err := TestApp.GetVerificationCode(bobEmail)
			Expect(err).ToNot(HaveOccurred())
			Expect(bobSession.VerifyProvisionalIdentity(core.EmailVerification{bobEmail, *code})).To(Succeed())

			attachResult2, err := bobSession.AttachProvisionalIdentity(*bobProvisional)
			Expect(err).ToNot(HaveOccurred())
			Expect(attachResult2.Status).To(Equal(core.StatusReady))
		})

		It("Retrieves a user's device list", func() {
			bobSession, _ := bobLaptop.Start()
			defer bobSession.Stop() // nolint: errCheck
			bobLaptopID, _ := bobSession.GetDeviceID()

			device1, _ := bob.CreateDevice()
			session1, _ := device1.Start()
			defer session1.Stop() // nolint: errCheck
			deviceID1, _ := session1.GetDeviceID()
			Expect(bobSession.RevokeDevice(*deviceID1)).To(Succeed())

			device2, _ := bob.CreateDevice()
			session2, _ := device2.Start()
			defer session2.Stop() // nolint: errCheck
			deviceID2, _ := session2.GetDeviceID()

			devices, err := bobSession.GetDeviceList()
			Expect(err).ToNot(HaveOccurred())
			Expect(devices).To(ConsistOf(
				core.DeviceDescription{*deviceID1, true},
				core.DeviceDescription{*deviceID2, false},
				core.DeviceDescription{*bobLaptopID, false},
			))
		})
	})
})
