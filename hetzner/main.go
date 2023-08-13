package hetzner

import (
	"fmt"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/op/go-logging"
)

// Manages an instance of a server on a cloud provider
type ServerManager struct {
	provider  CloudProvider
	name      string
	snapCount int
	log       *logging.Logger
}

// Creates a server manager
func NewServerManager(serverName string,
	backupSnapCount int,
	token string,
	appVer string,
	log *logging.Logger) (m *ServerManager, err error) {

	log.Debug("initializing server manager")
	m = &ServerManager{}
	m.name = serverName
	// Always keep one base snap, plus any backups
	m.snapCount = backupSnapCount + 1
	m.log = log
	m.provider = NewHetznerProvider(token, appVer, log)

	// Check if a managed snapshot for the given server exists
	exists, err := m.provider.doesManagedSnapshotExist(serverName)
	if err != nil {
		return nil, err
	}
	if exists {
		return
	}

	// If no snapshots exist yet, check if the server exists
	exists, err = m.provider.doesServerExist(serverName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf(`no server or snapshots found
    
		If this is the first time using hetzplay on this server, ensure the server '%s'
		exists with the cloud provider`, serverName)
	}

	// If a snapshot does not exist, but the server does, take an initial snapshot
	m.provider.takeSnapshot(serverName)
	return
}

type CloudProvider interface {
	// Determines if a server exists, running or otherwise
	doesServerExist(string) (bool, error)
	// Determine if a managed snapshot for a giver server exists
	doesManagedSnapshotExist(string) (bool, error)
	// Take a snapshot of a server
	takeSnapshot(string) error
	// Prune snapshots to a specified count of snapshots
	pruneSnapshots(string, int) error
	// createServer(string) error
	// destroyServer(string) error
}

type HetznerProvider struct {
	client *hcloud.Client
	log    *logging.Logger
}

func NewHetznerProvider(token string, appVer string, log *logging.Logger) (provider HetznerProvider) {
	provider.log = log
	provider.client = hcloud.NewClient(hcloud.WithToken(token), hcloud.WithApplication("hetzplay", appVer))
	return
}
