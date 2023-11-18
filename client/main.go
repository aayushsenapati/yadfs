package main

import (
	"client/dfs"
)

func main(){
	serverIp := "nn-container-devel"
	dfs.SendCmd(serverIp, "2200")
}
