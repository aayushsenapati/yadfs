package dnServer

import (
    "fmt"
    "net"
    "strconv"
    "encoding/binary"
    "sort"
    "bytes"
    "io/ioutil"
    "io"
)

func receiveFile(conn net.Conn) error {
    // Read the header containing the file name and size
    headerBuf := make([]byte, 9)
    _, err := conn.Read(headerBuf)
    if err != nil {
        return err
    }

    id := uint8(headerBuf[0])
    //header := string(headerBuf[1:n])
    //fileSizeStr:=strings.Split(header,":")[1]
    //fileSize, err := strconv.Atoi(fileSizeStr)
    //if err != nil {
    //    return err
    //}
    var buf bytes.Buffer
    fileSize:=binary.BigEndian.Uint64(headerBuf[1:])
    var reportBuf bytes.Buffer
    _, err = io.CopyN(&buf, conn, int64(fileSize))
    if err != nil {
        return err
    }
    //read the block log from blocklog/id.bin and stor in buffer
    var logBuf bytes.Buffer
    log,err:=ioutil.ReadFile("blocklogs/"+strconv.Itoa(int(id))+".bin")
    if err != nil {
            fmt.Println("Error reading file:", err)
            return err
    }
    logBuf.Write(log)

    reportSlice := byteSliceToUint64SliceBE(reportBuf.Bytes())
    logSlice := byteSliceToUint64SliceBE(logBuf.Bytes())

    sort.Slice(logSlice, func(i, j int) bool { return logSlice[i] < logSlice[j] })
    sort.Slice(reportSlice, func(i, j int) bool { return reportSlice[i] < reportSlice[j] })

    diffBuf := uint64SliceToByteSliceBE(findDifferences(reportSlice, logSlice))

    // Send the size of the differences slice
    sizeBuf := make([]byte, 8)
    binary.BigEndian.PutUint64(sizeBuf, uint64(len(diffBuf)))
    _, err = conn.Write(sizeBuf)
    if err != nil {
        fmt.Println("Error sending size:", err)
        return err
    }
    conn.Read(make([]byte, 3))
    // Send the differences slice
    fmt.Println("Sending differences:", diffBuf)
    _, err = conn.Write(diffBuf)
    if err != nil {
        fmt.Println("Error sending differences:", err)
        return err
    }


    fmt.Println("File received successfully.")
    return nil
}



func uint64SliceToByteSliceBE(slice []uint64) []byte {
    buffer := make([]byte, len(slice)*8)
    for i, v := range slice {
        binary.BigEndian.PutUint64(buffer[i*8:(i+1)*8], v)
    }
    return buffer
}


func byteSliceToUint64SliceBE(buffer []byte) []uint64 {
    slice := make([]uint64, len(buffer)/8)
    for i := 0; i < len(slice); i++ {
        slice[i] = binary.BigEndian.Uint64(buffer[i*8 : (i+1)*8])
    }
    return slice
}

func findDifferences(slice1, slice2 []uint64) []uint64 {
    var differences []uint64

    i, j := 0, 0
    for i < len(slice1) && j < len(slice2) {
        if slice1[i] < slice2[j] {
            differences = append(differences, slice1[i])
            i++
        } else if slice1[i] > slice2[j] {
            j++
        } else {
            i++
            j++
        }
    }

    // Append remaining elements from slice1
    for i < len(slice1) {
        differences = append(differences, slice1[i])
        i++
    }

    return differences
}




func ListenReport(ip,port string){
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
    
        go func(c net.Conn) {
            defer c.Close()
            for {
                err := receiveFile(c)
                if err != nil {
                    fmt.Println("Error receiving file:", err)
                    return
                }
            }
        }(conn)
    }
}