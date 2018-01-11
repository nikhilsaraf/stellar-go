package main

import (
    "fmt"
    "flag"
    "log"
    "github.com/stellar/go/keypair"
    "github.com/stellar/go/clients/horizon"
    b "github.com/stellar/go/build"
)

func getAddress(seed string) string {
    issuerKP, err := keypair.Parse(seed)
    if err != nil {
        log.Fatal(err)
    }
    return issuerKP.Address()
}

func main() {
    codePtr := flag.String("code", "", "the code for the asset")
    issuerAddressPtr := flag.String("issuer", "", "the issuer's address")
    receiverSeedPtr := flag.String("receiverSeed", "", "the receiver's seed")
    flag.Parse()

    if *codePtr == "" || *issuerAddressPtr == "" || *receiverSeedPtr == "" {
        flag.PrintDefaults()
        return
    }
    issuerAddress := *issuerAddressPtr
    receiverSeed := *receiverSeedPtr
    receiverAddress := getAddress(receiverSeed)
    code := *codePtr
    fmt.Println("code:", code)
    fmt.Println("issuerAddress:", issuerAddress)
    fmt.Println("receiverSeed:", receiverSeed)
    fmt.Println("receiverAddress:", receiverAddress)

    client := horizon.DefaultPublicNetClient
    loadAccount(client, issuerAddress)
    loadAccount(client, receiverAddress)

    txn := b.Transaction(
        b.SourceAccount{receiverAddress},
        b.AutoSequence{client},
        b.PublicNetwork,
        b.Trust(code, issuerAddress),
    )
    txnS := txn.Sign(receiverSeed)
    txn64, err := txnS.Base64()
    if err != nil {
        log.Fatal(err)
    }
    
    resp, errS := client.SubmitTransaction(txn64)
    if errS != nil {
        log.Fatal(errS)
    }
    fmt.Println("transaction posted in ledger:", resp.Ledger)
}

func loadAccount(client *horizon.Client, address string) horizon.Account {
    account, err := client.LoadAccount(address)
    if err != nil {
        log.Fatal(err)
    }
    return account
}
