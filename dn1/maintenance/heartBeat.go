package maintenance

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

)

type Packet struct {
	data []byte
	addr net.Addr
}

func readIDFromFile(filename string) (uint8, error) {
    file, err := os.Open(filename)
    if err != nil {
        return 0, err
    }
    defer file.Close()

    var id uint8
    err = binary.Read(file, binary.BigEndian, &id)
    if err != nil {
        return 0, err
    }

    return id, nil
}

func writeIDToFile(filename string, id uint8) error {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, id)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

// variable for unique id for this datanode
func SendHb(ipString, portString string) {

	conn, err := net.Dial("tcp", ipString+":"+portString)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()
	var fileExists bool = true
	id, err := readIDFromFile("id.bin")
	if err != nil {
		fmt.Println("Error reading id from file:", err, "id does not exist")
		fileExists = false
		id = 0
	}

	dataPipe := make(chan Packet)
	ticker := time.NewTicker(3 * time.Second)
	timer := time.NewTimer(6 * time.Second)
	defer ticker.Stop()
	defer timer.Stop()
	message := "heartbeat"

	go receTCP(conn, dataPipe)
	for {
		select {
		case <-ticker.C:
			fmt.Println("Sending to server:", message)
			byteBuffer := make([]byte, len(message)+1)
			byteBuffer[0]=byte(id)
			copy(byteBuffer[1:], []byte(message))
			_, err := conn.Write(byteBuffer)
			if err != nil {
				fmt.Println("Error sending message:", err)
				os.Exit(1)
			}
		case packet := <-dataPipe:
			id = uint8(packet.data[0])
			if !fileExists {
				err = writeIDToFile("id.bin", id)
				if err != nil {
					fmt.Println("Error writing id to file:", err)
					os.Exit(1)
				}
				fileExists = true
			}
			fmt.Println("Received from server:", string(packet.data[1:]), id)
			timer.Reset(6 * time.Second)
		case <-timer.C:
			fmt.Println("Timeout, no response from server")
			conn.Close()
			conn, err = net.Dial("tcp", ipString)
			if err != nil {
				fmt.Println("Error connecting to server:", err)
				os.Exit(1)
			}
			defer conn.Close()
			go receTCP(conn, dataPipe)
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
