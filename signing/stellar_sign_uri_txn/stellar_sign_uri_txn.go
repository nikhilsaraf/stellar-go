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
    "flag"
    "fmt"
    "log"
    "net/url"
    "os"

    b "github.com/stellar/go/build"
    "github.com/stellar/go/clients/horizon"
    "github.com/stellar/go/xdr"
    kp "github.com/stellar/go/keypair"
)

var emptyAddress = kp.Master("").Address()

func main() {
    secretKey, uriString := parseInputs()

    // 1. extract the URL-encoded xdr string from URI
    uri, e := url.ParseRequestURI(uriString)
    if e != nil {
        log.Fatal(e)
    }
    encodedInputTxn := uri.Query().Get("xdr")

    // 2. decode the Transaction
    unescapedTxn := unescape(encodedInputTxn)

    // 3. decode the base64 XDR
    txn := decodeFromBase64(unescapedTxn)

    // 4. check the source account and mutate the transaction inside the transaction envelope if needed:
    //     a. update the source account
    //     b. set the sequence number
    //     c. set the network passphrase
    horizonClient := horizon.DefaultTestNetClient
    if txn.E.Tx.SourceAccount.Address() == emptyAddress {
        e = txn.MutateTX(
            // we assume that the accountID uses the master key, this can also be the accountID
            &b.SourceAccount{AddressOrSeed: secretKey},
            &b.AutoSequence{SequenceProvider: horizonClient},
            // need to reset the network passphrase
            b.TestNetwork,
        )
        if e != nil {
            log.Fatal(e)
        }
    } else if txn.E.Tx.SeqNum == 0 {
        e = txn.MutateTX(
            // do not need to set the source account here, only the sequence number
            &b.AutoSequence{SequenceProvider: horizonClient},
            // need to reset the network passphrase
            b.TestNetwork,
        )
        if e != nil {
            log.Fatal(e)
        }
    }

    // 5. sign the transaction envelope
    e = txn.Mutate(&b.Sign{Seed: secretKey})
    if e != nil {
        log.Fatal(e)
    }

    // 6. convert the transaction to base64
    reencodedTxnBase64, e := txn.Base64()
    if e != nil {
        log.Fatal("failed to convert to base64: ", e)
    }

    // 7. submit to the network
    resp, e := horizonClient.SubmitTransaction(reencodedTxnBase64)
    if e != nil {
        log.Fatal(e)
    }
    fmt.Println("transaction posted in ledger:", resp.Ledger)
}

// unescape decodes the URL-encoded and base64 encoded txn
func unescape(escaped string) string {
    unescaped, e := url.QueryUnescape(escaped)
    if e != nil {
        log.Fatal(e)
    }
    return unescaped
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
func parseInputs() (secretKey string, uriString string) {
    // assumes that the signing account uses only the master key to sign transactions
    secretKeyPtr := flag.String("secretKey", "", "secret key to sign the transaction")
    uriPtr := flag.String("uri", "", "URI Request that contains the XDR Transaction to be signed and submitted, only supports a limited set of operations for SEP7")
    flag.Parse()

    if *secretKeyPtr == "" || *uriPtr == "" {
        fmt.Println("Params:")
        flag.PrintDefaults()
        os.Exit(1)
    }
    return *secretKeyPtr, *uriPtr
}
