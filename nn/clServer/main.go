package clServer

import (
    "fmt"
    "net"
    "strings"
    "time"
    "os"
    "sync"
)

type Packet struct {
    data []byte
    addr net.Addr
}

var mutex = &sync.Mutex{}

func Listen(ip, port string) {
    fmt.Println("test from clServer")
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
        go handleNewClient(conn)
    }
}

//upon client connection,send back ACK
func handleNewClient(conn net.Conn) {
    fmt.Println("New connection established")
    conn.Write([]byte("ACK\n"))
    //look for timeouts in one routine and messages in another
    timer := time.NewTimer(10 * time.Second)
    defer timer.Stop()

    dataPipe := make(chan Packet)
    go receTCP(conn, dataPipe)

    for {
        select {
        case <-timer.C:
            conn.Close()
            return
        case packet := <-dataPipe:
            timer.Reset(10 * time.Second)
            fmt.Println("Received client packet from", packet.addr)
            command := string(packet.data)
            fmt.Println("Command:", command)
            var returnMessage string
            cmdArgs := strings.Split(command, " ")
            if len(cmdArgs) <= 1 {
                returnMessage = "undefined behaviour"
            } else {
                switch cmdArgs[0] {
                case "mkdir":
                    path := "root/"+cmdArgs[1]
                    mutex.Lock()
                    fileInfo, err := os.Stat(path) //Checking if path exists
                    if err != nil {
                        if os.IsNotExist(err) {
                            err := os.MkdirAll(path, os.ModePerm)
                            if err != nil {
                                fmt.Println(err)
                            }
                            returnMessage = "Directory created at: " + path
                        }
                    } else if fileInfo.IsDir() {
                        returnMessage = "Directory already exists"
                    } else {
                        returnMessage = "File exists at path"
                    }
                    mutex.Unlock()
                }
                byteBuffer := []byte(returnMessage)
                conn.Write(byteBuffer)
            }

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
