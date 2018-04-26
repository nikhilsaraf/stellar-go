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
	data := "Hello World!"

	// sign it
	urlEncodedBase64Signature := sign(data, stellarPrivateKey)
	log.Println("url-encoded base64 signature:", urlEncodedBase64Signature)

	// verify the signature
	e := verify(data, urlEncodedBase64Signature, stellarPublicKey)
	if e != nil {
		log.Fatal(e)
	}
	log.Println("data is valid:", data)
}

// -------------------------------------------------------------------------
// ---------------------------- S I G N I N G ------------------------------
// -------------------------------------------------------------------------
func sign(data string, stellarPrivateKey string) string {
	kp := keypair.MustParse(stellarPrivateKey)

	// sign the data
	signatureBytes, e := kp.Sign([]byte(data))
	if e != nil {
		log.Fatal(e)
	}

	// encode the signature as base64
	base64Signature := base64.StdEncoding.EncodeToString(signatureBytes)

	// url-encode it
	urlEncodedBase64Signature := url.QueryEscape(base64Signature)
	return urlEncodedBase64Signature
}

// -------------------------------------------------------------------------
// ------------------------- V E R I F I C A T I O N -----------------------
// -------------------------------------------------------------------------
func verify(data string, urlEncodedBase64Signature string, stellarPublicKey string) error {
	kp := keypair.MustParse(stellarPublicKey)

	// decode the url-encoded signature
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
	return kp.Verify([]byte(data), signatureBytes)
}
