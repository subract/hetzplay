package hetzner

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// List hetzplay's snapshots for a named server.
func (p HetznerProvider) listSnapshots(serverName string) (snapshots []*hcloud.Image, err error) {
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
	p.log.Debugf("listing snapshots for %s", serverName)
	snapshots, _, err = p.client.Image.List(context.Background(), opts)
	for _, snapshot := range snapshots {
		p.log.Debugf("found snap %s (%s)", snapshot.Description, snapshot.Created)
	}

	return
}

// Create a snapshot of a server.
func (p HetznerProvider) takeSnapshot(serverName string) (err error) {
	// Get the server
	server, err := p.getServer(serverName)
	if err != nil {
		return
	}
	if server == nil {
		err = fmt.Errorf("server %s does not exist", serverName)
		return
	}

	// Calculate ID of new snapshot
	snapshots, err := p.listSnapshots(serverName)
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
		Description: hcloud.Ptr(fmt.Sprintf("hetzplay_%s_%d", serverName, newSnapID)),
		Labels: map[string]string{
			"hetzplay_id":          fmt.Sprintf("%d", newSnapID),
			"hetzplay_server_name": serverName,
		},
	})

	// Create the image
	// TODO: Figure out how Actions work, use to wait for this to complete
	// TODO: Print snapshot progress
	p.log.Infof("taking snapshot of %s", serverName)
	_, _, err = p.client.Server.CreateImage(context.Background(), server, opts)
	return
}

// Cleans up old snapshots
func (p HetznerProvider) pruneSnapshots(serverName string, desiredSnapCount int) (err error) {
	// Get current snapshots
	snapshots, err := p.listSnapshots(serverName)
	if err != nil {
		return
	}

	// Delete oldest snapshots until the desired count remains
	// listSnapshots ensures oldest snaps are at head of the slice
	for i := 0; len(snapshots)-i > desiredSnapCount; i++ {
		snapToDelete := snapshots[i]
		p.log.Info("deleting snapshot %s", snapToDelete.Description)
		_, err = p.client.Image.Delete(context.Background(), snapToDelete)
		if err != nil {
			return
		}
	}

	return
}
