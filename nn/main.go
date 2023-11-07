package main

import (
    "nn/dnServer"
    "nn/clServer"
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup

    wg.Add(2) // Add 2 because we have 2 goroutines

    go func() {
        dnServer.Listen("172.19.0.4","12345")
        wg.Done() // Call Done when the function returns
    }()

    go func() {
        clServer.Listen("172.19.0.4","12344")
        wg.Done() // Call Done when the function returns
    }()

    fmt.Println("Server listening on")

    wg.Wait() // Wait for all goroutines to finish
}
