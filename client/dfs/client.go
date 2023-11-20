package dfs

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"encoding/binary"
	"sync"
	"path/filepath"

)


func SendCmd(ip,port string) {
	var wg sync.WaitGroup
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
				fmt.Println("Usage: put source destination")
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
			fmt.Println("Sent command")
			returnBuf := make([]byte, 1024)
			n, err := conn.Read(returnBuf)
			if err != nil {
				fmt.Println("Error reading:", err)
				return
			}
			returnMessage := string(returnBuf[1:n])
			myid:=uint8(returnBuf[0])
			fmt.Println("Received message:", returnMessage)

			wg.Add(1)


			go func() {
				conn2, err := net.Dial("tcp", ip+":2201")
				if err != nil {
					fmt.Println("Error connecting to server on another port:", err)
					return
				}
				defer conn2.Close()



				var idbuf bytes.Buffer
				idbuf.WriteByte(myid)
				idbuf.WriteString("put")
				fmt.Println(idbuf.Bytes()) // replace with your command

				// Send the buffer over the connection
				_, err = conn2.Write(idbuf.Bytes())
				if err != nil {
					fmt.Println("Error sending id and command:", err)
					return
				}
				fmt.Println("Sent id and command")

				headerBuf := make([]byte, 8)
				_, err = conn2.Read(headerBuf)
				if err != nil {
					return 
				}
				fmt.Println("Received header:", headerBuf)

				_,err=conn2.Write([]byte("ack"))
				if err != nil {
					fmt.Println("Error sending ack:", err)
					return
				}
				fmt.Println("Sent ack")
				


				listSize := binary.BigEndian.Uint64(headerBuf[:8])
				
				var buf bytes.Buffer

				// Copy data from the connection to the buffer
				_, err = io.CopyN(&buf, conn2, int64(listSize))
				if err != nil {
					fmt.Println("Error copying data:", err)
					return
				}
				fmt.Println("Received data:", buf.Bytes())
				// Convert the buffer to a string
				data := buf.Bytes()
				func() {
					blockSize := 128 * 1024 // 128 KB
					inputFile, err := os.Open(src)
					if err != nil {
						fmt.Println("Error opening file:", err)
						return
					}
					defer inputFile.Close()
				
					for i := 0; i < len(data); i += (12*3) {
						// Parse the IP and block ID from the data
						ip := net.IP(data[i : i+4]).String()
						blockid := binary.BigEndian.Uint64(data[i+4 : i+12])
				
						// Read and write the block to the output file
						blockbuffer := make([]byte, blockSize)
						f_size, err := inputFile.Read(blockbuffer)
						if err != nil {
							if err == io.EOF {
								fmt.Println("End of file reached")
							} else {
								fmt.Println("Error reading from input file:", err)
								return
							}
						}
				
						conn3, err := net.Dial("tcp", ip+":3200")
						if err != nil {
							fmt.Println("Error connecting to server:", err)
							return
						}
						defer conn3.Close()
						fmt.Println("Connected to server")
						headerBuf := make([]byte, 3+8+8+1)
						copy(headerBuf[:3], []byte("put"))
						binary.BigEndian.PutUint64(headerBuf[3:], blockid)
						binary.BigEndian.PutUint64(headerBuf[11:], uint64(f_size))
						headerBuf[19] = byte(3)
						headerBuf = append(headerBuf, data[i+12:i+36]...)
				
						_, err = conn3.Write(headerBuf)
						if err != nil {
							fmt.Println("Error sending message:", err)
							return
						}
						
						//upon receiving ack send buffer
						ackBuf := make([]byte, 3)
						_, err = conn3.Read(ackBuf)
						if err != nil {
							fmt.Println("Error reading:", err)
							return
						}
						fmt.Println("Received ack:", ackBuf)
						if string(ackBuf) != "ack" {
							fmt.Println("Error receiving ack")
							return
						}
						fmt.Println("\n\n\n\n\n\n\n",len(blockbuffer[:f_size]))
						_,err=conn3.Write(blockbuffer[:f_size])
						if err != nil {
							fmt.Println("Error sending block:", err)
							return
						}
						fmt.Println("Sent block")

						// Read the response
						responseBuf := make([]byte, 3)
						_, err = conn3.Read(responseBuf)
						if err != nil {
							fmt.Println("Error reading response:", err)
							return
						}
						if(string(responseBuf) != "ack"){
							fmt.Println("Error receiving ack")
							return
						}
						fmt.Println("file stored successfully")
					}
					return
				}()

				wg.Done()
			}()
			wg.Wait()
		case "get":
			// Get the source file name and destination from the command line arguments
			if len(argv) < 3 {
				fmt.Println("Usage: get source destination")
				os.Exit(1)
			}
			src := argv[2]
			dest := argv[3]

			// Get the size of the source file
			_, err := os.Stat(dest)
			if err == nil {
				fmt.Println("File exists")
				os.Exit(1)
			}
			// Get the parent directory of the destination file
			parentDir := filepath.Dir(dest)

			// Check if the parent directory exists
			_, err = os.Stat(parentDir)
			if os.IsNotExist(err) {
				fmt.Println("Parent directory does not exist")
				os.Exit(1)
			}


			// Create the message
			message := fmt.Sprintf("get %s", src)
			fmt.Println("Sending message:", message)
			byteBuffer := []byte(message)

			_, err = conn.Write(byteBuffer)
			if err != nil {
				fmt.Println("Error sending message:", err)
				os.Exit(1)
			}
			fmt.Println("Sent command")
			returnBuf := make([]byte, 1024)
			n, err := conn.Read(returnBuf)
			if err != nil {
				fmt.Println("Error reading:", err)
				return
			}
			returnMessage := string(returnBuf[1:n])
			myid:=uint8(returnBuf[0])
			fmt.Println("Received message:", returnMessage)

			wg.Add(1)


			go func() {
				conn2, err := net.Dial("tcp", ip+":2201")
				if err != nil {
					fmt.Println("Error connecting to server on another port:", err)
					return
				}
				defer conn2.Close()



				var idbuf bytes.Buffer
				idbuf.WriteByte(myid)
				idbuf.WriteString("get")
				fmt.Println(idbuf.Bytes()) // replace with your command

				// Send the buffer over the connection
				_, err = conn2.Write(idbuf.Bytes())
				if err != nil {
					fmt.Println("Error sending id and command:", err)
					return
				}
				fmt.Println("Sent id and command")

				headerBuf := make([]byte, 8)
				_, err = conn2.Read(headerBuf)
				if err != nil {
					return 
				}
				fmt.Println("Received header:", headerBuf)

				_,err=conn2.Write([]byte("ack"))
				if err != nil {
					fmt.Println("Error sending ack:", err)
					return
				}
				fmt.Println("Sent ack")
				


				listSize := binary.BigEndian.Uint64(headerBuf[:8])//filesize of blocklist
				
				var buf bytes.Buffer

				// Copy data from the connection to the buffer
				_, err = io.CopyN(&buf, conn2, int64(listSize))
				if err != nil {
					fmt.Println("Error copying data:", err)
					return
				}
				fmt.Println("Received data:", buf.Bytes())
				// Convert the buffer to a string
				data := buf.Bytes()
				func() {
					// Create the destination file
					destFile, err := os.OpenFile(dest, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Println("Error opening destination file:", err)
						os.Exit(1)
					}
					defer destFile.Close()//dnewjfbewhrbfherwfbwhrfbwhrefbhwerfbwhrebfhwrebfhwbfehbrfjwherbfjhwbrefjhbwrejbfwbjer
					for i := 0; i < len(data); i += 12 {
						// Parse the IP and block ID from the data
						ip := net.IP(data[i : i+4]).String()
						blockid := binary.BigEndian.Uint64(data[i+4 : i+12])
				
						// Dial to the IP with port 3200
						conn3, err := net.Dial("tcp", ip+":3200")
						if err != nil {
							fmt.Println("Error connecting to server:", err)
							return
						}
						defer conn3.Close()
				
						// Send the "get" command followed by the block ID
						var getBuf bytes.Buffer
						getBuf.WriteString("get")
						binary.Write(&getBuf, binary.BigEndian, blockid)
						_, err = conn3.Write(getBuf.Bytes())
						if err != nil {
							fmt.Println("Error sending get command:", err)
							return
						}
						blockSizeBuff:=make([]byte,8)
						_,err=conn3.Read(blockSizeBuff)
						if err != nil {
							fmt.Println("Error reading block size:", err)
							return
						}
						blockSize:=binary.BigEndian.Uint64(blockSizeBuff)
						conn3.Write([]byte("ack"))
						var blockBuffer bytes.Buffer
						_,err=io.CopyN(&blockBuffer,conn3,int64(blockSize))
						if err != nil {
							fmt.Println("Error reading block:", err)
							return
						}
						// Append the data from blockBuffer to destFile
						_, err = destFile.Write(blockBuffer.Bytes())
						if err != nil {
							fmt.Println("Error writing to destination file:", err)
							return
						}

					}
				}()

				wg.Done()
			}()
			wg.Wait()
		
		
		
		
		}
	}
}
