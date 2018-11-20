package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	b "github.com/stellar/go/build"
	"github.com/stellar/go/xdr"
	"golang.org/x/crypto/ssh/terminal"
)

type inputs struct {
	xdr       string
	secretKey string
	network   b.Network
}

func main() {
	fmt.Printf("====================================================================================================\n")
	ip := parseInputs()

	// decode the base64 XDR
	txn := decodeFromBase64(ip.xdr)

	fmt.Printf("setting the network passphrase to '%s'...", ip.network.Passphrase)
	e := txn.MutateTX(
		ip.network,
	)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("done.\n")

	fmt.Printf("signing the transaction...")
	e = txn.Mutate(&b.Sign{Seed: ip.secretKey})
	if e != nil {
		log.Fatal(e)
	}
	fmt.Printf("done.\n")

	fmt.Printf("converting the signed XDR to base64...")
	signedBase64Tx, e := txn.Base64()
	if e != nil {
		log.Fatal("failed to convert to base64: ", e)
	}
	fmt.Printf("done.\n")

	fmt.Printf("\n----------------------------------------------------------------------------------------------------\n")
	fmt.Printf("original XDR:\n")
	fmt.Printf("%s\n", ip.xdr)

	fmt.Printf("\nsignedBase64Tx:\n")
	fmt.Printf("%s\n", signedBase64Tx)

	fmt.Printf("\nurl-encoded signedBase64Tx:\n")
	urlEncoded := url.QueryEscape(signedBase64Tx)
	fmt.Printf("%s\n", urlEncoded)

	fmt.Printf("\nsubmit command:\n")
	horizonBaseURL := "https://horizon-testnet.stellar.org"
	if ip.network == b.PublicNetwork {
		horizonBaseURL = "https://horizon.stellar.org"
	}
	fmt.Printf("curl -X POST \"%s/transactions\" -d \"tx=%s\"\n", horizonBaseURL, urlEncoded)
	fmt.Printf("====================================================================================================\n")
}

// decodeFromBase64 decodes the transaction from a base64 string into a TransactionEnvelopeBuilder
func decodeFromBase64(encodedXdr string) *b.TransactionEnvelopeBuilder {
	// Unmarshall from base64 encoded XDR format
	var decoded xdr.TransactionEnvelope
	e := xdr.SafeUnmarshalBase64(encodedXdr, &decoded)
	if e != nil {
		log.Fatal(e)
	}

	// convert to TransactionEnvelopeBuilder
	txEnvelopeBuilder := b.TransactionEnvelopeBuilder{E: &decoded}
	txEnvelopeBuilder.Init()

	return &txEnvelopeBuilder
}

// returns the input XDR that needs to be signed
func parseInputs() inputs {
	xdrPtr := flag.String("xdr", "", "base-64 encoded XDR to be signed")
	flag.Parse()

	if *xdrPtr == "" {
		fmt.Println("Params:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter secret key: ")
	secret, e := terminal.ReadPassword(0)
	if e != nil {
		log.Fatal(e)
	}
	fmt.Println()
	fmt.Printf("Which network (t/p)? [t]: ")
	networkChoice, _ := reader.ReadString('\n')
	networkChoice = strings.Replace(networkChoice, "\n", "", -1)
	network := b.TestNetwork
	if networkChoice == "p" {
		network = b.PublicNetwork
	}

	return inputs{
		xdr:       *xdrPtr,
		secretKey: string(secret),
		network:   network,
	}
}
