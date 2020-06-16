package core_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/TankerHQ/identity-go/identity"
	"github.com/TankerHQ/sdk-go/v2/core"
	"github.com/TankerHQ/sdk-go/v2/helpers"
)

func getOidcIdToken(oidcConfig helpers.OidcConfig, userName string) (*string, error) {

	payload, err := json.Marshal(map[string]string{
		"client_id":     oidcConfig.ClientId,
		"client_secret": oidcConfig.ClientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": oidcConfig.Users[userName].RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	response, err := http.Post("https://www.googleapis.com/oauth2/v4/token", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	var result map[string]json.RawMessage
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	var id_token string
	if err = json.Unmarshal(result["id_token"], &id_token); err != nil {
		return nil, err
	}
	return &id_token, nil
}

func doVerification(tanker *core.Tanker, identity string, verification interface{}) (core.Status, error) {
	status, err := tanker.Start(identity)
	if err != nil {
		return 0, err
	}
	Expect(status).To(Equal(core.StatusIdentityVerificationNeeded))
	if err = tanker.VerifyIdentity(verification); err != nil {
		return 0, err
	}
	return tanker.GetStatus(), nil
}

func HaveVerificationMethods(methods ...core.VerificationMethod) types.GomegaMatcher {
	return SatisfyAll(
		HaveLen(len(methods)),
		ConsistOf(methods))
}

var _ = Describe("functional", func() {
	It("Generates a verification key", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, err := aliceLaptop.CreateSession()
		Expect(err).ToNot(HaveOccurred())
		Expect(session.Start(alice.Identity)).To(Equal(core.StatusIdentityRegistrationNeeded))
		Expect(session.GenerateVerificationKey()).ToNot(BeNil())
		defer session.Stop() // nolint: errCheck
	})

	It("Gets verification methods", func() {
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		session, _ := aliceLaptop.Start()
		defer session.Stop() // nolint: errCheck
		methods, err := session.GetVerificationMethods()
		Expect(err).ToNot(HaveOccurred())
		Expect(methods).To(HaveVerificationMethods(core.VerificationMethod{Type: core.VerificationMethodPassphrase}))
	})

	It("Set verification method", func() {
		alice := TestApp.CreateUser()
		aliceEmail := "alice.test@tanker.io"
		code, err := TestApp.GetVerificationCode(aliceEmail)
		Expect(err).ToNot(HaveOccurred())

		aliceLaptop, _ := alice.CreateDevice()
		session, _ := aliceLaptop.Start()
		defer session.Stop() // nolint: errCheck
		Expect(session.SetVerificationMethod(
			core.EmailVerification{Email: aliceEmail, VerificationCode: *code},
		)).To(Succeed())
		methods, err := session.GetVerificationMethods()
		Expect(err).ToNot(HaveOccurred())
		Expect(methods).To(HaveVerificationMethods(
			core.VerificationMethod{Type: core.VerificationMethodPassphrase},
			core.VerificationMethod{Type: core.VerificationMethodEmail, Email: &aliceEmail},
		))
	})

	Context("Martine and Kevin use oidc", func() {
		var (
			martine                 helpers.User
			martineLaptopDevice     *helpers.Device
			martinePhoneDevice      *helpers.Device
			martineLaptop           *core.Tanker
			martinePhone            *core.Tanker
			martineOidcVerification core.OidcVerification
			kevinOidcVerification   core.OidcVerification
		)

		BeforeEach(func() {
			var err error
			if err = TestApp.EnableOidc(); err != nil {
				panic(err)
			}
			martine = TestApp.CreateUser()

			martineLaptopDevice, _ = martine.CreateDevice()
			martineLaptop, _ = martineLaptopDevice.CreateSession()
			Expect(martineLaptop.Start(martine.Identity)).To(Equal(core.StatusIdentityRegistrationNeeded))

			martinePhoneDevice, _ = martine.CreateDevice()
			martinePhone, _ = martinePhoneDevice.CreateSession()

			martineIdToken, err := getOidcIdToken(TestApp.OidcConfig, "martine")
			Expect(err).ToNot(HaveOccurred())
			martineOidcVerification = core.OidcVerification{*martineIdToken}
			kevinIdToken, err := getOidcIdToken(TestApp.OidcConfig, "kevin")
			Expect(err).ToNot(HaveOccurred())
			kevinOidcVerification = core.OidcVerification{*kevinIdToken}
		})

		AfterEach(func() {
			martineLaptop.Stop() // nolint: errCheck
			martinePhone.Stop()  // nolint: errCheck
		})

		It("Registers and verifies identity with an oidc id token", func() {
			Expect(martineLaptop.RegisterIdentity(martineOidcVerification)).To(Succeed())
			Expect(doVerification(martinePhone, martine.Identity, martineOidcVerification)).To(Equal(core.StatusReady))
		})

		It("Fails to verify a valid token for the wrong user", func() {
			Expect(martineLaptop.RegisterIdentity(martineOidcVerification)).To(Succeed())
			_, err := doVerification(martinePhone, martine.Identity, kevinOidcVerification)
			Expect(err).To(HaveOccurred())
		})

		It("Updates and verifies with an oidc token", func() {
			Expect(martineLaptop.RegisterIdentity(core.PassphraseVerification{"*****"})).To(Succeed())
			Expect(martineLaptop.SetVerificationMethod(martineOidcVerification)).To(Succeed())
			Expect(martinePhone.Start(martine.Identity)).To(Equal(core.StatusIdentityVerificationNeeded))
			Expect(martinePhone.VerifyIdentity(martineOidcVerification)).To(Succeed())
			methods, err := martinePhone.GetVerificationMethods()
			Expect(err).ToNot(HaveOccurred())
			Expect(methods).To(HaveVerificationMethods(
				core.VerificationMethod{Type: core.VerificationMethodPassphrase},
				core.VerificationMethod{Type: core.VerificationMethodOidcIdToken},
			))
		})

		It("Decrypts data shared with an attached proisional identity", func() {
			alice := TestApp.CreateUser()
			aliceDevice, _ := alice.CreateDevice()
			aliceLaptop, _ := aliceDevice.Start()
			defer aliceLaptop.Stop() // nolint: errCheck

			Expect(martineLaptop.RegisterIdentity(core.PassphraseVerification{"*****"})).To(Succeed())
			martineEmail := TestApp.OidcConfig.Users["martine"].Email
			martineProvisionalIdentity, _ := identity.CreateProvisional(TestApp.IdConfig, martineEmail)
			martinePublicIdentity, _ := identity.GetPublicIdentity(*martineProvisionalIdentity)
			clearText := helpers.RandomBytes(15)
			encrypted, _ := aliceLaptop.Encrypt(clearText, &core.EncryptionOptions{Recipients: []string{*martinePublicIdentity}})

			attachResult, err := martineLaptop.AttachProvisionalIdentity(*martineProvisionalIdentity)
			Expect(err).ToNot(HaveOccurred())
			Expect(attachResult.Status).To(Equal(core.StatusIdentityVerificationNeeded))
			Expect(attachResult.Method.Type).To(Equal(core.VerificationMethodEmail))
			Expect(attachResult.Method.Email).To(Equal(&martineEmail))
			Expect(martineLaptop.VerifyProvisionalIdentity(martineOidcVerification)).To(Succeed())
			decrypted, err := martineLaptop.Decrypt(encrypted)
			Expect(err).ToNot(HaveOccurred())
			Expect(decrypted).To(Equal(clearText))
		})
	})
})
