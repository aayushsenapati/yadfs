package filemanager

import (
    "fmt"
    "net"
    "encoding/binary"
    "os"
    "bytes"
    "io"
)



func ClientListener(ip, port string) {
    ipString := ip + ":" + port
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
        go handleNewClient(conn)
    }
}


func handleNewClient(conn net.Conn) {
	commandbuf:=make([]byte, 1024)
	n,err:=conn.Read(commandbuf)
	if err!=nil {
		fmt.Println("Error reading command:", err)
	}
    commandbuf=commandbuf[:n]
    command:=string(commandbuf[:3])
	switch command {
	case "put":
		fmt.Println("Put command received!")
        blockid := binary.BigEndian.Uint64(commandbuf[3:11])
        f_size := binary.BigEndian.Uint64(commandbuf[11:19])
        replicationFactor := commandbuf[19]
        remainingBytes := commandbuf[20:]
        _,err=conn.Write([]byte("ack"))
        if err!=nil {
            fmt.Println("Error sending ack:", err)
        }
        //receive fileblock from conn and write to disk and send to other datanodes from remaining bytes
            // Create a buffer to hold the file block
        fileBlock := make([]byte, f_size)
        _, err = conn.Read(fileBlock)
        if err != nil {
            fmt.Println("Error reading file block:", err)
            return
        }


        // Create the files directory if it doesn't exist
        err := os.MkdirAll("files", 0755)
        if err != nil {
            fmt.Println("Error creating files directory:", err)
            return
        }
        // Write the file block to disk
        outputFile, err := os.Create(fmt.Sprintf("files/%d.bin", blockid))
        if err != nil {
            fmt.Println("Error creating output file:", err)
            return
        }
        _, err = outputFile.Write(fileBlock)
        if err != nil {
            fmt.Println("Error writing to output file:", err)
            return
        }
        outputFile.Close()

        // Append blockid to blocklist.bin
        blockListFile, err := os.OpenFile("blocklist.bin", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            fmt.Println("Error opening blocklist file:", err)
            return
        }
        err = binary.Write(blockListFile, binary.BigEndian, blockid)
        if err != nil {
            fmt.Println("Error writing to blocklist file:", err)
            return
        }
        blockListFile.Close()

            // Send the file block to other data nodes
        for i := 0; i < int(replicationFactor); i++ {
            ip := net.IP(remainingBytes[i*12 : i*12+4]).String()
            conn2, err := net.Dial("tcp", ip+":3200")
            if err != nil {
                fmt.Println("Error connecting to data node:", err)
                continue
            }

                    // Replace blockid in commandbuf
            newBlockId := binary.BigEndian.Uint64(remainingBytes[i*12+4 : i*12+12])
            binary.BigEndian.PutUint64(commandbuf[3:], newBlockId)

            // Set replication factor to 0 in commandbuf
            commandbuf[19] = byte(0)

            _, err = conn2.Write(commandbuf)
            if err != nil {
                fmt.Println("Error sending file block to data node:", err)
            }
            
            ackBuf := make([]byte, 3)
            _, err = conn2.Read(ackBuf)
            if err != nil {
                fmt.Println("Error reading:", err)
                return
            }
            fmt.Println("Received ack:", ackBuf)
            if string(ackBuf) != "ack" {
                fmt.Println("Error receiving ack")
                return
            }
            _,err=conn2.Write(fileBlock[:f_size])
            if err != nil {
                fmt.Println("Error sending block:", err)
                return
            }
            fmt.Println("Sent block to datanode")
            conn2.Close()
        }

        _,err=conn.Write([]byte("ack"))
        if err!=nil {
            fmt.Println("Error sending ack:", err)
        }
        fmt.Println("Put command completed!")

    case "get":
        // Parse the block ID from the command
        blockID := binary.BigEndian.Uint64(commandbuf[3:])
        if err != nil {
            fmt.Println("Error parsing block ID:", err)
            return
        }
    
        // Open the block file
        blockFile, err := os.Open(fmt.Sprintf("files/%d.bin", blockID))
        if err != nil {
            fmt.Println("Error opening block file:", err)
            return
        }
        defer blockFile.Close()
    
        // Get the size of the block file
        fileInfo, err := blockFile.Stat()
        if err != nil {
            fmt.Println("Error getting file info:", err)
            return
        }
        size := fileInfo.Size()
    
        // Send the size of the block file
        var sizeBuf bytes.Buffer
        binary.Write(&sizeBuf, binary.BigEndian, size)
        _, err = conn.Write(sizeBuf.Bytes())
        if err != nil {
            fmt.Println("Error sending file size:", err)
            return
        }
    
        // Wait for the ack
        ackBuf := make([]byte, 3)
        _, err = conn.Read(ackBuf)
        if err != nil {
            fmt.Println("Error reading ack:", err)
            return
        }
    
        // Check the ack
        if string(ackBuf) != "ack" {
            fmt.Println("Did not receive ack")
            return
        }
    
        // Send the block file
        _, err = io.Copy(conn, blockFile)
        if err != nil {
            fmt.Println("Error sending block file:", err)
            return
        }

	}

}