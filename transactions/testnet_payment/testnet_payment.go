package main

import (
    "fmt"
    "log"
    "flag"
    "strings"
    "net/http"
    "github.com/stellar/go/keypair"
    "github.com/stellar/go/clients/horizon"
    b "github.com/stellar/go/build"
)

const baseUrlDefault = "https://horizon-testnet.stellar.org"
const baseUrlLocal = "http://localhost:8000"

func main() {
    localPtr := flag.Bool("l", false, "(optional) whether we should use the local horizon server @ " + baseUrlLocal)
    fromSeedPtr := flag.String("fromSeed", "", "seed of the source's account")
    toAddressPtr := flag.String("toAddress", "", "destination address of the receiver's account")
    amountPtr := flag.Float64("amount", 0.0, "amount to be sent, must be > 0.0")
    memoPtr := flag.String("memo", "", "(optional) memo to include with the payment")
    assetPtr := flag.String("asset", "", "(optional) asset to pay with, of the form code:issuer")
    flag.Parse()

    if *fromSeedPtr == "" || *toAddressPtr == "" || *amountPtr <= 0 {
        flag.PrintDefaults()
        return
    }

    baseUrl := baseUrlDefault
    if *localPtr {
        baseUrl = baseUrlLocal
    }

    sourceSeed := *fromSeedPtr
    destinationAddress := *toAddressPtr
    amount := *amountPtr
    memo := *memoPtr
    asset := *assetPtr
    sourceKP, err := keypair.Parse(sourceSeed)
    if err != nil {
        log.Fatal(err)
    }
    sourceAddress := sourceKP.Address()

    fmt.Println("local:", *localPtr)
    fmt.Println("baseUrl:", baseUrl)
    fmt.Println("fromSeed:", sourceSeed)
    fmt.Println("fromAddress:", sourceAddress)
    fmt.Println("toAddress:", destinationAddress)
    fmt.Println("amount:", amount)
    fmt.Println("memo:", memo)
    fmt.Println("asset:", asset)
    fmt.Println()

    horizonClient := &horizon.Client{
        URL: baseUrl,
        HTTP: http.DefaultClient,
    }

    // validate accounts
    sourceAccount := loadAccount(horizonClient, sourceAddress, "source")
    destinationAccount := loadAccount(horizonClient, destinationAddress, "destination")
    
    amountStr := fmt.Sprintf("%v", amount)
    var assetAmount b.PaymentMutator
    if asset != "" {
        assetParts := strings.SplitN(asset, ":", 2)
        issuerAddress := assetParts[1]
        creditAmount := b.CreditAmount{assetParts[0], issuerAddress, amountStr}
        fmt.Println("using non-native asset:", creditAmount)

        // if source account is issuer it does not need to trust the asset
        if !hasAsset(&sourceAccount, &creditAmount) && sourceAddress != issuerAddress {
            log.Fatal(fmt.Sprintf("source account does not trust asset: %v", creditAmount))
        }

        // if destination account is issuer it does not need to trust the asset
        if !hasAsset(&destinationAccount, &creditAmount) && destinationAddress != issuerAddress {
            log.Fatal(fmt.Sprintf("destination account does not trust asset: %v", creditAmount))
        }

        assetAmount = creditAmount
    } else {
        assetAmount = b.NativeAmount{amountStr}
    }

    txn, e := b.Transaction(
        b.SourceAccount{sourceSeed},
        b.AutoSequence{horizonClient},
        b.TestNetwork,
        b.Payment(
            b.Destination{destinationAddress},
            assetAmount,
        ),
    )
    if e != nil {
        log.Fatal(e)
    }
    if memo != "" {
        txn.Mutate(b.MemoText{memo})
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

    // print final balances by reloading accounts
    loadAccount(horizonClient, sourceAddress, "source")
    loadAccount(horizonClient, destinationAddress, "destination")
}

func loadAccount(horizonClient *horizon.Client, publicKey string, accountName string) horizon.Account {
    account, err := horizonClient.LoadAccount(publicKey)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println()
    fmt.Println("Balances for account (" + accountName + "):")
    for _, balance := range account.Balances {
        log.Println("   ", balance)
    }
    return account
}

func hasAsset(account *horizon.Account, creditAmount *b.CreditAmount) bool {
    for _, balance := range (*account).Balances {
        if balance.Asset.Code == (*creditAmount).Code && balance.Asset.Issuer == (*creditAmount).Issuer {
            return true
        }
    }
    return false
}
