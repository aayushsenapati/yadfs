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
    "bufio"
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
var ClientChanMapMutex = &sync.Mutex{}


var blockSize int = 128*1024
var ClientChanMap map[uint8]chan []byte

func ListenCommand(ip, port string) {
    ipString := ip + ":" + port
    listener, err := net.Listen("tcp", ipString)
    if err != nil {
        fmt.Println("Error listening:", err)
        return
    }
    defer listener.Close()
    ClientChanMap = make(map[uint8]chan []byte)

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


    clientChan := make(chan []byte)

    var clientID uint8
    ClientChanMapMutex.Lock()
    for {
        clientID = uint8(rand.Intn(256))
        if _, exists := ClientChanMap[clientID]; !exists {
            break
        }
    }
    ClientChanMap[clientID] = clientChan
    ClientChanMapMutex.Unlock()

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
                                _, err := os.Stat(dest + ".json")
                                if err == nil {
                                    // File exists
                                    returnMessage = "File " + dest + ".json" + " already exists"
                                }else{
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
                                                /*----------------------------check if blockid is there in blocklog if it is there generate new crand------------------*/
                                                filename:=fmt.Sprintf("blocklogs/%d.bin",selectedDNID[j])
                                                blockLog, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                                                if err != nil {
                                                    fmt.Println("Error opening blocklog:", err)
                                                    return
                                                }
                                                blockLogScanner := bufio.NewScanner(blockLog)
                                                for blockLogScanner.Scan() {
                                                    existingBlockID, _ := strconv.ParseUint(blockLogScanner.Text(), 10, 64)
                                                    if binary.BigEndian.Uint64(blockID) == existingBlockID {
                                                        // Block ID already exists, generate a new one
                                                        crand.Read(blockID[1:])
                                                    }
                                                }
                                                fmt.Println("Block ID:", binary.BigEndian.Uint64(blockID))
                                                /*-----------------------------------add blockIds[j] to blocklog------------------------------------------------*/
                                                
                                                
                                                if err != nil {
                                                    fmt.Println("Error opening blocklog:", err)
                                                    return
                                                }
                                                _, err = blockLog.Write(blockID)
                                                if err != nil {
                                                    fmt.Println("Error writing to blocklog:", err)
                                                    return
                                                }
                                                blockLog.Close()
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
                                                    // Parse the IP address to a net.IP
                                                    ipAddr := net.ParseIP(strings.Split(ip.Conn.RemoteAddr().String(),":")[0])
                                                    fmt.Println("IP:", ipAddr)
                                                    // Convert the IP to a 4-byte representation
                                                    ipBytes := ipAddr.To4()
                                                    fmt.Println("IP bytes:", ipBytes)

                                                    // Convert the block ID to a byte slice
                                                    blockIDBuffer := make([]byte, 8)
                                                    binary.BigEndian.PutUint64(blockIDBuffer, key)

                                                    // Append the IP bytes and block ID buffer
                                                    ipAndBlockIDBuffer := append(ipBytes, blockIDBuffer...)

                                                    // Append the combined buffer to the blockList
                                                    blockList = append(blockList, ipAndBlockIDBuffer...)
                                                } else {
                                                    fmt.Println("No IP found for key", keyUint8)
                                                }
                                            }
                                        }
                                        fmt.Println("Blocklist:", blockList)
                                        go func() {
                                            clientChan <- blockList
                                        }()
                    
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
                                            returnMessage = "Metadata stored successfully"
                                        }
                                    }
                                }
                            }
                        }
                    }




                }
                clientIDBytes := []byte{clientID}
                byteBuffer := []byte(returnMessage)
                byteBuffer = append(clientIDBytes, byteBuffer...)
                conn.Write(byteBuffer)
                fmt.Println("Sent message to client:", returnMessage)
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