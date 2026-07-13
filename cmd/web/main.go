// Webserver CLI ./cmd/web/main.go
// Run: mkdir -p storage && go run ./cmd/web -port 8080 sqlite ./metadata.db fs ./storage

package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/steadyloop/truetone/internal/web"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Host     string `yaml:"host"`
	Metadata struct {
		Type string `yaml:"type"`
		Opts string `yaml:"options"`
	} `yaml:"metadata"`
	Content struct {
		Type string `yaml:"type"`
		Opts string `yaml:"options"`
	} `yaml:"content"`
}

func main() {
	// Optional path to config file
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()
	if len(flag.Args()) > 1 {
		flag.PrintDefaults()
		return
	}
	data, err := os.ReadFile(*configPath)
	if err != nil {
		fmt.Errorf("failed to read config file: ", err)
		return
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Errorf("failed to load config file: ", err)
	}

	// 2. Setup scheme to setup metadata
	var metadataService web.MetadataStore
	fmt.Println("Creating metadata service of type", cfg.Metadata.Type, "with options", cfg.Metadata.Opts)
	switch cfg.Metadata.Type {
	case "sqlite":
		svc, err := web.NewSQLiteVideoMetadataService(metadataServiceOptions)

		// metadataService = web.NewSQLiteVideoMetadataService(metadataServiceOptions)
		// TODO: metadataService = web.NewSQLiteVideoMetadataService(metadataServiceOptions) (handle the error)
	default:
		fmt.Printf("Unsupported metadata service: %s\n", metadataServiceType)
		return
	}

	// 3. Content service
	var contentService web.VideoContentService
	fmt.Println("Creating content service of type", contentServiceType, "with options", contentServiceOptions)
	switch contentServiceType {
	case "fs":
		// TODO: contentService = web.NewFSVideoContentService(contentServiceOptions)
	default:
		fmt.Printf("Unsupported content service: %s\n", contentServiceType)
		return
	}

	if metadataService == nil || contentService == nil {
		fmt.Println("Services not constructed yet — finish the TODOs in the two switches above.")
		return
	}

	// Start the server
	server := web.NewServer(metadataService, contentService)
	listenAddr := fmt.Sprintf("%s:%d", *host, *port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println("Error starting listener:", err)
		return
	}
	defer lis.Close()
	fmt.Println("Starting web server on", listenAddr)
	if err := server.Start(lis); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
