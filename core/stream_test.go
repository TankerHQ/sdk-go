package core_test

import (
	"bytes"
	"crypto/sha1"
	"io"
	"io/ioutil"

	"github.com/TankerHQ/sdk-go/v2/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func createBigStream(size int) *bytes.Reader {
	return bytes.NewReader(helpers.RandomBytes(size))
}

var _ = Describe("Streams", func() {
	It("Encrypts and decrypts with stream", func() {
		source := createBigStream((1024 * 1024 * 3))
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		aliceSession, _ := aliceLaptop.Start()
		defer aliceSession.Stop() // nolint: errCheck
		encryptedStream, err := aliceSession.StreamEncrypt(source, nil)
		Expect(err).ToNot(HaveOccurred())
		decryptedStream, err := aliceSession.StreamDecrypt(encryptedStream)
		Expect(err).ToNot(HaveOccurred())
		bdecrypted, err := ioutil.ReadAll(decryptedStream)
		Expect(err).ToNot(HaveOccurred())
		Expect(source.Seek(0, io.SeekStart)).To(Equal(int64(0)))
		bsource, err := ioutil.ReadAll(source)
		// err is nil when sucessful, we do not want EOF
		Expect(err).To(BeNil())
		Expect(len(bsource)).To(Equal(len(bdecrypted)))
		Expect(sha1.Sum(bdecrypted)).To(Equal(sha1.Sum(bsource)))
		encryptedStream.Destroy()
		decryptedStream.Destroy()
	})

	It("Retrieves the ID of a stream", func() {
		source := createBigStream((1024 * 1024 * 3))
		alice := TestApp.CreateUser()
		aliceLaptop, _ := alice.CreateDevice()
		aliceSession, _ := aliceLaptop.Start()
		defer aliceSession.Stop() // nolint: errCheck
		encryptedStream, err := aliceSession.StreamEncrypt(source, nil)
		Expect(err).ToNot(HaveOccurred())
		id, err := encryptedStream.GetResourceID()
		Expect(err).ToNot(HaveOccurred())
		Expect(*id).ToNot(BeEmpty())
		encryptedStream.Destroy()
	})
})
