package clServer

import (
	"fmt"
	"net"
	"time"
)




func Listen(ip,port String){
    ipString:=ip+":"+port
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
        go handleNewClient(conn)
    }
}