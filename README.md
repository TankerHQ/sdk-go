[license-badge]: https://img.shields.io/badge/License-Apache%202.0-blue.svg
[license-link]: https://opensource.org/licenses/Apache-2.0

[last-commit-badge]: https://img.shields.io/github/last-commit/TankerHQ/sdk-go.svg?label=Last%20commit&logo=github
[last-commit-link]: https://github.com/TankerHQ/sdk-go/commits/master

[![License][license-badge]][license-link]
[![Last Commit][last-commit-badge]][last-commit-link]

# Encryption SDKs for Go

<a href="#readme"><img src="https://tanker.io/images/github-logo.png" alt="Tanker logo" width="180" /></a>


[Overview](#overview) · [Core](#tanker-core) · [Identity](#identity-management) · [Other platforms](#other-platforms) · [Contributing](#contributing) · [License](#license)

## Overview

Tanker is an open-source solution to protect sensitive data in any application, with a simple end-user experience and good performance. No cryptographic skills are required to implement it.


## Tanker Core

Tanker **Core** is the foundation, it provides powerful **end-to-end encryption** of any type of data, textual or binary. Tanker **Core** handles multi-device, identity verification, user groups and pre-registration sharing.

<details><summary>Tanker Core usage example</summary>

The Core SDK takes care of all the difficult cryptography in the background, leaving you with simple high-level APIs.
The Core SDK automatically handles complex key exchanges, cryptographic operations, and identity verification for you.

You can copy/paste the following example:

```go
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/TankerHQ/sdk-go/v2/core"
)

const (
	AppID   = <your app id>
	AppURL  = "https://api.tanker.io"
	AuthURL = "https://fakeauth.tanker.io"
)

func base64ToUrlBase64(param string) (res string, err error) {
	bin, err := base64.StdEncoding.DecodeString(param)
	if err != nil {
		return
	}
	res = base64.URLEncoding.EncodeToString(bin)
	return
}

func GetIdentity() (identity string, err error) {
	urlAppID, err := base64ToUrlBase64(AppID)
	if err != nil {
		return
	}
	resp, err := http.Get(fmt.Sprintf("%s/apps/%s/disposable_private_identity", AuthURL, urlAppID))
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Cannot fetch identity from server '%s'", resp.Status)
		return
	}
	defer resp.Body.Close()
	bin, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var res map[string]string
	if err = json.Unmarshal(bin, &res); err != nil {
		return
	}
	if len(res["code"]) != 0 {
		err = fmt.Errorf("Failed to retrieve identity '%s', '%s'", res["code"], res["message"])
		return
	}
	identity = res["private_permanent_identity"]
	return
}

func main() {
	fmt.Println("Creating tanker ...")
	tankerOpts := core.TankerOptions{AppID: AppID, WritablePath: os.TempDir()}
	tanker, err := core.NewTanker(tankerOpts)
	if err != nil {
		log.Fatal("Could not create Tanker", err)
	}
	core.SetLogHandler(func(core.LogRecord) {})
	fmt.Println("Fetching identity ...")
	aliceIdentity, err := GetIdentity()
	if err != nil {
		log.Fatal("Could not get identity")
		return
	}

	fmt.Println("Starting tanker ...")
	status, err := tanker.Start(string(aliceIdentity))
	if err != nil {
		log.Fatal("Could not start tanker", err)
	}
	switch status {
	case core.StatusIdentityVerificationNeeded:
		err = tanker.VerifyIdentity(core.PassphraseVerification{"*******"})
	case core.StatusIdentityRegistrationNeeded:
		err = tanker.RegisterIdentity(core.PassphraseVerification{"*******"})
	}
	if err != nil {
		log.Fatal("Could not register identity:", err)
	}

	message := "This is my story"
	fmt.Println("Encrypting message ...")
	encrypted, err := tanker.Encrypt([]byte(message), nil)
	if err != nil {
		log.Fatal("Failed to encrypt message", err)
	}

	fmt.Println("Decrypting message ...")
	clearBytes, err := tanker.Decrypt(encrypted)
	if err != nil {
		log.Fatal("Failed to decrypt  message", err)
	}

	if clearText != message {
		log.Fatal("Unexpected decrypted message: got '%s', want '%s'", clearText, message)
	}

	fmt.Println("Success!")
}
```

Before running it, set the AppID with the one you have created on your [dashboard](https://dashboard.tanker.io).
You MUST enable the test mode for this example to work.

Then:
```bash
go build -o example-go && ./example-go
```

</details>

For more details and advanced examples, please refer to:

* [Core SDK implementation guide](https://docs.tanker.io/latest/guide/basic-concepts/)
* [Core API reference](https://docs.tanker.io/latest/api/tanker/)


## Identity management

End-to-end encryption requires that all users have cryptographic identities. The following packages help to handle them:

Tanker **Identity** is a server side package to link Tanker identities with your users in your application backend.
It is available in multiple languages. Check [identity-go](https://github.com/TankerHQ/identity-go) for more details, other implementation exists for different language.

## Contributing

We welcome feedback, [bug reports](https://github.com/TankerHQ/sdk-go/issues), and bug fixes in the form of [pull requests](https://github.com/TankerHQ/sdk-go/pulls).

## Other platforms

Tanker is also available for your **mobile applications**: use our open-source **[iOS](https://github.com/TankerHQ/sdk-ios)** and **[Android](https://github.com/TankerHQ/sdk-android)** SDKs.

## License

The Tanker Golang SDK is licensed under the [Apache License, version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

