package main

import (
    "flag"
    "fmt"
    "log"
    "net/url"
    "os"

    b "github.com/stellar/go/build"
    "github.com/stellar/go/clients/horizon"
    "github.com/stellar/go/xdr"
)

func main() {
    secretKey, partialTxnURLEncoded := parseInputs()

    // 1. decode the URL-encoded partial transaction
    unescapedTxn := unescape(partialTxnURLEncoded)

    // 2. decode the base64 XDR
    txn := decodeFromBase64(unescapedTxn)

    // 3. mutate the transaction inside the transaction envelope:
    //     a. update the source account
    //     b. set the sequence number
    //     c. set the network passphrase
    horizonClient := horizon.DefaultTestNetClient
    e := txn.MutateTX(
        // we assume that the accountId uses the master key
        &b.SourceAccount{AddressOrSeed: secretKey},
        &b.AutoSequence{SequenceProvider: horizonClient},
        // need to reset the network passphrase here
        b.TestNetwork,
    )
    if e != nil {
        log.Fatal(e)
    }

    // 4. sign the transaction envelope
    e = txn.Mutate(&b.Sign{Seed: secretKey})
    if e != nil {
        log.Fatal(e)
    }

    // 5. convert the transaction to base64
    reencodedTxnBase64, e := txn.Base64()
    if e != nil {
        log.Fatal("failed to convert to base64: ", e)
    }

    // 6. submit to the network
    resp, e := horizonClient.SubmitTransaction(reencodedTxnBase64)
    if e != nil {
        log.Fatal(e)
    }
    fmt.Println("transaction posted in ledger:", resp.Ledger)
}

// unescape decodes the URL-encoding
func unescape(partialTxnURLEncoded string) string {
    unescapedTxn, e := url.QueryUnescape(partialTxnURLEncoded)
    if e != nil {
        log.Fatal(e)
    }
    return unescapedTxn
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

// boilerplate to parse command line args and to make this implementation functional
func parseInputs() (secretKey string, partialTxnURLEncoded string) {
    // assumes that the signing account uses only the master key to sign transactions
    secretKeyPtr := flag.String("secretKey", "", "secret key to sign the transaction")
    txnPtr := flag.String("txn", "", "partial transaction to be signed and submitted")
    flag.Parse()

    if *secretKeyPtr == "" || *txnPtr == "" {
        flag.PrintDefaults()
        os.Exit(1)
    }

    secretKey = *secretKeyPtr
    partialTxnURLEncoded = *txnPtr
    return
}
