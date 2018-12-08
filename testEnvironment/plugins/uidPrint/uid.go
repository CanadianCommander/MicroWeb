package main

import (
	"fmt"
	"net/http"
	"syscall"
)

//HandleVirtualRequest writes the uid of this process to res.
func HandleVirtualRequest(req *http.Request, res http.ResponseWriter) bool {
	uid := syscall.Getuid()
	gid := syscall.Getgid()
	fmt.Fprintf(res, "Process uid: %d Process gid: %d", uid, gid)
	return true
}
