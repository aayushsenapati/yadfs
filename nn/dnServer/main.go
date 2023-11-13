package dnServer

import (
    "fmt"
    "net"
    "time"
    "runtime"
    "sync"
    "math/rand"
    "encoding/binary"
    "strings"
    "strconv"
    "os"
    "io"
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

func receiveFile(conn net.Conn,id uint64,size int64) error {

    // Create a new file for writing
    file, err := os.Create("blockreports/" + strconv.FormatUint(id,10))  
    if err != nil {
        return err
    }
    defer file.Close()

    // Copy data from the connection to the file
    _, err = io.CopyN(file, conn, size)
    if err != nil {
        return err
    }

    fmt.Println("File received successfully.")
    return nil
}





func handleNewDataNode(conn net.Conn, connMap map[uint64]DataNode) {
    dataPipe := make(chan Packet)
    receivingFile := make(chan bool,1)
    fmt.Println("before go receTCP")
    go receTCP(conn, dataPipe, receivingFile)
    receivingFile <- false
    fmt.Println("after go receTCP")
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
            timer.Reset(6 * time.Second)
            message := string(packet.data[8:])
            fmt.Println(message, ":", packet.addr)
            id = binary.BigEndian.Uint64(packet.data[:8])
            if strings.HasPrefix(message,"blockreport"){
                //blockreport is followed by filesize string store that in a variable
                token:=strings.Split(message,":")[1]
                brsize,err:=strconv.Atoi(token)
                if err!=nil{
                    fmt.Println("Error in converting string to int")
                }
                receivingFile <- true
                timer.Stop() // stop the timer
                receiveFile(conn,id,int64(brsize))
                timer.Reset(6 * time.Second) // Reset the timer
                receivingFile <- false
            }else{
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

            }
            
        }
    }
}

func receTCP(conn net.Conn, dataPipe chan Packet, receivingFile chan bool) {
    fmt.Println("Waiting for data")
    buf := make([]byte, 1024)
    for {
        if !<-receivingFile {
            fmt.Println("Waiting for data not file")
            n, err := conn.Read(buf)
            if err != nil {
                fmt.Println("Error reading:", err)
                return
            }
            dataPipe <- Packet{data: buf[:n], addr: conn.RemoteAddr()}
        }
        
    }
}
