package main

import (
    "fmt"
    "net"
    "os"
    "time"
    "encoding/binary"
)

type Packet struct {
    data []byte
    addr net.Addr
}

func main() {
    ipString:="172.18.0.2:12345"
    conn, err := net.Dial("tcp", ipString)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        os.Exit(1)
    }
    defer conn.Close()

    //variable for unique id for this datanode
    var id uint64 = 0

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
                byteBuffer := make([]byte, len(message)+8)
                binary.BigEndian.PutUint64(byteBuffer, id)
                copy(byteBuffer[8:], []byte(message))
                _, err := conn.Write(byteBuffer)
                if err != nil {
                    fmt.Println("Error sending message:", err)
                    os.Exit(1)
                }
            case packet := <-dataPipe:
                id= binary.BigEndian.Uint64(packet.data[:8])
                fmt.Println("Received from server:", string(packet.data[8:]),id)
                timer.Reset(6 * time.Second)
            case <-timer.C:
                fmt.Println("Timeout, no response from server")
                conn.Close()
                conn, err = net.Dial("tcp", ipString)
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