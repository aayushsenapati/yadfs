package clServer

import (
    "fmt"
    "net"
)

func Metadata(ip,port string) {
	ipString := ip + ":" + port
    listener, err := net.Listen("tcp", ipString)
    if err != nil {
        fmt.Println("Error listening:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server listening on", ipString)
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }
        go handleNewClient1(conn)
    }
}


func handleNewClient1(conn net.Conn) {
	blockList := <-BlockListChan
	fmt.Println("Sending blocklist")
	conn.Write(blockList)

}