package core_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
	"github.com/TankerHQ/sdk-go/v2/helpers"
)

var _ = Describe("functional", func() {

	It("Creates a group", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		aliceSession, _ := aliceLaptop.Start()
		martine := TestApp.CreateUser()
		martineLaptop, _ := martine.CreateDevice()
		martineSession, _ := martineLaptop.Start()
		bobProvisional, err := identity.CreateProvisional(TestApp.IdConfig, "bob@tanker.io")
		Expect(err).ToNot(HaveOccurred())
		bobPublicProvisional, err := identity.GetPublicIdentity(*bobProvisional)
		Expect(err).ToNot(HaveOccurred())
		groupID, err := aliceSession.CreateGroup([]string{alice.PublicIdentity, *bobPublicProvisional, martine.PublicIdentity})
		Expect(err).ToNot(HaveOccurred())
		Expect(*groupID).ToNot(BeEmpty())

		clearData := helpers.RandomBytes(1024 * 1024 * 3)
		encrypted, err := aliceSession.Encrypt(clearData, &core.EncryptionOptions{ShareWithGroups: []string{*groupID}})
		Expect(err).ToNot(HaveOccurred())

		decrypted, err := martineSession.Decrypt(encrypted)
		Expect(err).ToNot(HaveOccurred())
		Expect(decrypted).To(Equal(clearData))

		Expect(martineSession.Destroy()).To(Succeed())
		Expect(aliceSession.Destroy()).To(Succeed())
	})

	It("Updates a group", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		aliceSession, _ := aliceLaptop.Start()
		groupID, err := aliceSession.CreateGroup([]string{alice.PublicIdentity})
		Expect(err).ToNot(HaveOccurred())

		martine := TestApp.CreateUser()
		martineLaptop, _ := martine.CreateDevice()
		martineSession, _ := martineLaptop.Start()
		bobProvisional, err := identity.CreateProvisional(TestApp.IdConfig, "bob@tanker.io")
		Expect(err).ToNot(HaveOccurred())
		bobPublicProvisional, err := identity.GetPublicIdentity(*bobProvisional)
		Expect(err).ToNot(HaveOccurred())
		Expect(aliceSession.UpdateGroupMembers(*groupID, []string{martine.PublicIdentity, *bobPublicProvisional})).To(Succeed())

		clearData := helpers.RandomBytes(1024 * 1024 * 3)
		encrypted, err := aliceSession.Encrypt(clearData, &core.EncryptionOptions{ShareWithGroups: []string{*groupID}})
		Expect(err).ToNot(HaveOccurred())

		decrypted, err := martineSession.Decrypt(encrypted)
		Expect(err).ToNot(HaveOccurred())
		Expect(decrypted).To(Equal(clearData))

		Expect(martineSession.Destroy()).To(Succeed())
		Expect(aliceSession.Destroy()).To(Succeed())
	})

})
