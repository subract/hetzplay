package hetzner

import (
	"context"
	"encoding/json"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

// Gets the server with the specified name
func (p HetznerProvider) getServer(serverName string) (server *hcloud.Server, err error) {
	server, _, err = p.client.Server.GetByName(context.Background(), serverName)
	if err != nil {
		p.log.Errorf("error getting server %s: %e", serverName, err)
	}
	serverJSON, err := json.Marshal(server)
	p.log.Debug("Found server", string(serverJSON))
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

// func (p HetznerProvider) createServer(name string,
// 	typeName string,
//   location string,
//   SSHKeys
//   ) (err error) {
// 	// Get the most recent snapshot
// 	snapshots, err := p.listSnapshots(name)
// 	if err != nil {
// 		return
// 	}

// 	serverType, _, err := p.client.ServerType.GetByName(context.Background(), typeName)
// 	if err != nil {
// 		return
// 	}

// 	opts := hcloud.ServerCreateOpts{
// 		Name:       name,
// 		Image:      snapshots[len(snapshots)-1],
// 		ServerType: serverType,
// 	}

// 	// type ServerCreateOpts struct {
// 	// 	Name             string
// 	// 		ServerType       *ServerType
// 	// 	Image            *Image
// 	// 		SSHKeys          []*SSHKey
// 	// 		Location         *Location
// 	// 		Datacenter       *Datacenter
// 	// 			UserData         string
// 	// 	StartAfterCreate *bool
// 	// 		Labels           map[string]string
// 	// 		Automount        *bool
// 	// 		Volumes          []*Volume
// 	// 		Networks         []*Network
// 	// 		Firewalls        []*ServerCreateFirewall
// 	// 		PlacementGroup   *PlacementGroup
// 	// 		PublicNet        *ServerCreatePublicNet
// 	// }

// }
