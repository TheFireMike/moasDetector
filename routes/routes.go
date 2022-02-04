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

type Statistics struct {
	IPv4Prefixes     int              `json:"ipv4_prefixes"`
	IPv6Prefixes     int              `json:"ipv6_prefixes"`
	IPv4MOASPrefixes int              `json:"ipv4_moas_prefixes"`
	IPv6MOASPrefixes int              `json:"ipv6_moas_prefixes"`
	Peers            []PeerStatistics `json:"peers"`
}

type PeerStatistics struct {
	Peer
	IPv4Prefixes     int `json:"ipv4_prefixes"`
	IPv6Prefixes     int `json:"ipv6_prefixes"`
	IPv4MOASPrefixes int `json:"ipv4_moas_prefixes"`
	IPv6MOASPrefixes int `json:"ipv6_moas_prefixes"`
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

func (r *Routes) PrintMOASPrefixes(directory string) error {
	moasIPv4 := r.routesIPv4.getMOASPrefixes()
	err := printJSON(moasIPv4, directory, "moasIPv4.json")
	if err != nil {
		return errors.Wrap(err, "failed to print IPv4 MOAS file")
	}

	moasIPv6 := r.routesIPv6.getMOASPrefixes()
	err = printJSON(moasIPv6, directory, "moasIPv6.json")
	if err != nil {
		return errors.Wrap(err, "failed to print IPv6 MOAS file")
	}

	err = printJSON(r.getStatistics(moasIPv4, moasIPv6), directory, "statistics.json")
	if err != nil {
		return errors.Wrap(err, "failed to print statistics file")
	}

	return nil
}

func (r *routeData) getMOASPrefixes() []MOASPrefix {
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

	return moas
}

func (r *Routes) getStatistics(moasIPv4, moasIPv6 []MOASPrefix) Statistics {
	statistics := Statistics{
		IPv4Prefixes:     len(r.routesIPv4.prefixes),
		IPv6Prefixes:     len(r.routesIPv6.prefixes),
		IPv4MOASPrefixes: len(moasIPv4),
		IPv6MOASPrefixes: len(moasIPv6),
	}

	peerStatistics := make(map[Peer]PeerStatistics)

	for prefix, origins := range r.routesIPv4.prefixes {
		var prefixPeers []Peer
		var prefixIsMOAS bool
		for _, moasPrefix := range moasIPv4 {
			if moasPrefix.Prefix == prefix {
				prefixIsMOAS = true
				break
			}
		}
		for _, receivedByPeers := range origins {
			prefixPeers = append(prefixPeers, receivedByPeers...)
		}
		for _, prefixPeer := range getUniquePeers(prefixPeers) {
			peerStatistic, ok := peerStatistics[prefixPeer]
			if !ok {
				peerStatistic.Peer = prefixPeer
			}
			peerStatistic.IPv4Prefixes += 1
			if prefixIsMOAS {
				peerStatistic.IPv4MOASPrefixes += 1
			}
			peerStatistics[prefixPeer] = peerStatistic
		}
	}

	for prefix, origins := range r.routesIPv6.prefixes {
		var prefixPeers []Peer
		var prefixIsMOAS bool
		for _, moasPrefix := range moasIPv6 {
			if moasPrefix.Prefix == prefix {
				prefixIsMOAS = true
				break
			}
		}
		for _, receivedByPeers := range origins {
			prefixPeers = append(prefixPeers, receivedByPeers...)
		}
		for _, prefixPeer := range getUniquePeers(prefixPeers) {
			peerStatistic, ok := peerStatistics[prefixPeer]
			if !ok {
				peerStatistic.Peer = prefixPeer
			}
			peerStatistic.IPv6Prefixes += 1
			if prefixIsMOAS {
				peerStatistic.IPv6MOASPrefixes += 1
			}
			peerStatistics[prefixPeer] = peerStatistic
		}
	}

	for _, peer := range r.peers {
		if peerStatistic, ok := peerStatistics[peer]; ok {
			statistics.Peers = append(statistics.Peers, peerStatistic)
		} else {
			statistics.Peers = append(statistics.Peers, PeerStatistics{
				Peer: peer,
			})
		}
	}

	return statistics
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

func printJSON(data interface{}, directory, filename string) error {
	d, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal data to JSON")
	}
	err = os.WriteFile(filepath.Join(directory, filename), d, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}
	return nil
}
