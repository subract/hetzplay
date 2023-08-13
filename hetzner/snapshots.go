package hetzner

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

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
		Sort: []string{"created:asc"},
	}

	// Retrieve list of snapshots
	snapshots, _, err = client.Image.List(context.Background(), opts)

	return
}

// Create a snapshot of a server.
func TakeSnapshot(client *hcloud.Client, server *hcloud.Server) (snapshot hcloud.Image, err error) {
	// Calculate ID of new snapshot
	snapshots, err := ListSnapshots(client, server.Name)
	if err != nil {
		return
	}
	newSnapID := 0
	if len(snapshots) > 0 {
		latestSnap := snapshots[len(snapshots)-1]
		newSnapID, err = strconv.Atoi(latestSnap.Labels["hetzplay_id"])
		if err != nil {
			return
		}
		newSnapID++
	}

	// Define snapshot options
	//
	// For whatever reason, Server. CreateImage is the _only_ function in the entire library
	// that wants a pointer to its opts, not the value of the opts.
	// TODO: Submit a PR to fix this
	// https://github.com/hetznercloud/hcloud-go/blob/f68e8530c9c3e94cd3a35b8d1d280335f124ebb8/hcloud/server.go#L658
	opts := hcloud.Ptr(hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Description: hcloud.Ptr(fmt.Sprintf("hetzplay_%s_%d", server.Name, newSnapID)),
		Labels: map[string]string{
			"hetzplay_id":          fmt.Sprintf("%d", newSnapID),
			"hetzplay_server_name": server.Name,
		},
	})

	// Create the image
	// TODO: Figure out how Actions work, use to wait for this to complete
	// TODO: Print snapshot progress
	res, _, err := client.Server.CreateImage(context.Background(), server, opts)
	if err != nil {
		return
	}

	snapshot = *res.Image
	return snapshot, err
}

// Cleans up old snapshots
func PruneSnapshots(client *hcloud.Client, serverName string, backupSnapCount uint) (err error) {
	// Always keep a single primary snapshot, regardless of how many backups the user requests
	desiredSnapCount := int(backupSnapCount) + 1

	// Get current snapshots
	snapshots, err := ListSnapshots(client, serverName)
	if err != nil {
		return
	}

	// Delete oldest snapshots until the desired count remains
	// ListSnapshots ensures oldest snaps are at head of the slice
	for i := 0; len(snapshots)-i > desiredSnapCount; i++ {
		snapToDelete := snapshots[i]

		_, err = client.Image.Delete(context.Background(), snapToDelete)
		if err != nil {
			return
		}
	}

	return
}
