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
    networkPtr := flag.String("network", "", "t for testnet, p for pubnet")
    limitPtr := flag.Int("limit", 0, "(optional) limit for trust, 0 for max limit")
    flag.Parse()

    if *codePtr == "" || *issuerAddressPtr == "" || *receiverSeedPtr == "" || (*networkPtr != "t" && *networkPtr != "p") {
        flag.PrintDefaults()
        return
    }
    code := *codePtr
    issuerAddress := *issuerAddressPtr
    receiverSeed := *receiverSeedPtr
    receiverAddress := getAddress(receiverSeed)
    network := *networkPtr
    limit := *limitPtr
    fmt.Println("code:", code)
    fmt.Println("issuerAddress:", issuerAddress)
    fmt.Println("receiverSeed:", receiverSeed)
    fmt.Println("receiverAddress:", receiverAddress)
    fmt.Println("network:", network)
    fmt.Println("limit:", limit)

    var client *horizon.Client
    var net b.Network
    if network == "p" {
        client = horizon.DefaultPublicNetClient
        net = b.PublicNetwork
        fmt.Println("using pubnet")
    } else {
        client = horizon.DefaultTestNetClient
        net = b.TestNetwork
        fmt.Println("using testnet")
    }
    
    trust := b.Trust(code, issuerAddress)
    if limit > 0 {
        trustAmount := fmt.Sprintf("%d", limit)
        fmt.Println("setting trust amount:", trustAmount)
        trust = b.Trust(code, issuerAddress, b.Limit(trustAmount))
    }

    loadAccount(client, issuerAddress)
    loadAccount(client, receiverAddress)
    txn, err := b.Transaction(
        b.SourceAccount{receiverAddress},
        b.AutoSequence{client},
        net,
        trust,
    )
    if err != nil {
        log.Fatal(err)
    }

    txnS, err := txn.Sign(receiverSeed)
    if err != nil {
        log.Fatal(err)
    }

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
