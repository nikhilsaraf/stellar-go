package main

import (
    "fmt"
    "log"
    "flag"
    "net/http"
    "github.com/stellar/go/keypair"
    "github.com/stellar/go/clients/horizon"
    b "github.com/stellar/go/build"
    "github.com/kr/pretty"
)

const baseUrlDefault = "https://horizon-testnet.stellar.org"
const baseUrlLocal = "http://localhost:8000"

func parseAsset(code *string, issuer *string) b.Asset {
    if *code == "native" {
        return b.NativeAsset()
    } else {
        return b.CreditAsset(*code, *issuer)
    }
}

func main() {
    localPtr := flag.Bool("l", false, "(optional) whether we should use the local horizon server @ " + baseUrlLocal)
    sourceSeedPtr := flag.String("s", "", "sourceSeed - seed of the source's account")
    sellingAssetCodePtr := flag.String("sc", "", "sellingCode - code for asset being sold (USD, BTC, native, etc.)")
    sellingIssuerCodePtr := flag.String("si", "", "sellingIssuer - if sellingAssetCode is not native, then this needs to be the issuer for the assets being sold")
    buyingAssetCodePtr := flag.String("bc", "", "buyingCode - code for asset being bought (USD, BTC, native, etc.)")
    buyingIssuerCodePtr := flag.String("bi", "", "buyingIssuer - if buyingAssetCode is not native, then this needs to be the issuer for the assets being bought")
    pricePtr := flag.String("p", "", "price - price of 1 unit of selling in terms of buying. For example, if you wanted to sell 30 XLM and buy 5 BTC, the price would be 0.1667")
    amountPtr := flag.Int("amt", -1, "amount - amount of selling being sold. Set to 0 if you want to delete an existing offer")
    passivePtr := flag.Bool("passive", false, "(optional) whether this is a passive offer or not")
    offerIdPtr := flag.Int("offerId", -1, "(not needed if passive) offerId - the ID of the offer. 0 for new offer. Set to existing offer ID to update or delete")
    flag.Parse()

    if *sourceSeedPtr == "" || *sellingAssetCodePtr == "" || *buyingAssetCodePtr == "" || *pricePtr == "" || (*pricePtr)[0] == '-' || *amountPtr < 0 {
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
    if !(*passivePtr) && *offerIdPtr < 0 {
        flag.PrintDefaults()
        return
    }
    if *passivePtr && *offerIdPtr >= 0 {
        flag.PrintDefaults()
        return
    }
    if *amountPtr == 0 && *offerIdPtr == 0 {
        flag.PrintDefaults()
        return
    }

    baseUrl := baseUrlDefault
    if *localPtr {
        baseUrl = baseUrlLocal
    }

    sourceSeed := *sourceSeedPtr
    sourceKP, e := keypair.Parse(sourceSeed)
    if e != nil {
        log.Fatal(e)
    }
    sourceAddress := sourceKP.Address()
    sellingAsset := parseAsset(sellingAssetCodePtr, sellingIssuerCodePtr)
    buyingAsset := parseAsset(buyingAssetCodePtr, buyingIssuerCodePtr)
    price := *pricePtr
    amount := b.Amount(fmt.Sprintf("%v", *amountPtr))
    passive := *passivePtr
    offerId := b.OfferID(uint64(*offerIdPtr))

    fmt.Println("local:", *localPtr)
    fmt.Println("baseUrl:", baseUrl)
    fmt.Println("sourceSeed:", sourceSeed)
    fmt.Println("sourceAddress:", sourceAddress)
    fmt.Println("sellingAsset (code, issuer, isNative):", sellingAsset)
    fmt.Println("buyingAsset (code, issuer, isNative):", buyingAsset)
    fmt.Println("price:", price)
    fmt.Println("amount:", amount)
    fmt.Println("passive:", passive)
    fmt.Println("offerId:", offerId)
    fmt.Println()

    horizonClient := &horizon.Client{
        URL: baseUrl,
        HTTP: http.DefaultClient,
    }

    // validate accounts
    loadAccount(horizonClient, sourceAddress, "source")

    rate := b.Rate{ sellingAsset, buyingAsset, b.Price(price) }
    var ob b.ManageOfferBuilder
    if amount == "0" {
        ob = b.DeleteOffer(rate, offerId)
    } else if passive {
        ob = b.CreatePassiveOffer(rate, amount)
    } else if offerId != 0 {
        ob = b.UpdateOffer(rate, amount, offerId)
    } else {
        ob = b.CreateOffer(rate, amount)
    }

    txn, e := b.Transaction(
        b.SourceAccount{sourceSeed},
        b.AutoSequence{horizonClient},
        b.TestNetwork,
        ob,
    )
    if e != nil {
        log.Fatal(e)
    }
    // sign
    txnS, e := txn.Sign(sourceSeed)
    if e != nil {
        log.Fatal(e)
    }
    // convert to base64
    txnS64, e := txnS.Base64()
    if e != nil {
        log.Fatal(e)
    }
    fmt.Printf("tx base64: %s\n", txnS64)

    // submit the transaction
    resp, e := horizonClient.SubmitTransaction(txnS64)
    if e != nil {
        log.Fatal(e)
    }
    fmt.Println("transaction posted in ledger:", resp.Ledger)
    fmt.Println("response:")
    pretty.Println(resp)

    // print final balances by reloading accounts
    loadAccount(horizonClient, sourceAddress, "source")
}

func loadAccount(horizonClient *horizon.Client, publicKey string, accountName string) horizon.Account {
    account, e := horizonClient.LoadAccount(publicKey)
    if e != nil {
        log.Fatal(e)
    }
    fmt.Println("Balances for account (" + accountName + "):")
    for _, balance := range account.Balances {
        log.Println("   ", balance)
    }
    return account
}
