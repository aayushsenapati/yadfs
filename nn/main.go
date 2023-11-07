package main

import (
    "dnServer.go"
    "clServer.go"
)

func main() {
    go dnServer.Listen("172.21.0.2","12345")
    go clServer.Listen("172.21.0.2","12344")
    
}
