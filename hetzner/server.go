package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Gets the server with the specified name
func (p HetznerProvider) getServer(serverName string) (server *hcloud.Server, err error) {
	server, _, err = p.client.Server.GetByName(context.Background(), serverName)
	if err != nil {
		p.log.Errorf("error getting server %s: %e", serverName, err)
	}
	return
}

func (p HetznerProvider) doesServerExist(serverName string) (exists bool, err error) {
	server, err := p.getServer(serverName)
	if err != nil {
		return false, err
	}
	exists = server != nil
	return
}

func (p HetznerProvider) doesManagedSnapshotExist(serverName string) (exists bool, err error) {
	snaps, err := p.listSnapshots(serverName)
	if err != nil {
		return false, err
	}
	exists = len(snaps) > 0
	return
}
