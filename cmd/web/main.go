// Webserver CLI ./cmd/web/main.go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/steadyloop/truetone/internal/web"
	"gopkg.in/yaml.v3"
)

type config struct {
	TempDir  string
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Metadata struct {
		Backend string `yaml:"backend"`
		Opts    string `yaml:"options"`
	} `yaml:"metadata"`
	Content struct {
		Backend string `yaml:"backend"`
		Opts    string `yaml:"options"`
	} `yaml:"content"`
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	// 0. Resolve config file
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()
	if len(flag.Args()) > 0 {
		flag.Usage()
		os.Exit(2)
	}
	data, err := os.ReadFile(*configPath)
	if err != nil {
		logger.Error("failed to read config", "path", *configPath, "error", err)
		os.Exit(1)
	}
	var config config
	if err := yaml.Unmarshal(data, &config); err != nil {
		logger.Error("failed to parse config", "path", *configPath, "error", err)
		os.Exit(1)
	}
	// 1. Establish temp directory for the server
	tempDir := config.TempDir
	os.MkdirAll(tempDir, 0755)
	// 2. Establish metadata backend scheme
	var metadata web.MetadataStore
	logger.Info("establishing metadata backend", "backend", config.Metadata.Backend, "opts", config.Metadata.Opts)
	switch config.Metadata.Backend {
	case "sqlite":
		store, err := web.NewSQLiteMetadataStore(config.Metadata.Opts)
		if err != nil {
			logger.Error("failed to create sqlite metadata store", "error", err)
			os.Exit(1)
		}
		metadata = store
	default:
		logger.Error("unsupported metadata backend", "backend", config.Metadata.Backend)
		os.Exit(1)
	}
	// 3. Establish content backend scheme
	var content web.ContentStore
	logger.Info("establishing content backend", "backend", config.Content.Backend, "opts", config.Content.Opts)
	switch config.Content.Backend {
	case "fs":
		content = web.NewFileSystemStore(config.Content.Opts)
	default:
		logger.Error("unsupported content backend", "backend", config.Content.Backend)
		os.Exit(1)
	}
	// 4. Start the server
	server := web.NewServer(metadata, content, tempDir)
	listenAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Error("failed to start listener", "address", listenAddr, "error", err)
		os.Exit(1)
	}
	defer lis.Close()
	logger.Info("starting web server", "address", listenAddr)
	if err := server.Start(lis); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
