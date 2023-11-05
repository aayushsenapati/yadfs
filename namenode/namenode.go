package main

import (
    "fmt"
    "net"
    "time"
    "runtime"
    "sync"
)

type Packet struct {
    data []byte
    addr net.Addr
}

var (
    connMap map[string]net.Conn
    mutex   sync.Mutex
)

func main() {
    listener, err := net.Listen("tcp", "172.18.0.3:12345")
    if err != nil {
        fmt.Println("Error listening:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server listening on 172.18.0.3:12345")
    connMap = make(map[string]net.Conn)


    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            return
        }
        fmt.Println("New connection established")
        fmt.Println("Number of goroutines:", runtime.NumGoroutine())
        mutex.Lock()

        connMap[conn.RemoteAddr().String()] = conn
        fmt.Println("Current connections:", connMap)
        mutex.Unlock()

        go handleNewClient(conn)
    }
}

func handleNewClient(conn net.Conn) {
    dataPipe := make(chan Packet)
    go receTCP(conn, dataPipe)
    timer := time.NewTimer(6 * time.Second)
    defer timer.Stop()
    for {
        select {
        case <-timer.C:
            fmt.Println("timeout:",conn.RemoteAddr())

            mutex.Lock()
            delete(connMap, conn.RemoteAddr().String())
            fmt.Println("Current connections:", connMap)
            mutex.Unlock()

            return
        case packet := <-dataPipe:
            fmt.Println(string(packet.data), ":", packet.addr)
            timer.Reset(6 * time.Second)
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