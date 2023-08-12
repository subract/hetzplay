package hetzner

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Create a snapshot of a server.
func TakeSnapshot(client *hcloud.Client, server *hcloud.Server, id uint) (snapshot hcloud.Image, err error) {
	// Define snapshot options
	//
	// For whatever reason, Server.CreateImage is the _only_ function in the entire library
	// that wants a pointer to its opts, not the value of the opts.
	// TODO: Submit a PR to fix this
	// https://github.com/hetznercloud/hcloud-go/blob/f68e8530c9c3e94cd3a35b8d1d280335f124ebb8/hcloud/server.go#L658
	opts := hcloud.Ptr(hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Description: hcloud.Ptr(fmt.Sprintf("hetzplay_%d", id)),
		Labels: map[string]string{
			"hetzplay_id":          fmt.Sprintf("%d", id),
			"hetzplay_server_name": server.Name,
		},
	})

	// Create the image
	// TODO: Figure out how Actions work, in order to wait for this to complete
	// TODO: Print snapshot progress
	res, _, err := client.Server.CreateImage(context.Background(), server, opts)

	snapshot = *res.Image
	return snapshot, err
}

// List hetzplay's snapshots for a named server.
func ListSnapshots(client *hcloud.Client, serverName string) (snapshots []*hcloud.Image, err error) {
	// Define selectors for snapshots hetzplay creates
	selector := fmt.Sprintf("hetzplay_id, hetzplay_server_name==%s", serverName)
	types := []hcloud.ImageType{hcloud.ImageTypeSnapshot}

	// Define image list options
	opts := hcloud.ImageListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: selector,
		},
		Type: types,
	}

	// Retrieve list of snapshots
	snapshots, _, err = client.Image.List(context.Background(), opts)

	return
}
