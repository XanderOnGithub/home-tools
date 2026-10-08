// Command home-tools is the single server binary. It loads config, mounts each
// tool (tools/<name>) under its hostname, and serves API + embedded SPAs.
package main

import "fmt"

func main() {
	// TODO(xander): config load, tool registration, HTTP server.
	fmt.Println("home-tools")
}
