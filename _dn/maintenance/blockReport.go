package maintenance

import (
    "fmt"
    "net"
    "os"
    "time"
	"io"
    "encoding/binary"
    "io/ioutil"
    "sort"
)



func SendReport(ipString, portString string) {
    conn, err := net.Dial("tcp", ipString+":"+portString)
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        os.Exit(1)
    }
    defer conn.Close()

	var id uint8
    ticker := time.NewTicker(15 * time.Second)
	var fileExists bool = false
    for {
        <-ticker.C
		if(!fileExists){
			var err error
			id, err = readIDFromFile("id.bin")
			if err != nil {
				continue
			}
			fileExists = true
		}
        err := sendFile(conn, "blocklist.bin",id)
        if err != nil {
            fmt.Println("Error sending file:", err)
        }
    }
}


func sendFile(conn net.Conn, filePath string,id uint8) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    // Get file information
    fileInfo, err := file.Stat()
    if err != nil {
        return err
    }

    // Send the file name and size as a header
    header:=make([]byte,9)
    header[0]=byte(id)
    binary.BigEndian.PutUint64(header[1:],uint64(fileInfo.Size()))
    _, err = conn.Write(header)
    if err != nil {
        return err
    }

    // Wait for the ack
    _, err = conn.Read(make([]byte, 3))
    if err != nil {
        fmt.Println("Error receiving ack:", err)
        return err
    }

    // Send the file content
    _, err = io.Copy(conn, file)
    if err != nil {
        return err
    }

    // Receive the file size
    sizeBuf := make([]byte, 8)
    _, err = conn.Read(sizeBuf)
    if err != nil {
        fmt.Println("Error receiving size:", err)
        return err
    }
    conn.Write([]byte("ack"))

    fileSize := binary.BigEndian.Uint64(sizeBuf)

    // Retrieve the buffer
    buffer := make([]byte, fileSize)
    _, err = conn.Read(buffer)
    if err != nil {
        fmt.Println("Error receiving buffer:", err)
        return err
    }

    // Convert the buffer to a slice of uint64
    bufferUint64 := byteSliceToUint64SliceBE(buffer)

    //traverse the buffer and remove the id.bin from files directory

    for i:=0;i<len(bufferUint64);i++{
        os.Remove(fmt.Sprintf("files/%d.bin",bufferUint64[i]))
    }

    //read blocklist.bin
    blocklist, err := ioutil.ReadFile("blocklist.bin")
    if err != nil {
        fmt.Println("Error reading blocklist:", err)
        return err
    }

    // Convert the blocklist to a slice of uint64
    blocklistUint64 := byteSliceToUint64SliceBE(blocklist)

    //sort the slices
    sort.Slice(bufferUint64, func(i, j int) bool { return bufferUint64[i] < bufferUint64[j] })
    sort.Slice(blocklistUint64, func(i, j int) bool { return blocklistUint64[i] < blocklistUint64[j] })

    //find the differences
    differences := findDifferences(blocklistUint64,bufferUint64)

    //convert the differences to a byte slice
    differencesByteSlice := uint64SliceToByteSliceBE(differences)

    //write it back to blocklist.bin
    err = ioutil.WriteFile("blocklist.bin", differencesByteSlice, 0644)
    if err != nil {
        fmt.Println("Error writing blocklist:", err)
        return err
    }




    

    fmt.Println("File sent successfully.")
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
            // Skip common elements
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
