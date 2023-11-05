package main

import (
    "fmt"
    "net"
    "os"
    "time"
)

func main() {
    conn, err := net.Dial("tcp", "172.18.0.3:12345")
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        os.Exit(1)
    }
    defer conn.Close()

    message := "heartbeat"
    for {
        fmt.Println("Sending to server:", message)
        _, err := conn.Write([]byte(message))
        if err != nil {
            fmt.Println("Error sending message:", err)
            os.Exit(1)
        }
        time.Sleep(3 * time.Second)
    }
}

