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

func parseAsset(code *string, issuer *string) horizon.Asset {
    if *code == "native" {
        return horizon.Asset{ "native", "", "" }
    } else if len(*code) <= 4 {
        return horizon.Asset{ "credit_alphanum4", *code, *issuer }
    } else {
        return horizon.Asset{ "credit_alphanum12", *code, *issuer }
    }
}

func main() {
    localPtr := flag.Bool("l", false, "(optional) whether we should use the local horizon server @ " + baseUrlLocal)
    sellingAssetCodePtr := flag.String("sc", "", "sellingCode - code for asset being sold (USD, BTC, native, etc.)")
    sellingIssuerCodePtr := flag.String("si", "", "sellingIssuer - if sellingAssetCode is not native, then this needs to be the issuer for the assets being sold")
    buyingAssetCodePtr := flag.String("bc", "", "buyingCode - code for asset being bought (USD, BTC, native, etc.)")
    buyingIssuerCodePtr := flag.String("bi", "", "buyingIssuer - if buyingAssetCode is not native, then this needs to be the issuer for the assets being bought")
    flag.Parse()

    if *sellingAssetCodePtr == "" || *buyingAssetCodePtr == "" {
        flag.PrintDefaults()
        return
    }
    if *sellingAssetCodePtr != "native" && *sellingIssuerCodePtr == "" {
        flag.PrintDefaults()
        return
    }
    if *buyingAssetCodePtr != "native" && *buyingIssuerCodePtr == "" {
        flag.PrintDefaults()
        return
    }

    baseUrl := baseUrlDefault
    if *localPtr {
        baseUrl = baseUrlLocal
    }

    sellingAsset := parseAsset(sellingAssetCodePtr, sellingIssuerCodePtr)
    buyingAsset := parseAsset(buyingAssetCodePtr, buyingIssuerCodePtr)

    fmt.Println("local:", *localPtr)
    fmt.Println("baseUrl:", baseUrl)
    fmt.Println("sellingAsset (code, issuer, isNative):", sellingAsset)
    fmt.Println("buyingAsset (code, issuer, isNative):", buyingAsset)
    fmt.Println()

    horizonClient := &horizon.Client{
        URL: baseUrl,
        HTTP: http.DefaultClient,
    }

    orderBook, err := horizonClient.LoadOrderBook(sellingAsset, buyingAsset)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("OrderBookSummary:")
    pretty.Println(orderBook)
}
