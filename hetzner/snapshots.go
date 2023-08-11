package hetzner

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Lists hetzplay's snapshots for a server.
func ListSnapshots(client *hcloud.Client, serverID int64) (snapshots []*hcloud.Image) {

	// Define image list options
	opts := hcloud.ImageListOpts{
		Type: []hcloud.ImageType{"snapshot"},
	}

	// Retrieve list of images
	images, _, err := client.Image.List(context.Background(), opts)
	if err != nil {
		log.Fatalf("error retrieving snapshots: %s\n", err)

	}

	// Filter to images hetzplay created for this server
	for _, image := range images {
		fmt.Println("Found image", image.Description)
		if strings.HasPrefix(image.Description, "hetzplay_") {
			snapshots = append(snapshots, image)
		}
	}

	return snapshots
}
