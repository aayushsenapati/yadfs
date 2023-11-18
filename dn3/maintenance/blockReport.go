package maintenance

import (
    "fmt"
    "net"
    "os"
    "time"
	"io"
    "encoding/binary"
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
        err := sendFile(conn, "blockreport.txt",id)
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
    header := fmt.Sprintf("blockreport:%d", fileInfo.Size())
	byteBuffer:=make([]byte, len(header)+1)
	binary.BigEndian.PutUint8(byteBuffer, id)
	copy(byteBuffer[1:], []byte(header))
    _, err = conn.Write(byteBuffer)
    if err != nil {
        return err
    }

    // Send the file content
    _, err = io.Copy(conn, file)
    if err != nil {
        return err
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