package main

import (
    "fmt"
    "log"
    "bufio"
    "os"
    "strings"
    "github.com/stellar/go/keypair"
)

func main() {
    // pipe secret key directly after decryption for security
    reader := bufio.NewReader(os.Stdin)
    secret, _ := reader.ReadString('\n')
    secret = strings.Replace(secret, "\n", "", -1)
    fmt.Println("\nreceived secret key, generating public key now.")

    sourceKP, err := keypair.Parse(secret)
    if err != nil {
        log.Fatal(err)
    }
    address := sourceKP.Address()

    fmt.Println("address:", address)
    fmt.Println()
}
