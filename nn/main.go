package main

import (
    "nn/dnServer"
    "nn/clServer"
    "fmt"
    "sync"
)

func main() {
    ipString:="nn-container-devel"
    var wg sync.WaitGroup

    wg.Add(2) // Add 2 because we have 2 goroutines

    go func() {
        dnServer.ListenHb(ipString,"1200")
        wg.Done() // Call Done when the function returns
    }()


    go func() {
        dnServer.ListenReport(ipString,"1201")
        wg.Done() // Call Done when the function returns
    }()



    go func() {
        clServer.Listen(ipString,"12344")
        wg.Done() // Call Done when the function returns
    }()

    fmt.Println("Server listening on")

    wg.Wait() // Wait for all goroutines to finish
}
