package core_test

import (
	"crypto/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
)

func randomBytes(size int) []byte {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil
	}
	return bytes
}

var _ = Describe("functional", func() {

	It("Starts and stops a session", func() {
		alice := TestApp.CreateUser()
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
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, _ := aliceLaptop.Start()
		Expect(session.Stop()).To(Succeed())
		Expect(session.GetStatus()).To(Equal(core.TankerStatusStopped))

		status, err := session.Start(alice.Identity)
		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(core.TankerStatusReady))
	})

	It("Aborts Registration", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, err := aliceLaptop.CreateSession()
		Expect(err).ToNot(HaveOccurred())
		status, err := session.Start(alice.Identity)
		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(core.TankerStatusIdentityRegistrationNeeded))
		Expect(session.Stop()).To(Succeed())
		Expect(session.GetStatus()).To(Equal(core.TankerStatusStopped))
	})

	// Note: crashes on Windows
	XIt("Fails when it open the same device twice", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, err := aliceLaptop.Start()
		Expect(err).ToNot(HaveOccurred())
		Expect(session.GetStatus()).To(Equal(core.TankerStatusReady))
		_, err = aliceLaptop.Start()
		Expect(err).To(HaveOccurred())
	})

	It("Opens a second device with the same user", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
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

	It("Gets the deviceID", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, _ := aliceLaptop.Start()
		ID, err := session.GetDeviceID()
		Expect(err).ToNot(HaveOccurred())
		Expect(*ID).ToNot(BeEmpty())
	})

	It("Encrypts and Decrypts", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, _ := aliceLaptop.Start()
		clearData := randomBytes(1024 * 1024 * 3)
		encrypted, err := session.Encrypt(clearData, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(encrypted).ToNot(HaveLen(0))
		decrypted, err := session.Decrypt(encrypted)
		Expect(err).ToNot(HaveOccurred())
		Expect(decrypted).To(Equal(clearData))
	})

	It("Encrypts and shares with bob", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		aliceSession, _ := aliceLaptop.Start()
		bob := TestApp.CreateUser()
		bobLaptop, _ := bob.CreateDevice()
		bobSession, _ := bobLaptop.Start()
		clearData := randomBytes(1024 * 1024 * 3)
		encrypted, err := aliceSession.Encrypt(clearData, &core.EncryptOptions{Recipients: []string{bob.PublicIdentity}})
		Expect(err).ToNot(HaveOccurred())
		Expect(encrypted).ToNot(HaveLen(0))
		decrypted, err := bobSession.Decrypt(encrypted)
		Expect(err).ToNot(HaveOccurred())
		Expect(decrypted).To(Equal(clearData))
	})

	It("Encrypts then shares with bob", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		aliceSession, _ := aliceLaptop.Start()
		bob := TestApp.CreateUser()
		bobLaptop, _ := bob.CreateDevice()
		bobSession, _ := bobLaptop.Start()
		clearData := randomBytes(1024 * 1024 * 3)
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
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		aliceSession, _ := aliceLaptop.Start()
		bobEmail := "bob@gmail.com"
		bobProvisional, err := identity.CreateProvisional(TestApp.IdConfig, bobEmail)
		Expect(err).ToNot(HaveOccurred())
		clearData := randomBytes(12)
		bobPublicProvisional, err := identity.GetPublicIdentity(*bobProvisional)
		Expect(err).ToNot(HaveOccurred())
		// Trigger the creation of the provisional Identity on the Tanker Server
		_, err = aliceSession.Encrypt(clearData, &core.EncryptOptions{Recipients: []string{*bobPublicProvisional}})
		Expect(err).ToNot(HaveOccurred())
		bob := TestApp.CreateUser()
		bobLaptop, _ := bob.CreateDevice()
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
