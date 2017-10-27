package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/stellar/go/clients/horizon"
    "bufio"
    "os"
    "strings"
    b "github.com/stellar/go/build"
)

const baseUrl = "https://horizon-testnet.stellar.org"

func main() {
    reader := bufio.NewReader(os.Stdin)
    secret, _ := reader.ReadString('\n')
    secret = strings.Replace(secret, "\n", "", -1)
    fmt.Println("\nreceived secret key, setting inflation destination now.\n")

    horizonClient := &horizon.Client{
        URL:  baseUrl,
        HTTP: http.DefaultClient,
    }
    txn := b.Transaction(
        b.SourceAccount{secret},
        b.AutoSequence{horizonClient},
        b.TestNetwork,
        b.Inflation(),
    )
    // sign
    txnS := txn.Sign(secret)
    // convert to base64
    txnS64, err := txnS.Base64()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("tx base64: %s\n\n", txnS64)

    // submit the transaction
    resp, err2 := horizonClient.SubmitTransaction(txnS64)
    if err2 != nil {
        log.Fatal(err2)
    }
    fmt.Println("transaction posted in ledger:", resp.Ledger)
}
