package dfs

import
(
	"net"
	"fmt"
	"os"
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

		case "cp":
			// Get the source file name and destination from the command line arguments
			if len(argv) < 3 {
				fmt.Println("Usage: cp destination")
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
			message := fmt.Sprintf("cp %s %d", dest, size)
			fmt.Println("Sending message:", message)
			byteBuffer := []byte(message)
		
			_, err = conn.Write(byteBuffer)
			if err != nil {
				fmt.Println("Error sending message:", err)
				os.Exit(1)
			}
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