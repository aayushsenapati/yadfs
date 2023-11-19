package clServer

import (
    "fmt"
    crand "crypto/rand"
    "math/rand"
    "net"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
    "strconv"
    "nn/dnServer"
    "encoding/json"
    "encoding/binary"
)

type Packet struct {
    data []byte
    addr net.Addr
}

type FileData struct {
    Filename string     `json:"filename"`
    Filesize int        `json:"filesize"`
    Blocks   [][]uint64 `json:"blocks"`
}


var mutex = &sync.Mutex{}


var blockSize int = 128*1024
var BlockListChan chan []byte

func Listen(ip, port string) {
    ipString := ip + ":" + port
    listener, err := net.Listen("tcp", ipString)
    if err != nil {
        fmt.Println("Error listening:", err)
        return
    }
    defer listener.Close()
    BlockListChan = make(chan []byte)

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



                case "put":
                    fmt.Println("in clserver:", dnServer.ConnMap)
                    if len(cmdArgs) < 3 {
                        returnMessage = "Usage: put destination source"
                    } else {
                        dest := "root/" + cmdArgs[1]
                
                        // Get the parent directory of the destination path
                        parentDir := filepath.Dir(dest)
                
                        // Check if the parent directory exists
                        _, err := os.Stat(parentDir)
                        if os.IsNotExist(err) {
                            returnMessage = "Path does not exist"
                        } else {
                            // Create the JSON file
                            if len(dnServer.ConnMap) < 3 {
                                returnMessage = "Not enough IDs in ConnMap"
                            } else {
                                file, err := os.Create(dest + ".json")
                                if err != nil {
                                    returnMessage = "Error creating file: " + err.Error()
                                } else {
                                    defer file.Close()
                                    fileSize, _ := strconv.Atoi(cmdArgs[2])
                                    numBlocks := (fileSize + blockSize - 1) / blockSize
                                    keysAndBlockIDs := make([][]uint64, numBlocks)
                
                                    for i := 0; i < numBlocks; i++ {
                                        // Generate an array of keys from the ConnMap
                                        keys := make([]int, 0, len(dnServer.ConnMap))
                                        for k := range dnServer.ConnMap {
                                            keys = append(keys, int(k))
                                        }
                
                                        // Shuffle the keys array
                                        rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
                
                                        // Select the first 3 keys from the shuffled array
                                        selectedDNID := keys[:3]
                
                                        // Generate three random 56-bit block IDs for this block and its 2 replicas
                                        blockIDs := make([]uint64, 3)
                                        for j := 0; j < 3; j++ {
                                            //generate random 56-bit block ID to append key(8 bit)
                                            blockID := make([]byte, 8)
                                            blockID[0] = byte(selectedDNID[j])
                                            crand.Read(blockID[1:])
                                            blockIDs[j] = binary.BigEndian.Uint64(blockID)
                                        }
                
                                        // Append these block IDs to the keys in the keyArray
                                        keysAndBlockIDs[i] = append(keysAndBlockIDs[i], blockIDs...)
                                    }
                                    blockList:=make([]byte,0)
                                    for _, blocks := range keysAndBlockIDs {
                                        for _, key := range blocks {
                                            // Convert the key back to uint8
                                            keyUint8 := uint8(key >> 56)
                                    
                                            // Get the IP from ConnMap using the key
                                            ip, exists := dnServer.ConnMap[keyUint8]
                                            if exists {
                                                //fmt.Println("IP found for key", keyUint8, ":", ip.Conn.RemoteAddr().String())

                                                byteBuffer := make([]byte,8)
                                                binary.BigEndian.PutUint64(byteBuffer, key)
                                                blockList = append(blockList, byteBuffer...)
                                                ip := fmt.Sprintf("%s\n", ip.Conn.RemoteAddr().String())
                                                blockList = append(blockList, []byte(ip)...)
                                            } else {
                                                fmt.Println("No IP found for key", keyUint8)
                                            }
                                        }
                                    }
                                    fmt.Println("Blocklist:", blockList)
                                    BlockListChan <- blockList
                
                                    fileData := FileData{
                                        Filename: filepath.Base(dest),
                                        Filesize: fileSize,
                                        Blocks:   keysAndBlockIDs,
                                    }
                                    encoder := json.NewEncoder(file)
                
                                    // Write the file data to the file as JSON
                                    err = encoder.Encode(fileData)
                                    if err != nil {
                                        returnMessage = "Error writing to file: " + err.Error()
                                    } else {
                                        returnMessage = "File stored successfully"
                                    }
                                }
                            }
                        }
                    }




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