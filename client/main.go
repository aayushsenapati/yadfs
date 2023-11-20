package main

import (
	"client/dfs"
	"os"
)

func main(){
	serverIp := os.Getenv("NN_CONTAINER_NAME")
	dfs.SendCmd(serverIp, "2200")
}
