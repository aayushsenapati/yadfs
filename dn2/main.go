package main

import (
    "fmt"
    "dn/maintenance"
    "sync"
)

func main() {
    ipString:="nn-container-devel"
    var wg sync.WaitGroup

    wg.Add(2) // Add 2 because we have 2 goroutines

    go func() {
        maintenance.SendHb(ipString,"1200") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

    go func() {
        maintenance.SendReport(ipString,"1201") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

    fmt.Println("Datanode listening on")

    wg.Wait() // Wait for all goroutines to finish
}
