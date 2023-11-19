package main

import (
    "fmt"
    "dn/maintenance"
    "sync"
    "dn/filemanager"
)

func main() {
    ipString:="nn-container-devel"
    var wg sync.WaitGroup

    wg.Add(4) // Add 2 because we have 2 goroutines

    go func() {
        maintenance.SendHb(ipString,"1200") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

    go func() {
        maintenance.SendReport(ipString,"1201") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

    go func() {
        filemanager.ClientListener(ipString,"3200") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

    go func() {
        filemanager.Replication(ipString,"4200") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

    wg.Wait() // Wait for all goroutines to finish
}
