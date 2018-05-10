package main

import (
    "log"
    "github.com/stellar/go/build"
    "github.com/stellar/go/clients/horizon"
)

const a1 = "GD2UH6Q4P5YEYIZ7JPHR63UVERMBP766J75CQWFDKN6DYEY6UJ3QYABE"
const s1 = "SD7ZPJFI6JT67T3JUDBDWMVWFLA5T7ME6XQTGBPFA5AKQC5CFHNX535R"
const a2 = "GDMWRETZI7QKQRJHMVGV4HJP4G5WPJVNTOCGFNQ3TKSWRTNMPVP665TU"
const s2 = "SAUHOJVPN7JYNBCX45FP5LTEIMHYTBNAJNJL6N2C2KYRGE2SKU6BIWO2"

func main() {
    tx, e := build.Transaction(
        build.SourceAccount{AddressOrSeed: a1},
        build.TestNetwork,
        build.AutoSequence{SequenceProvider: horizon.DefaultTestNetClient},
        build.SetOptions(
            build.SetAuthRequired(),
        ),
        build.Trust(
            "A1",
            a1,
            build.SourceAccount{AddressOrSeed: a2},
        ),
        build.AllowTrust(
            build.AllowTrustAsset{Code: "A1"},
            build.Authorize{Value: true},
            build.Trustor{Address: a2},
        ),
        build.Payment(
            build.Destination{AddressOrSeed: a2},
            build.CreditAmount{Code: "A1", Issuer: a1, Amount: "1.0"},
        ),
        build.SetOptions(
//            build.ClearAuthRequired(),        // this line needs to be commented because auth required is needed to revoke trust in the next operation
            build.SetAuthRevocable(),
        ),
        build.AllowTrust(
            build.AllowTrustAsset{Code: "A1"},
            build.Authorize{Value: false},
            build.Trustor{Address: a2},
        ),
    )
    if e != nil {
        log.Fatal(e)
    }

    txS, e := tx.Sign(s1, s2)
    if e != nil {
        log.Fatal(e)
    }

    b64, e := txS.Base64()
    if e != nil {
        log.Fatal(e)
    }

    resp, e := horizon.DefaultTestNetClient.SubmitTransaction(b64)
    if e != nil {
        switch t := e.(type) {
        default:
            log.Fatal("error while submitting to network: ", t)
        case *horizon.Error:
            log.Println("horizon.error while submitting to network: ", t.Problem.ToProblem())
            //log.Printf("Type: %v,Title: %s,Status: %v,Detail: %s,Instance: %s\n", t.Problem horizon.error while submitting to network: ", t.Problem.ToProblem())
            r, _ := t.ResultCodes()
            log.Fatal("result codes: ", r)
        }
    }
    
    log.Println(resp)
}
