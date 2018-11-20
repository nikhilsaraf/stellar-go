package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	b "github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

const inflationAddress = "GCCD6AJOYZCUAQLX32ZJF2MKFFAUJ53PVCFQI3RHWKL3V47QYE2BNAUT"
const startingBalanceXlm = "100.0"

type inputs struct {
	fromAccount string
	destAccount string
	network     b.Network
}

func main() {
	fmt.Printf("====================================================================================================\n")
	ip := parseInputs()

	horizonClient := horizon.DefaultTestNetClient
	if ip.network == b.PublicNetwork {
		horizonClient = horizon.DefaultPublicNetClient
	}
	txn, e := b.Transaction(
		b.SourceAccount{AddressOrSeed: ip.fromAccount},
		b.AutoSequence{SequenceProvider: horizonClient},
		ip.network,
		b.CreateAccount(
			b.Destination{AddressOrSeed: ip.destAccount},
			b.NativeAmount{Amount: startingBalanceXlm},
		),
		b.AccountMerge(
			b.Destination{AddressOrSeed: ip.destAccount},
		),
		b.SetOptions(
			b.SourceAccount{AddressOrSeed: ip.destAccount},
			b.InflationDest(inflationAddress),
		),
	)
	if e != nil {
		log.Fatal(e)
	}

	// sign with empty signature so it gets converted to a transaction envelope
	txEnv, e := txn.Sign()
	if e != nil {
		log.Fatalf("failed to sign: %s", e)
	}

	// convert to base64
	txEnvBase64, e := txEnv.Base64()
	if e != nil {
		log.Fatalf("failed to convert to base64: %s", e)
	}

	fmt.Printf("\nxdr:\n")
	fmt.Printf("%s\n", txEnvBase64)
	fmt.Printf("====================================================================================================\n")
}

// returns the input XDR that needs to be signed
func parseInputs() inputs {
	fromAccountPtr := flag.String("from", "", "stellar account that needs to be migrated")
	destAccountPtr := flag.String("dest", "", "destination stellar account where we want to migrate to")
	flag.Parse()

	if *fromAccountPtr == "" || *destAccountPtr == "" {
		fmt.Println("Params:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Which network (t/p)? [t]: ")
	networkChoice, _ := reader.ReadString('\n')
	networkChoice = strings.Replace(networkChoice, "\n", "", -1)
	network := b.TestNetwork
	if networkChoice == "p" {
		network = b.PublicNetwork
	}

	return inputs{
		fromAccount: *fromAccountPtr,
		destAccount: *destAccountPtr,
		network:     network,
	}
}
