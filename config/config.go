package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type MasterConfig struct {
	Host              string `yaml:"host"`
	HttpPort          int    `yaml:"http_port"`
	TcpPort           int    `yaml:"tcp_port"`
	WebSocketPort     int    `yaml:"ws_port"`
	HeartBeatInterval int    `yaml:"heartbeat_interval"`
	LockTimeout       int    `yaml:"lock_timeout"`
	ChunkSize         int    `yaml:"chunk_size"`
	ReplicationFactor int    `yaml:"replication_factor"`
	ReadQuorum        int    `yaml:"read_quorum"`
	WriteQuorum       int    `yaml:"write_quorum"`
}

type Node struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	Healthy bool   `yaml:"health"`
}
type Config struct {
	Master      MasterConfig
	SlaveNodes  []Node `yaml:"slaveDataNodes"`
	BackupNodes []Node `yaml:"backupNodes"`
}

var ReadConfig Config

func LoadConfig() {
	file, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatal("Couldn't load config", err)
	}
	err = yaml.Unmarshal(file, &ReadConfig)
	if err != nil {
		log.Fatal("Couldn't read config")
	}
	log.Println("Config Loaded Sucessfully!!!")
}
