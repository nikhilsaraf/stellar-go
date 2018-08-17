package main

import (
    "fmt"
    "flag"
    "log"
    "github.com/stellar/go/clients/horizon"
    "bufio"
    "os"
    "strings"
    b "github.com/stellar/go/build"
)

func main() {
    addressPtr := flag.String("a", "", "string representing the inflation destination address to be used")
    usePublicPtr := flag.Bool("p", false, "use the public network (defaults to test network")
    flag.Parse()

    inflationAddress := *addressPtr
    horizonClient := horizon.DefaultTestNetClient
    net := b.TestNetwork
    if *usePublicPtr {
        horizonClient = horizon.DefaultPublicNetClient
        net = b.PublicNetwork
    }
    fmt.Println("inflation destination:", inflationAddress)
    fmt.Println("network:", net)

    if inflationAddress != "" {
        loadAccount(horizonClient, inflationAddress, "inflation address")
    }
    log.Println()

    fmt.Print("Enter secret key: ")
    reader := bufio.NewReader(os.Stdin)
    secret, _ := reader.ReadString('\n')
    secret = strings.Replace(secret, "\n", "", -1)
    fmt.Println("\nreceived secret key, setting inflation destination now.\n")

    txn, err := b.Transaction(
        b.SourceAccount{secret},
        b.AutoSequence{horizonClient},
        net,
        b.SetOptions(
            b.InflationDest(inflationAddress),
        ),
    )
    if err != nil {
        log.Fatal(err)
    }
    // sign
    txnS, err := txn.Sign(secret)
    if err != nil {
        log.Fatal(err)
    }
    // convert to base64
    txnS64, err := txnS.Base64()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("tx base64: %s\n\n", txnS64)

    // submit the transaction
    resp, err := horizonClient.SubmitTransaction(txnS64)
    if err != nil {
        log.Fatal(err)
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
