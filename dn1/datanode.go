package main

import (
    "fmt"
    "net"
    "os"
    "time"
    "encoding/binary"
    "bytes"
    "io/ioutil"
)

type Packet struct {
    data []byte
    addr net.Addr
}

func readIDFromFile(filename string) (uint64, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return 0, err
    }
    id := binary.BigEndian.Uint64(data)
    return id, nil
}

func writeIDToFile(filename string, id uint64) error {
    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, id)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}


func main() {
    ipString:="172.18.0.3:12345"
    conn, err := net.Dial("tcp", ipString)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        os.Exit(1)
    }
    defer conn.Close()

    //variable for unique id for this datanode
    var fileExists bool=true
    id,err:=readIDFromFile("id.bin")
    if err != nil {
        fmt.Println("Error reading id from file:", err,"id does not exist")
        fileExists=false
        id=0
    }

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
                if !fileExists{
                    err=writeIDToFile("id.bin",id)
                    if err != nil {
                        fmt.Println("Error writing id to file:", err)
                        os.Exit(1)
                    }
                    fileExists=true
                }
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