/*
Copyright 2018 Lightyear.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/base64"
	"log"
	"net/url"

	"github.com/stellar/go/keypair"
)

func main() {
	const stellarPrivateKey = "SBPOVRVKTTV7W3IOX2FJPSMPCJ5L2WU2YKTP3HCLYPXNI5MDIGREVNYC"
	const stellarPublicKey = "GD7ACHBPHSC5OJMJZZBXA7Z5IAUFTH6E6XVLNBPASDQYJ7LO5UIYBDQW"

	// data is your payload
	data := "web+stellar:pay?destination=GCALNQQBXAPZ2WIRSDDBMSTAKCUH5SG6U76YBFLQLIXJTF7FE5AX7AOO&amount=120.1234567&memo=skdjfasf&msg=pay%20me%20with%20lumens&origin_domain=someDomain.com"

	// sign it
	urlEncodedBase64Signature := sign(data, stellarPrivateKey)
	log.Println("url-encoded base64 signature:", urlEncodedBase64Signature)

	// verify the signature
	e := verify(data, urlEncodedBase64Signature, stellarPublicKey)
	if e != nil {
		log.Fatal(e)
	}
	log.Println("data is valid")

	// append signature to original URI request
	log.Printf("signed URI request: %s&signature=%s\n", data, urlEncodedBase64Signature)
}

// -------------------------------------------------------------------------
// ---------------------------- P A Y L O A D ------------------------------
// -------------------------------------------------------------------------
func constuctPayload(data string) []byte {
	// prefix 4 to denote application-based signing using 36 bytes
	var prefixSelectorBytes [36]byte
	prefixSelectorBytes = [36]byte{}
	prefixSelectorBytes[35] = 4

	// standardized namespace prefix for this signing use case
	prefix := "stellar.sep.7 - URI Scheme"

	// variable number of bytes for the prefix + data
	var uriWithPrefixBytes []byte
	uriWithPrefixBytes = []byte(prefix + data)

	var result []byte
	result = append(result, prefixSelectorBytes[:]...) // 36 bytes
	result = append(result, uriWithPrefixBytes[:]...)  // variable length bytes
	return result
}

// -------------------------------------------------------------------------
// ---------------------------- S I G N I N G ------------------------------
// -------------------------------------------------------------------------
func sign(data string, stellarPrivateKey string) string {
	// construct the payload
	payloadBytes := constuctPayload(data)

	// sign the data
	kp := keypair.MustParse(stellarPrivateKey)
	signatureBytes, e := kp.Sign(payloadBytes)
	if e != nil {
		log.Fatal(e)
	}

	// encode the signature as base64
	base64Signature := base64.StdEncoding.EncodeToString(signatureBytes)
	log.Println("base64 signature:", base64Signature)

	// url-encode it
	urlEncodedBase64Signature := url.QueryEscape(base64Signature)
	return urlEncodedBase64Signature
}

// -------------------------------------------------------------------------
// ------------------------- V E R I F I C A T I O N -----------------------
// -------------------------------------------------------------------------
func verify(data string, urlEncodedBase64Signature string, stellarPublicKey string) error {
	// construct the payload so we can verify it
	payloadBytes := constuctPayload(data)

	// decode the url-encoded signature
	kp := keypair.MustParse(stellarPublicKey)
	base64Signature, e := url.QueryUnescape(urlEncodedBase64Signature)
	if e != nil {
		log.Fatal(e)
	}

	// decode the base64 signature
	signatureBytes, e := base64.StdEncoding.DecodeString(base64Signature)
	if e != nil {
		log.Fatal(e)
	}

	// validate it against the public key
	return kp.Verify(payloadBytes, signatureBytes)
}
