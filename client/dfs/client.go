package dfs

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
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
					file, err := os.Open(src)
					if err != nil {
						fmt.Println("Error opening file:", err)
						return
					}
					defer file.Close()

					for(i:=0;i<len(lines);i+=3){
						//parse the first line to get ip(format is ip:blockid) and every three lines corresponds to a block
						ip:=strings.Split(lines[i],":")[0]
						blockid:=strings.Split(lines[i],":")[1]
						//split file at src into len(lines)/3 blocks and send each block to the corresponding ip
						outputFile, err := os.Create(blockid+".bin")
						if err != nil {
							fmt.Println("Error creating output file:", err)
							return
						}
						defer outputFile.Close()


						// Read and write the block to the output file
						buffer := make([]byte, blockSize)
						_, err = inputFile.Read(buffer)
						if err != nil {
							fmt.Println("Error reading from input file:", err)
							return
						}

						blockSize, err = outputFile.Write(buffer)
						if err != nil {
							fmt.Println("Error writing to output file:", err)
							return
						}

						_, err = inputFile.Seek(offset, os.SEEK_CUR)
						if err != nil {
							fmt.Println("Error seeking input file:", err)
							return
						}
						
						net.Dial("tcp",ip+":3200")
						buffer := make([]byte,8+8+1)
						binary.BigEndian.PutUint64(buffer,blockid)
						binary.BigEndian.PutUint64(buffer[8:],uint64(blockSize))
						buffer[8]=byte(3)
						buffer.append(lines[i+1] + "\n" + lines[i+2])
						_, err := conn.Write(buffer)
						if err != nil {
							fmt.Println("Error sending message:", err)
							return


						}
						err := os.Remove(blockid+".bin")
						if err != nil {
							fmt.Println("Error deleting file:", err)
							return
						}






					}
					fmt.Println("Received data:", len(data))
					return 


				}()


			}
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
