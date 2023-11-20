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
    "bytes"
    "io/ioutil"
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
                                                // Check if block ID is there in blocklog, if it is there generate new crand
                                                filename := fmt.Sprintf("blocklogs/%d.bin", selectedDNID[j])
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
                case "get":
                    if len(cmdArgs) < 2 {
                        returnMessage = "Usage: get destination source"
                        break
                    }

                    src := "root/" + cmdArgs[1]
                    _, err := os.Stat(src + ".json")
                    if os.IsNotExist(err) {
                        returnMessage = "Path does not exist"
                        break
                    }

                    file, err := os.Open(src + ".json")
                    if err != nil {
                        returnMessage = "Error opening file: " + err.Error()
                        break
                    }
                    // Create a bytes.Buffer to store the block IDs
                    // Create a bytes.Buffer to store the block IDs and IP addresses
                    var blockIDBuffer bytes.Buffer
                    decoder := json.NewDecoder(file)
                    var fileData FileData
                    err = decoder.Decode(&fileData)
                    if err != nil {
                        returnMessage = "Error decoding file: " + err.Error()
                        break
                    }
                    // Iterate over the Blocks field of the FileData object
                    for _, block := range fileData.Blocks {
                        for _, blockID := range block {
                            // Check if the data node ID (the first byte of the block ID) exists in the dnServer.ConnMap
                            dataNodeID := uint8(blockID >> 56)
                            if dn, exists := dnServer.ConnMap[dataNodeID]; exists {
                                // Get the IP address of the data node
                                ip := dn.Conn.RemoteAddr().(*net.TCPAddr).IP

                                // Write the IP address to the buffer
                                blockIDBuffer.Write(ip.To4())

                                // Write the block ID to the buffer
                                binary.Write(&blockIDBuffer, binary.BigEndian, blockID)
                                break
                            }
                        }
                    }

                    // Now you can use blockIDBuffer.Bytes() to get the block IDs and IP addresses as a byte slice
                    blockListBuf := blockIDBuffer.Bytes()
                    go func() {
                        clientChan <- blockListBuf
                    }()

                case "rm":
                    if len(cmdArgs) < 2 {
                        returnMessage = "Usage: rm destination"
                        break
                    }
                    filepath := "root/" + cmdArgs[1] + ".json"
                
                    // Check if the file exists
                    _, err := os.Stat(filepath)
                    if os.IsNotExist(err) {
                        returnMessage = "File does not exist"
                        break
                    }
                
                    // Open the file
                    file, err := os.Open(filepath)
                    if err != nil {
                        returnMessage = "Error opening file: " + err.Error()
                        break
                    }
                    defer file.Close()
                
                    // Decode the file
                    decoder := json.NewDecoder(file)
                    var fileData FileData
                    err = decoder.Decode(&fileData)
                    if err != nil {
                        returnMessage = "Error decoding file: " + err.Error()
                        break
                    }
                    fmt.Println("\n\n\n\n\nFile data:", fileData.Blocks)
                    // Iterate over the Blocks field of the FileData object
                    for _, block := range fileData.Blocks {
                        for _, blockID := range block {
                            // Get the data node ID (the first byte of the block ID)
                            dataNodeID := uint8(blockID >> 56)
                
                            // Open the data node file
                            dnFile, err := os.OpenFile(fmt.Sprintf("blocklogs/%d.bin", dataNodeID), os.O_RDWR, 0644)
                            if err != nil {
                                returnMessage = "Error opening data node file: " + err.Error()
                                break
                            }
                
                            // Read the data node file into a byte slice
                            dnData, err := ioutil.ReadAll(dnFile)
                            if err != nil {
                                returnMessage = "Error reading data node file: " + err.Error()
                                break
                            }
                
                            // Convert the byte slice to a uint64 slice
                            dnBlockIDs := make([]uint64, len(dnData)/8)
                            for i := range dnBlockIDs {
                                dnBlockIDs[i] = binary.BigEndian.Uint64(dnData[i*8 : (i+1)*8])
                            }
                
                            // Remove the block ID from the slice
                            for i, dnBlockID := range dnBlockIDs {
                                if dnBlockID == blockID {
                                    dnBlockIDs = append(dnBlockIDs[:i], dnBlockIDs[i+1:]...)
                                    break
                                }
                            }
                            //fmt.Println("\n\n\n\n\nbefore delete", dnBlockIDs)
                            // Convert the uint64 slice back to a byte slice
                            dnData = make([]byte, len(dnBlockIDs)*8)
                            for i, dnBlockID := range dnBlockIDs {
                                binary.BigEndian.PutUint64(dnData[i*8:(i+1)*8], dnBlockID)
                            }
                
                            // Write the byte slice back to the data node file
                            _, err = dnFile.WriteAt(dnData, 0)
                            if err != nil {
                                returnMessage = "Error writing to data node file: " + err.Error()
                                break
                            }
                            dnFile.Close()
                        }
                    }
                
                    // Delete the file
                    err = os.Remove(filepath)
                    if err != nil {
                        returnMessage = "Error deleting file: " + err.Error()
                        break
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