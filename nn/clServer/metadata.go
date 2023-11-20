package clServer

import (
    "fmt"
    "net"
    "bytes"
    "encoding/binary"
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
    // Create a buffer to read from the connection
    buf := make([]byte, 1024)
    n, err := conn.Read(buf)
    if err != nil {
        fmt.Println("Error reading:", err)
        return
    }

    // The first byte is the clientID
    clientID := buf[0]

    // The rest is the command
    command := string(buf[1:n])

    // Use a switch statement for the command
    switch command {
    case "put":
        blockList, ok := <-ClientChanMap[clientID]
        if ok {
            fmt.Println("Sending blocklist")
            var byteBuffer bytes.Buffer

            // Write the size of blockList as an int64 to the byteBuffer
            binary.Write(&byteBuffer, binary.BigEndian, int64(len(blockList)))

            // Write the byteBuffer to the connection
            conn.Write(byteBuffer.Bytes())
            ackBuf := make([]byte, 3)
            _, err = conn.Read(ackBuf)
            if err != nil {
                fmt.Println("Error reading:", err)
                return
            }
            conn.Write(blockList)
        } else {
            fmt.Println("Error: No blocklist for clientID", clientID)
        }

    case "get":
        blockList, ok := <-ClientChanMap[clientID]
        if ok {
            fmt.Println("Sending blocklist")
            var byteBuffer bytes.Buffer

            // Write the size of blockList as an int64 to the byteBuffer
            binary.Write(&byteBuffer, binary.BigEndian, int64(len(blockList)))

            // Write the byteBuffer to the connection
            conn.Write(byteBuffer.Bytes())
            ackBuf := make([]byte, 3)
            _, err = conn.Read(ackBuf)
            if err != nil {
                fmt.Println("Error reading:", err)
                return
            }
            conn.Write(blockList)
        } else {
            fmt.Println("Error: No blocklist for clientID", clientID)
        }

    default:
        fmt.Println("Unknown command:", command)
    }
}