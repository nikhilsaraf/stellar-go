package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "github.com/stellar/go/clients/horizon"
)

const baseUrlDefault = "https://horizon-testnet.stellar.org"
const baseUrlLocal = "http://localhost:8000"

func main() {
    localPtr := flag.Bool("l", false, "boolean representing whether we should use the local horizon server @ " + baseUrlLocal)
    addressPtr := flag.String("a", "", "string representing the address to be used")
    flag.Parse()
    fmt.Println("local:", *localPtr)
    fmt.Println("address:", *addressPtr)

    baseUrl := baseUrlDefault
    if *localPtr {
        baseUrl = baseUrlLocal
    }
    address := *addressPtr

    //account, err := horizon.DefaultTestNetClient.LoadAccount(address)
    horizonClient := &horizon.Client{
        URL:  baseUrl,
        HTTP: http.DefaultClient,
    }
    account, err := horizonClient.LoadAccount(address)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Balances for account:", address)

    for _, balance := range account.Balances {
        log.Println(balance)
    }
}
