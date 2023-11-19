package filemanager

import (
    "fmt"
    "net"
    "encoding/binary"
    "os"
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
	fmt.Println("New client connected!")
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
	}

}