package core_test

import (
	"bytes"
	"io/ioutil"

	"github.com/TankerHQ/sdk-go/v2/core"
	"github.com/TankerHQ/sdk-go/v2/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encryption sessions", func() {

	var (
		alice       helpers.User
		aliceLaptop *helpers.Device
		bob         helpers.User
		bobLaptop   *helpers.Device
	)

	BeforeEach(func() {
		alice = TestApp.CreateUser()
		aliceLaptop, _ = alice.CreateDevice()
		bob = TestApp.CreateUser()
		bobLaptop, _ = bob.CreateDevice()
	})

	It("Resource ID of the session matches the ciphertext", func() {
		msg := []byte("Reston - Court House")
		aliceSession, _ := aliceLaptop.Start()
		encSess, err := aliceSession.CreateEncryptionSession(nil)
		Expect(err).ToNot(HaveOccurred())
		defer aliceSession.Stop() // nolint: errcheck
		encrypted, err := encSess.Encrypt(msg)
		Expect(err).ToNot(HaveOccurred())

		sessId := encSess.GetResourceId()
		cipherId, err := aliceSession.GetResourceId(encrypted)
		Expect(err).ToNot(HaveOccurred())
		Expect(sessId).To(Equal(*cipherId))
	})

	It("Can share with a user using an encryption session", func() {
		msg := []byte("New Carollton Amtrak Station")
		aliceSession, _ := aliceLaptop.Start()
		bobSession, _ := bobLaptop.Start()
		defer aliceSession.Stop() // nolint: errcheck
		defer bobSession.Stop()   // nolint: errcheck

		encSess, err := aliceSession.CreateEncryptionSession(&core.EncryptionOptions{ShareWithUsers: []string{bob.PublicIdentity}})
		Expect(err).ToNot(HaveOccurred())
		encrypted, err := encSess.Encrypt(msg)
		Expect(err).ToNot(HaveOccurred())

		decrypted, err := bobSession.Decrypt(encrypted)
		Expect(err).ToNot(HaveOccurred())
		Expect(msg).To(Equal(decrypted))
	})

	It("Can share with a group using an encryption session", func() {
		msg := []byte("Penn Station Light Rail")
		aliceSession, _ := aliceLaptop.Start()
		bobSession, _ := bobLaptop.Start()
		defer aliceSession.Stop() // nolint: errcheck
		defer bobSession.Stop()   // nolint: errcheck

		groupId, err := aliceSession.CreateGroup([]string{bob.PublicIdentity})
		Expect(err).ToNot(HaveOccurred())
		encSess, err := aliceSession.CreateEncryptionSession(&core.EncryptionOptions{ShareWithGroups: []string{*groupId}})
		Expect(err).ToNot(HaveOccurred())
		encrypted, err := encSess.Encrypt(msg)
		Expect(err).ToNot(HaveOccurred())

		decrypted, err := bobSession.Decrypt(encrypted)
		Expect(err).ToNot(HaveOccurred())
		Expect(msg).To(Equal(decrypted))
	})

	It("Can encrypt streams with an encryption session", func() {
		msg := []byte("Camden Yards")
		sourceStream := bytes.NewReader(msg)
		aliceSession, _ := aliceLaptop.Start()
		bobSession, _ := bobLaptop.Start()
		defer aliceSession.Stop() // nolint: errcheck
		defer bobSession.Stop()   // nolint: errcheck

		encSess, err := aliceSession.CreateEncryptionSession(&core.EncryptionOptions{ShareWithUsers: []string{bob.PublicIdentity}})
		Expect(err).ToNot(HaveOccurred())
		encrypted, err := encSess.StreamEncrypt(sourceStream)
		Expect(err).ToNot(HaveOccurred())

		decryptedStream, err := bobSession.StreamDecrypt(encrypted)
		Expect(err).ToNot(HaveOccurred())
		decrypted, err := ioutil.ReadAll(decryptedStream)
		Expect(err).ToNot(HaveOccurred())
		Expect(msg).To(Equal(decrypted))
	})
})
