package dnServer

import (
    "fmt"
    "net"
    "time"
    "runtime"
    "sync"
    "math/rand"
    "encoding/binary"
)

type Packet struct {
    data []byte
    addr net.Addr
}

type DataNode struct {
    conn net.Conn
    addr net.Addr
}

var (
    connMap map[uint64]DataNode
    mutex   sync.Mutex
)


 func Listen(ip,port string){
    fmt.Println("test from dnServer")
    ipString:=ip+":"+port
    listener, err := net.Listen("tcp", ipString)
    if err != nil {
        fmt.Println("Error listening:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server listening on", ipString)
    connMap = make(map[uint64]DataNode)
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            return
        }
        fmt.Println("New connection established")
        fmt.Println("Number of goroutines:", runtime.NumGoroutine())
        go handleNewDataNode(conn,connMap)
    }
}


func handleNewDataNode(conn net.Conn, connMap map[uint64]DataNode) {
    dataPipe := make(chan Packet)
    go receTCP(conn, dataPipe)
    timer := time.NewTimer(6 * time.Second)
    defer timer.Stop()
    var id uint64=0
    for {
        select {


        case <-timer.C:
            fmt.Println("timeout:",conn.RemoteAddr())

            mutex.Lock()
            delete(connMap, id)
            fmt.Println("Current connections:", connMap)
            mutex.Unlock()

            return



        case packet := <-dataPipe:
            fmt.Println(string(packet.data[8:]), ":", packet.addr)
            id = binary.BigEndian.Uint64(packet.data[:8])
            if id == 0 {
                newUint := rand.Uint64()
                mutex.Lock()
                connMap[newUint] = DataNode{conn: conn, addr: packet.addr}
                fmt.Println("Current connections:", connMap)
                mutex.Unlock()
                fmt.Println("Generating New ID:", newUint)
                ack := make([]byte, 8)
                binary.BigEndian.PutUint64(ack, newUint)
                ack = append(ack, []byte("ACK")...)
                _, err := conn.Write(ack)
                if err != nil {
                    fmt.Println("Error sending ACK:", err)
                }
            } else {
                //check if connmap(id) exists
                if _, ok := connMap[id]; !ok {
                    fmt.Println("ID not found")
                    mutex.Lock()
                    connMap[id] = DataNode{conn: conn, addr: packet.addr}
                    fmt.Println("Current connections:", connMap)
                    mutex.Unlock()
                }
                ack := make([]byte, 8)
                binary.BigEndian.PutUint64(ack, id)
                ack = append(ack, []byte("ACK")...)
                _, err := conn.Write(ack)
                if err != nil {
                    fmt.Println("Error sending ACK:", err)
                }
            }
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
