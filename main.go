package main

import (
	"github.com/vinsec/sess-manager/manager"
	_ "github.com/vinsec/sess-manager/providers"
)

var globalSessions *manager.Manager

func init() {
	globalSessions, _ = manager.NewManager("memory", "gosessionid", 3600)
	go globalSessions.GC()
}

func main() {

}
