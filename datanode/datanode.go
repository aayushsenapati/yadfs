package main

import (
    "fmt"
    "net"
    "os"
    "time"
)

type Packet struct {
    data []byte
    addr net.Addr
}

func main() {
    conn, err := net.Dial("tcp", "172.18.0.3:12345")
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        os.Exit(1)
    }
    defer conn.Close()

    dataPipe := make(chan Packet)
    ticker := time.NewTicker(3 * time.Second)
    timer := time.NewTimer(6 * time.Second)
    defer ticker.Stop()
    defer timer.Stop()
    message := "heartbeat"
    
    go receTCP(conn,dataPipe)
    for {
        select {
            case <-ticker.C:
                fmt.Println("Sending to server:", message)
                _, err := conn.Write([]byte(message))
                if err != nil {
                    fmt.Println("Error sending message:", err)
                    os.Exit(1)
                }
            case packet := <-dataPipe:
                fmt.Println("Received from server:", string(packet.data))
                timer.Reset(6 * time.Second)
            case <-timer.C:
                fmt.Println("Timeout, no response from server")
                conn.Close()
                conn, err = net.Dial("tcp", "172.18.0.3:12345")
                if err != nil {
                    fmt.Println("Error connecting to server:", err)
                    os.Exit(1)
                }
                defer conn.Close()
                go receTCP(conn,dataPipe)
        }
    }
}



func receTCP(conn net.Conn, dataPipe chan Packet) {
    buf := make([]byte, 1024)
    for {
        n, err := conn.Read(buf)
        if err != nil {
            fmt.Println("Error reading:", err)
            return
        }
        dataPipe <- Packet{data: buf[:n], addr: conn.RemoteAddr()}
    }
}