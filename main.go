package main

import (
	"godrive/config"
	"godrive/master"
	"godrive/slave"
)

func main() {
	config.LoadConfig()
	master.ConfigureMasterTcpServices()
	slave.StartSlaveNodes()
	go master.StartMasterHttp()
	// log.Println("hey")
	// time.Sleep(5 * time.Second)
	// log.Println("hey")
	// var t config.Node
	// t.Host = "127.0.0.1"
	// t.Port = "6001"
	// master.SendDataToSlave(t, "helo")
	// master.RequestDataFromSlave(t, "MyDaata")
	select {}
}
