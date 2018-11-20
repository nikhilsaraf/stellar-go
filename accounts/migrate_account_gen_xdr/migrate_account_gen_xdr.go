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
	"github.com/stellar/go/xdr"
)

const inflationAddress = "GCCD6AJOYZCUAQLX32ZJF2MKFFAUJ53PVCFQI3RHWKL3V47QYE2BNAUT"
const startingBalanceXlm = "100.0"

type inputs struct {
	fromAccount string
	destAccount string
	seqOffset   int64
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
		b.AutoSequence{
			SequenceProvider: OffsetSequenceProvider{
				inner:  horizonClient,
				offset: ip.seqOffset,
			},
		},
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
	seqOffsetPtr := flag.Int64("seq_offset", -1, "sequence number offset (0 for the next valid seq number, only +ve numbers)")
	flag.Parse()

	if *fromAccountPtr == "" || *destAccountPtr == "" || *seqOffsetPtr < 0 {
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
		seqOffset:   *seqOffsetPtr,
		network:     network,
	}
}

// OffsetSequenceProvider loads the sequence to use for the transaction from an external provider and increments the value by the offset
type OffsetSequenceProvider struct {
	inner  b.SequenceProvider
	offset int64
}

var _ b.SequenceProvider = OffsetSequenceProvider{}

// SequenceForAccount adds the offset to the result of the inner call
func (s OffsetSequenceProvider) SequenceForAccount(aid string) (xdr.SequenceNumber, error) {
	seq, e := s.inner.SequenceForAccount(aid)
	if e != nil {
		return seq, e
	}

	offsetSeq := xdr.SequenceNumber(int64(seq) + s.offset)
	// generated XDR will have a seq number of 1 more than this since this is fetching the current seq number only
	fmt.Printf("added offset of %d to convert current fetched sequence number from %d to %d\n", s.offset, int64(seq), int64(offsetSeq))
	return offsetSeq, nil
}
