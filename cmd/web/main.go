// Webserver CLI ./cmd/web/main.go
// Run: mkdir -p storage && go run ./cmd/web -port 8080 sqlite ./metadata.db fs ./storage

package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/steadyloop/truetone/internal/web"
)

func printUsage() {
	fmt.Println("Usage: ./program [OPTIONS] METADATA_TYPE METADATA_OPTIONS CONTENT_TYPE CONTENT_OPTIONS")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  METADATA_TYPE         Metadata service type (sqlite)")
	fmt.Println("  METADATA_OPTIONS      Options for metadata service (e.g., db path)")
	fmt.Println("  CONTENT_TYPE          Content service type (fs)")
	fmt.Println("  CONTENT_OPTIONS       Options for content service (e.g., base dir)")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example: ./program sqlite db.db fs /path/to/videos")
}

func main() {
	// 1. Flags for the web cli
	port := flag.Int("port", 8080, "Port number for the web server")
	host := flag.String("host", "localhost", "Host address for the web server")
	flag.Usage = printUsage
	flag.Parse()
	if len(flag.Args()) != 4 {
		fmt.Println("Error: Incorrect number of arguments")
		printUsage()
		return
	}
	metadataServiceType, metadataServiceOptions := flag.Arg(0), flag.Arg(1)
	contentServiceType, contentServiceOptions := flag.Arg(2), flag.Arg(3)
	if *port <= 0 {
		fmt.Println("Error: Invalid port number:", *port)
		printUsage()
		return
	}

	// 2. Metadata service
	var metadataService web.VideoMetadataService
	fmt.Println("Creating metadata service of type", metadataServiceType, "with options", metadataServiceOptions)
	switch metadataServiceType {
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
