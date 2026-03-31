// Command example demonstrates basic usage of the go_ytmusicapi client.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/JanikSachs/go_ytmusicapi/pkg/ytmusic"
)

func main() {
	client, err := ytmusic.NewClient()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	query := "Bohemian Rhapsody"
	if len(os.Args) > 1 {
		query = os.Args[1]
	}

	fmt.Printf("Searching for: %q\n", query)

	ctx := context.Background()
	results, err := client.Search(ctx, query, ytmusic.SearchOptions{
		Filter: "songs",
		Limit:  5,
	})
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	fmt.Printf("Found %d results:\n", len(results))
	for i, r := range results {
		b, _ := json.MarshalIndent(r, "  ", "  ")
		fmt.Printf("[%d] %s\n", i+1, b)
	}
}
