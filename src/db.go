package main

type DbConfig struct {
	Server   string `json:"server"`
	Database string `json:"database"`
	User     string `json:"user"`
	Password string `json:"password"`
	PoolSize int    `json:"connectionPoolSize"`
}
