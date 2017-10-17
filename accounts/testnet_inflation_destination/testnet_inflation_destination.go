package main

import (
    "fmt"
    "flag"
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
    addressPtr := flag.String("a", "", "string representing the inflation destination address to be used")
    flag.Parse()

    inflationAddress := *addressPtr
    fmt.Println("inflation destination:", inflationAddress)

    horizonClient := &horizon.Client{
        URL:  baseUrl,
        HTTP: http.DefaultClient,
    }
    loadAccount(horizonClient, inflationAddress, "inflation address")
    log.Println()

    reader := bufio.NewReader(os.Stdin)
    secret, _ := reader.ReadString('\n')
    secret = strings.Replace(secret, "\n", "", -1)
    fmt.Println("\nreceived secret key, setting inflation destination now.\n")

    txn := b.Transaction(
        b.SourceAccount{secret},
        b.AutoSequence{horizonClient},
        b.TestNetwork,
        b.SetOptions(
            b.InflationDest(inflationAddress),
        ),
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

func loadAccount(horizonClient *horizon.Client, publicKey string, accountName string) horizon.Account {
    account, err := horizonClient.LoadAccount(publicKey)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Balances for account (" + accountName + "):")
    for _, balance := range account.Balances {
        log.Println("   ", balance)
    }
    return account
}
