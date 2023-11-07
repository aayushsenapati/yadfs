package main

import
(
	"net"
	"fmt"
	"os"
	"time"
)


func main() {
	serverIp := "172.21.0.4"
	serverPort:="12344"
	conn, err := net.Dial("tcp",serverIp+":"+serverPort)
	if(err != nil){
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	argv := os.Args
	if(len(argv) == 0)
	{
		//make interactive fn
	}
	else if(len(argv) == 1){
		fmt.Println("Path not specified")
		os.Exit(1)
	}
	else{
		message := argv[1] + " " + argv[2]
		fmt.Println("Sending message:", message)
		byteBuffer := []byte(message)

		_, err := conn.Write(byteBuffer)
		if(err != nil){
			fmt.Println("Error sending message:", err)
			os.Exit(1)
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
	return 0
	
}
