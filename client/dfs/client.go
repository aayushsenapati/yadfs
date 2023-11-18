package dfs

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"encoding/binary"
)


func SendCmd(ip,port string) {

	conn, err := net.Dial("tcp",ip+":"+port)
	if(err != nil){
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	argv := os.Args
	if(len(argv) == 0){
		fmt.Println("No arguments provided")
	}else{
		switch(argv[1]){
		case "mkdir":
			message := argv[1] + " " + argv[2]
			fmt.Println("Sending message:", message)
			byteBuffer := []byte(message)

			_, err := conn.Write(byteBuffer)
			if(err != nil){
				fmt.Println("Error sending message:", err)
				os.Exit(1)
			}

		case "put":
			// Get the source file name and destination from the command line arguments
			if len(argv) < 3 {
				fmt.Println("Usage: put destination source")
				os.Exit(1)
			}
			src := argv[2]
			dest := argv[3]

			// Get the size of the source file
			fileInfo, err := os.Stat(src)
			if err != nil {
				fmt.Println("Error getting file info:", err)
				os.Exit(1)
			}
			size := fileInfo.Size()

			// Create the message
			message := fmt.Sprintf("put %s %d", dest, size)
			fmt.Println("Sending message:", message)
			byteBuffer := []byte(message)

			_, err = conn.Write(byteBuffer)
			if err != nil {
				fmt.Println("Error sending message:", err)
				os.Exit(1)
			}

			go func() {
				conn2, err := net.Dial("tcp", ip+":2201")
				if err != nil {
					fmt.Println("Error connecting to server on another port:", err)
					return
				}
				defer conn2.Close()

				headerBuf := make([]byte, 1024)
				n, err := conn2.Read(headerBuf)
				if err != nil {
					return 
				}
				header := string(headerBuf[:n])
				fileSizeStr:=strings.Split(header,":")[1]
				fileSize, err := strconv.Atoi(fileSizeStr)
				if err != nil {
					return
				}
				var buf bytes.Buffer

				// Copy data from the connection to the buffer
				_, err = io.CopyN(&buf, conn2, int64(fileSize))
				if err != nil {
					fmt.Println("Error copying data:", err)
					return
				}

				// Convert the buffer to a string
				data := buf.String()
				func(){
					//data string is split \n and stored in an array
					lines:=strings.Split(data,"\n")

					blockSize := 128 * 1024 // 128 KB
					inputFile, err := os.Open(src)
					if err != nil {
						fmt.Println("Error opening file:", err)
						return
					}
					defer inputFile.Close()

					for i := 0; i < len(lines); i += 3 {
						//parse the first line to get ip(format is ip:blockid) and every three lines corresponds to a block
						parts:=strings.Split(lines[i],":")
						ip,blockid:=parts[0],parts[1]
						//split file at src into len(lines)/3 blocks and send each block to the corresponding ip


						// Read and write the block to the output file
						buffer := make([]byte, blockSize)
						f_size, err := inputFile.Read(buffer)
						if err != nil {
							fmt.Println("Error reading from input file:", err)
							return
						}

						_, err = inputFile.Seek(int64(blockSize), os.SEEK_CUR)
						if err != nil {
							fmt.Println("Error seeking input file:", err)
							return
						}
						
						conn3,err:=net.Dial("tcp",ip+":3200")
						defer conn3.Close()
						headerBuf := make([]byte,3+8+8+1)
						copy(headerBuf[:3],[]byte("put"))


						blockIDUint64, err := strconv.ParseUint(blockid, 10, 64)
						if err != nil {
							fmt.Println("Error converting blockid to uint64:", err)
							return
						}
						binary.BigEndian.PutUint64(headerBuf[3:],blockIDUint64)
						binary.BigEndian.PutUint64(headerBuf[11:],uint64(f_size))
						headerBuf[19]=byte(3)
						headerBuf = append(headerBuf, []byte(lines[i+1]+"\n"+lines[i+2])...)

						_, err = conn3.Write(headerBuf)
						if err != nil {
							fmt.Println("Error sending message:", err)
							return
						}	
					}
					fmt.Println("Received data:", len(data))
					return 


				}()


			}()
			returnBuf := make([]byte, 1024)
			n, err := conn.Read(returnBuf)
			if err != nil {
				fmt.Println("Error reading:", err)
				return
			}
			returnMessage := string(returnBuf[:n])
			fmt.Println("Received message:", returnMessage)
		}
	}
}
