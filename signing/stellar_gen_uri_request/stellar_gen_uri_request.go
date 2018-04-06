package main

import (
    "flag"
    "fmt"
    "log"
    "net/url"
    "os"
    "strings"

    b "github.com/stellar/go/build"
    kp "github.com/stellar/go/keypair"
)

func main() {
    destinationAddress, memo, creditAmount := parseInputs()

    // 1. build the partial transaction (excludes the source account and sequence number)
    emptyAddress := kp.Master("").Address()
    txn, e := b.Transaction(
        // since the address is the empty sentinel value, the wallet will need to fill it in along with the sequence number
        b.SourceAccount{AddressOrSeed: emptyAddress},
        // meaningless to have a sequence number here since the source account is the empty address and will be replaced by the wallet
        b.TestNetwork,
        b.Payment(
            b.Destination{AddressOrSeed: destinationAddress},
            creditAmount,
        ),
    )
    if e != nil {
        log.Fatal(e)
    }
    if memo != "" {
        e = txn.Mutate(b.MemoText{Value: memo})
        if e != nil {
            log.Fatal(e)
        }
    }

    // 2. sign with empty signature so it gets converted to a transaction envelope
    txnE, e := txn.Sign()
    if e != nil {
        log.Fatal("failed to sign: ", e)
    }

    // 3. convert to base64
    txnB64, e := txnE.Base64()
    if e != nil {
        log.Fatal("failed to convert to base64: ", e)
    }

    // 4. url encode
    urlEncoded := url.QueryEscape(txnB64)

    fmt.Println("stellar://txn/" + urlEncoded)
}

// boilerplate to parse command line args and to make this implementation functional
func parseInputs() (destinationAddress string, memo string, creditAmount b.PaymentMutator) {
    toAddressPtr := flag.String("toAddress", "", "destination address")
    amountPtr := flag.Float64("amount", 0.0, "amount to be sent, must be > 0.0")
    memoPtr := flag.String("memo", "", "(optional) memo to include with the payment")
    assetPtr := flag.String("asset", "", "(optional) asset to pay with, of the form code:issuer")
    flag.Parse()

    if *toAddressPtr == "" || *amountPtr <= 0 {
        fmt.Println("Params:")
        flag.PrintDefaults()
        os.Exit(1)
    }

    amountStr := fmt.Sprintf("%v", *amountPtr)
    if *assetPtr != "" {
        assetParts := strings.SplitN(*assetPtr, ":", 2)
        issuerAddress := assetParts[1]
        creditAmount = b.CreditAmount{Code: assetParts[0], Issuer: issuerAddress, Amount: amountStr}
    } else {
        creditAmount = b.NativeAmount{Amount: amountStr}
    }

    return *toAddressPtr, *memoPtr, creditAmount
}
