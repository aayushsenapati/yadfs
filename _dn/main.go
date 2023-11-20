package main

import (
    "dn/maintenance"
    "sync"
    "dn/filemanager"
    "os"
)

func main() {
    ipString:=os.Getenv("NN_CONTAINER_NAME")
    var wg sync.WaitGroup

    wg.Add(3) // Add 2 because we have 2 goroutines

    go func() {
        maintenance.SendHb(ipString,"1200") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

    go func() {
        maintenance.SendReport(ipString,"1201") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

    go func() {
        filemanager.ClientListener(os.Getenv("CONTAINER_NAME"),"3200") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }()

  /*   go func() {
        filemanager.Replication(ipString,"4200") // Launch a goroutine
        wg.Done() // Call Done when the function returns
    }() */

    wg.Wait() // Wait for all goroutines to finish
}
