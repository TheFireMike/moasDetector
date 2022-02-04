package routes

import (
	"encoding/json"
	"github.com/pkg/errors"
	"net"
	"os"
	"path/filepath"
)

type Routes struct {
	routesIPv4 routeData
	routesIPv6 routeData
	peers      []Peer
}

type routeData struct {
	prefixes map[string]map[string][]Peer
}

type MOASPrefix struct {
	Prefix string             `json:"prefix"`
	Origin []MOASPrefixOrigin `json:"origin"`
}

type MOASPrefixOrigin struct {
	AS         string `json:"as"`
	Visibility []Peer `json:"visibility"`
}

type Peer struct {
	AS string `json:"as"`
	IP string `json:"ip"`
}

type RouteAnnouncement struct {
	Prefix     net.IPNet
	OriginAS   string
	ReceivedBy Peer
}

type Channels struct {
	IPv4   chan RouteAnnouncement
	IPv6   chan RouteAnnouncement
	Peers  chan []Peer
	Errors chan error
}

func NewChannels() Channels {
	return Channels{
		IPv4:   make(chan RouteAnnouncement),
		IPv6:   make(chan RouteAnnouncement),
		Peers:  make(chan []Peer),
		Errors: make(chan error),
	}
}

func (c *Channels) Close() {
	close(c.IPv4)
	close(c.IPv6)
	close(c.Peers)
	close(c.Errors)
}

func NewRoutes() Routes {
	return Routes{
		routesIPv4: routeData{make(map[string]map[string][]Peer)},
		routesIPv6: routeData{make(map[string]map[string][]Peer)},
	}
}

func (r *Routes) HandleAnnouncements(channels Channels) error {
	go r.routesIPv4.handleAnnouncements(channels.IPv4)
	go r.routesIPv6.handleAnnouncements(channels.IPv6)
	go r.handlePeers(channels.Peers)

	for err := range channels.Errors {
		return errors.Wrap(err, "failed to handle route announcement")
	}
	return nil
}

func (r *Routes) handlePeers(peersChan chan []Peer) {
	for peers := range peersChan {
		r.peers = getUniquePeers(append(r.peers, peers...))
	}
}

func (r *routeData) handleAnnouncements(announcementChan chan RouteAnnouncement) {
	for announcement := range announcementChan {
		r.addRoute(announcement)
	}
}

func (r *routeData) addRoute(announcement RouteAnnouncement) {
	if res, ok := r.prefixes[announcement.Prefix.String()]; !ok {
		r.prefixes[announcement.Prefix.String()] = map[string][]Peer{
			announcement.OriginAS: {announcement.ReceivedBy},
		}
	} else {
		if feeders, ok := res[announcement.OriginAS]; !ok {
			r.prefixes[announcement.Prefix.String()][announcement.OriginAS] = []Peer{announcement.ReceivedBy}
		} else {
			r.prefixes[announcement.Prefix.String()][announcement.OriginAS] = append(feeders, announcement.ReceivedBy)
		}
	}
}

func (r *Routes) PrintMOAS(directory string) error {
	err := r.routesIPv4.printMOASToFile(directory, "moasIPv4.json")
	if err != nil {
		return errors.Wrap(err, "failed to print IPv4 MOAS file")
	}
	err = r.routesIPv6.printMOASToFile(directory, "moasIPv6.json")
	if err != nil {
		return errors.Wrap(err, "failed to print IPv6 MOAS file")
	}
	err = r.printPeers(directory, "peers.json")
	if err != nil {
		return errors.Wrap(err, "failed to print peers file")
	}
	return nil
}

func (r *Routes) printPeers(directory, filename string) error {
	d, err := json.Marshal(r.peers)
	if err != nil {
		return errors.Wrap(err, "failed to marshal peers to JSON")
	}
	err = os.WriteFile(filepath.Join(directory, filename), d, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write peers file")
	}
	return nil
}

func (r *routeData) printMOASToFile(directory, filename string) error {
	var moas []MOASPrefix

	for prefix, origins := range r.prefixes {
		if len(origins) > 1 {
			moasPrefix := MOASPrefix{
				Prefix: prefix,
			}
			for origin, peers := range origins {
				moasPrefix.Origin = append(moasPrefix.Origin, MOASPrefixOrigin{
					AS:         origin,
					Visibility: getUniquePeers(peers),
				})
			}
			moas = append(moas, moasPrefix)
		}
	}

	d, err := json.Marshal(moas)
	if err != nil {
		return errors.Wrap(err, "failed to marshal MOAS to JSON")
	}
	err = os.WriteFile(filepath.Join(directory, filename), d, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write MOAS file")
	}
	return nil
}

func getUniquePeers(peers []Peer) []Peer {
	var uniquePeers []Peer
	peersHashmap := make(map[Peer]struct{})
	for _, peer := range peers {
		if _, ok := peersHashmap[peer]; !ok {
			peersHashmap[peer] = struct{}{}
			uniquePeers = append(uniquePeers, peer)
		}
	}
	return uniquePeers
}
