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
	select {}
}
