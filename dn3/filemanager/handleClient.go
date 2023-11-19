package filemanager

import (
    "fmt"
    "net"
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
	_,err:=conn.Read(commandbuf)
	if err!=nil {
		fmt.Println("Error reading command:", err)
	}
    command:=string(commandbuf[:3])
	switch command {
	case "put":
		fmt.Println("Put command received!")
	}

}