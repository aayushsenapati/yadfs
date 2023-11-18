package dnServer

import (
    "fmt"
    "net"
    "os"
    "io"
    "strconv"
    "strings"
)

func receiveFile(conn net.Conn) error {
    // Read the header containing the file name and size
    headerBuf := make([]byte, 1024)
    n, err := conn.Read(headerBuf)
    if err != nil {
        return err
    }

    id := uint8(headerBuf[0])
    header := string(headerBuf[1:n])
    fileSizeStr:=strings.Split(header,":")[1]
    fileSize, err := strconv.Atoi(fileSizeStr)
    if err != nil {
        return err
    }
    

    // Create the directory if it doesn't exist
    err = os.MkdirAll("blockreports", 0755)
    if err != nil {
        fmt.Println("Error creating directory:", err)
        return err
    }
    // Create a new file for writing
    file, err := os.Create("blockreports/" + strconv.Itoa(int(id)))
    if err != nil {
        fmt.Println("Error creating file:", err)
        return err
    }
    defer file.Close()

    // Copy data from the connection to the file
    _, err = io.CopyN(file, conn, int64(fileSize))
    if err != nil {
        return err
    }

    fmt.Println("File received successfully.")
    return nil
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