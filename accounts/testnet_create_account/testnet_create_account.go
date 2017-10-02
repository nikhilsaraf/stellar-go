package main

import (
    "bufio"
    "net/http"
    "io/ioutil"
    "log"
    "fmt"
    "os"
    "strings"
)

const baseUrl = "https://horizon-testnet.stellar.org"

func main() {
    fmt.Println("Using baseUrl:", baseUrl)

    reader := bufio.NewReader(os.Stdin)
    fmt.Println("Enter address for new account:")
    address, _ := reader.ReadString('\n')
    address = strings.Replace(address, "\n", "", -1)
    fmt.Println("Address entered:", address)

    resp, err := http.Get(baseUrl + "/friendbot?addr=" + address)
    if err != nil {
        log.Fatal(err)
    }

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(body))
}
