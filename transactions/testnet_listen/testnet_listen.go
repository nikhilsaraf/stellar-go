package main

import (
    "fmt"
    "flag"
    "golang.org/x/net/context"
    "github.com/stellar/go/clients/horizon"
)

func bindPaymentHandler(address string) func(horizon.Payment) {
    return func(p horizon.Payment) {
        if p.To != address {
            return
        }

        var asset string
        if p.AssetType == "native" {
            asset = "lumens"
        } else {
            asset = p.AssetCode + ":" + p.AssetIssuer
        }

        fmt.Printf("\nID=%v" +
            "\nType=%v" +
            "\nFrom=%v" +
            "\nTo=%v" +
            "\nPagingToken=%v" +
            "\nAsset=%v" +
            "\nAmount=%v" +
            "\nMemoType=%v" +
            "\nMemo=%v" +
            "\n",
            p.ID,
            p.Type,
            p.From,
            p.To,
            p.PagingToken,
            asset,
            p.Amount,
            p.Memo.Type,
            p.Memo.Value,
        )
    }
}

func main() {
    addressPtr := flag.String("a", "", "the address for which we want to stream payments")
    sinceTokenPtr := flag.String("s", "", "(optional) token after which we want to stream payments, excludes the token provided")
    flag.Parse()

    if *addressPtr == "" {
        flag.PrintDefaults()
        return
    }
    address := *addressPtr
    fmt.Println("address entered:", address)

    var sinceToken string
    if *sinceTokenPtr != "" {
        sinceToken = *sinceTokenPtr
    }
    fmt.Println("since token:", sinceToken)

    client := horizon.DefaultTestNetClient
    cursor := horizon.Cursor(sinceToken)
    client.StreamPayments(context.Background(), address, &cursor, bindPaymentHandler(address))
}
