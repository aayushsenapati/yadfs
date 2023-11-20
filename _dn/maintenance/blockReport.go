package maintenance

import (
    "fmt"
    "net"
    "os"
    "time"
	"io"
    "encoding/binary"
    "io/ioutil"
)


/* func readIDFromFile(filename string) (uint64, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	id := binary.BigEndian.Uint64(data)
	return id, nil
} */



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
    fmt.Println("\n\n\n\n\n\n\nReceived buffer:", buffer)
    if err != nil {
        fmt.Println("Error receiving buffer:", err)
        return err
    }

    // Traverse the buffer
    for i := 0; i < len(buffer); i += 8 {
        // Convert each 8 bytes to a uint64
        id := binary.BigEndian.Uint64(buffer[i : i+8])

        // Delete the .bin file with the matching name
        err := os.Remove(fmt.Sprintf("files/%d.bin", id))
        if err != nil {
            fmt.Println("Error deleting file:", err)
            continue
        }
        // Read the blocklist.bin file into a slice
        blocklist, err := ioutil.ReadFile("blocklist.bin")
        if err != nil {
            fmt.Println("Error reading blocklist:", err)
            return err
        }

        // Convert the blocklist to a slice of uint64
        blocklistUint64 := make([]uint64, len(blocklist)/8)
        for j := range blocklistUint64 {
            blocklistUint64[j] = binary.BigEndian.Uint64(blocklist[j*8 : (j+1)*8])
        }

        // Remove the id from the blocklist
        for j, blockId := range blocklistUint64 {
            if blockId == id {
                blocklistUint64 = append(blocklistUint64[:j], blocklistUint64[j+1:]...)
                break
            }
        }

        // Convert the blocklist back to a slice of bytes
        blocklist = make([]byte, len(blocklistUint64)*8)
        for j, blockId := range blocklistUint64 {
            binary.BigEndian.PutUint64(blocklist[j*8:(j+1)*8], blockId)
        }

        // Write the blocklist back to the file
        err = ioutil.WriteFile("blocklist.bin", blocklist, 0644)
        if err != nil {
            fmt.Println("Error writing blocklist:", err)
            return err
        }
    }



    

    fmt.Println("File sent successfully.")
    return nil
}




/* func receTCP(conn net.Conn, dataPipe chan Packet) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err)
			return
		}
		dataPipe <- Packet{data: buf[:n], addr: conn.RemoteAddr()}
	}
} */