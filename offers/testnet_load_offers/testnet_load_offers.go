package main

import (
    "fmt"
    "log"
    "flag"
    "net/http"
    "github.com/stellar/go/clients/horizon"
    "github.com/kr/pretty"
)

const baseUrlDefault = "https://horizon-testnet.stellar.org"
const baseUrlLocal = "http://localhost:8000"

func main() {
    localPtr := flag.Bool("l", false, "(optional) whether we should use the local horizon server @ " + baseUrlLocal)
    addressPtr := flag.String("a", "", "address - address of the offers to load")
    flag.Parse()

    if *addressPtr == "" {
        flag.PrintDefaults()
        return
    }
    address := *addressPtr

    baseUrl := baseUrlDefault
    if *localPtr {
        baseUrl = baseUrlLocal
    }

    fmt.Println("local:", *localPtr)
    fmt.Println("baseUrl:", baseUrl)
    fmt.Println("address:", address)
    fmt.Println()

    horizonClient := &horizon.Client{
        URL: baseUrl,
        HTTP: http.DefaultClient,
    }

    offers, err := horizonClient.LoadAccountOffers(address)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Offers:")
    for _, o := range offers.Embedded.Records {
        pretty.Println(o)
        fmt.Println("\n")
    }
}
