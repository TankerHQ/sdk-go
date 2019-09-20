[![License][license-badge]][license-link]
[![Last Commit][last-commit-badge]][last-commit-link]

# Encryption SDKs for Go

<a href="#readme"><img src="./src/public/tanker.png" alt="Tanker logo" width="180" /></a>


[Overview](#overview) · [Core](#tanker-core) · [Identity](#identity-management) · [Other platforms](#other-platforms) · [Contributing](#contributing) · [License](#license)

## Overview

Tanker is an open-source solution to protect sensitive data in any application, with a simple end-user experience and good performance. No cryptographic skills are required to implement it.


## Tanker Core

Tanker **Core** is the foundation, it provides powerful **end-to-end encryption** of any type of data, textual or binary. Tanker **Core** handles multi-device, identity verification, user groups and pre-registration sharing.

<details><summary>Tanker Core usage example</summary>

The Core SDK takes care of all the difficult cryptography in the background, leaving you with simple high-level APIs:

```go
import (
  Tanker "github.com/TankerHQ/sdk-go/core"
)
// FIXME
```

The Core SDK automatically handles complex key exchanges, cryptographic operations, and identity verification for you.
</details>

For more details and advanced examples, please refer to:

* [Core SDK implementation guide](https://docs.tanker.io/latest/guide/basic-concepts/)
* [Core API reference](https://docs.tanker.io/latest/api/tanker/)


## Identity management

End-to-end encryption requires that all users have cryptographic identities. The following packages help to handle them:

Tanker **Identity** is a server side package to link Tanker identities with your users in your application backend.
It is available in multiple languages. Check "github.com/TankerHQ/identity-go" for more details, other implementation exists for different language.

## Contributing

We welcome feedback, [bug reports](https://github.com/TankerHQ/sdk-go/issues), and bug fixes in the form of [pull requests](https://github.com/TankerHQ/sdk-go/pulls).

## Other platforms

Tanker is also available for your **mobile applications**: use our open-source **[iOS](https://github.com/TankerHQ/sdk-ios)** and **[Android](https://github.com/TankerHQ/sdk-android)** SDKs.

## License

The Tanker Golang SDK is licensed under the [Apache License, version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

