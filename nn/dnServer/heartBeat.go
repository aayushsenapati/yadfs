package dnServer

import (
    "fmt"
    "net"
    "time"
    "runtime"
    "sync"
    "math/rand"
    "os"
    "strconv"
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
    ConnMap map[uint8]DataNode
    mutex   sync.Mutex
)


 func ListenHb(ip,port string){
    ipString:=ip+":"+port
    listener, err := net.Listen("tcp", ipString)
    if err != nil {
        fmt.Println("Error listening:", err)
        return
    }
    defer listener.Close()

    fmt.Println("Server listening on", ipString)
    ConnMap = make(map[uint8]DataNode)
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            return
        }
        fmt.Println("New connection established")
        fmt.Println("Number of goroutines:", runtime.NumGoroutine())
        go handleNewDataNode(conn,ConnMap)
    }
}


func handleNewDataNode(conn net.Conn, ConnMap map[uint8]DataNode) {
    dataPipe := make(chan Packet)
    go receTCP(conn, dataPipe)
    timer := time.NewTimer(6 * time.Second)
    defer timer.Stop()
    var id uint8=0
    for {
        select {


        case <-timer.C:
            fmt.Println("timeout:",conn.RemoteAddr())

            mutex.Lock()
            delete(ConnMap, id)
            fmt.Println("Current connections:", ConnMap)
            mutex.Unlock()

            return



        case packet := <-dataPipe:
            fmt.Println(string(packet.data[1:]), ":", packet.addr)
            id = uint8(packet.data[0])
            if id == 0 {
                // Check if the "blockreports" directory exists. If not, create it.
                _, err := os.Stat("blockreports")
                if os.IsNotExist(err) {
                    err = os.MkdirAll("blockreports", 0755)
                    if err != nil {
                        fmt.Println("Error creating directory:", err)
                        return
                    }
                }
            
                // Generate a new random ID.
                newUint := uint8(rand.Intn(256))
            
                // Check if a file with the name of the new ID exists in the "blockreports" directory.
                _, err = os.Stat("blockreports/" + strconv.Itoa(int(newUint)))
                for os.IsExist(err) {
                    // If it does, generate a new ID and repeat the process until an unused ID is found.
                    newUint := uint8(rand.Intn(256))
                    _, err = os.Stat("blockreports/" + strconv.Itoa(int(newUint)))
                }
            
                mutex.Lock()
                ConnMap[newUint] = DataNode{conn: conn, addr: packet.addr}
                fmt.Println("Current connections:", ConnMap)
                mutex.Unlock()
                fmt.Println("Generating New ID:", newUint)
                ack := make([]byte, 1)
                ack[0]=byte(newUint)
                ack = append(ack, []byte("ACK")...)
                _, err = conn.Write(ack)
                if err != nil {
                    fmt.Println("Error sending ACK:", err)
                }
            } else {
                //check if connmap(id) exists
                if _, ok := ConnMap[id]; !ok {
                    fmt.Println("ID not found")
                    mutex.Lock()
                    ConnMap[id] = DataNode{conn: conn, addr: packet.addr}
                    fmt.Println("Current connections:", ConnMap)
                    mutex.Unlock()
                }
                ack := make([]byte, 1)
                ack[0]= byte(id)
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
