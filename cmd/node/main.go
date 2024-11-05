package main

import (
	"flag"

	"github.com/fydmer/fileserver/internal/app"
	"github.com/fydmer/fileserver/internal/repositories/diskfile"
	"github.com/fydmer/fileserver/internal/servers/tcpserver"
	"github.com/fydmer/fileserver/internal/services/node"
)

type Config struct {
	Port    int
	RootDir string
}

func main() {
	a := app.NewApp()

	config := &Config{}
	{
		flag.IntVar(&config.Port, "port", 8123, "Port to listen on")
		flag.StringVar(&config.RootDir, "root-dir", "./data", "Root directory to serve files from")
		flag.Parse()
	}

	diskfileRepo, err := diskfile.NewRepository(config.RootDir)
	if err != nil {
		a.Panic(err)
	}

	nodeService := node.NewNode(diskfileRepo)

	server, err := tcpserver.RunNodeServer(a.Context(), config.Port, nodeService)
	if err != nil {
		a.Panic(err)
	}
	a.AddStopFn(server.Close)

	a.Keep()
}
