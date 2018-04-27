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
    "fmt"
    "log"
    "bufio"
    "os"
    "strings"

    b "github.com/stellar/go/build"
    "github.com/stellar/go/xdr"
)

// sample reference implementation to collate signatures using multiple signed transactions for a multi-signature coordination service
func main() {
    xdrList := []string{}

    fmt.Printf("enter the first signed base64-encoded transaction xdr:\n")
    for {
        reader := bufio.NewReader(os.Stdin)
        tx, _ := reader.ReadString('\n')
        tx = strings.Replace(tx, "\n", "", -1)
        if len(tx) == 0 {
            fmt.Printf("received empty tx xdr, done entering transactions.\n")
            break
        }

        xdrList = append(xdrList, tx)
        fmt.Printf("\nenter the next signed base64-encoded transaction xdr (enter to continue):\n")
    }

    combinedTx := collate(xdrList)
    fmt.Printf("\n\ncollated transaction:\n%s\n", combinedTx)
}

// collate takes the list of base64-encoded transaction XDRs and combines the signatures to produce a single transaction XDR.
// in order to combine signatures, collate needs to verify that each transaction is the same.
func collate(xdrList []string) string {
    // we will use collated to collate all the transactions
    collated := decodeFromBase64(xdrList[0])

    for _, xdr := range xdrList[1:] {
        tx := decodeFromBase64(xdr)
        // implementations should take precautions before combining signatures, including but not limited to: deduping signatures, verifying signatures, checking that the transactions are the same, etc.
        collated.E.Signatures = append(collated.E.Signatures, tx.E.Signatures...)
    }

    collatedXdr, e := collated.Base64()
    if e != nil {
        log.Fatal("failed to convert to base64:", e)
    }
    return collatedXdr
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
