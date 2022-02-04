package parser

import (
	"compress/bzip2"
	"compress/gzip"
	"github.com/TheFireMike/go-mrt"
	"github.com/TheFireMike/moasDetector/routes"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

type mrtFile struct {
	name     string
	logger   zerolog.Logger
	channels routes.Channels
	peers    []routes.Peer
}

func ProcessFiles(directory string, channels routes.Channels) {
	defer channels.Close()

	wg := sync.WaitGroup{}
	err := filepath.WalkDir(directory, func(s string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			wg.Add(1)
			f := mrtFile{
				name:     s,
				logger:   log.With().Str("file", s).Logger(),
				channels: channels,
			}
			go f.process(&wg)
		}
		return nil
	})
	if err != nil {
		channels.Errors <- err
		return
	}

	wg.Wait()
}

func (f *mrtFile) process(wg *sync.WaitGroup) {
	defer wg.Done()

	fp, err := os.Open(f.name)
	if err != nil {
		f.channels.Errors <- err
		return
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			f.logger.Error().Err(err).Msg("closing file failed")
		}
	}(fp)

	var content io.Reader
	switch fileType := filepath.Ext(fp.Name()); fileType {
	case ".gz":
		content, err = gzip.NewReader(fp)
		if err != nil {
			f.channels.Errors <- err
			return
		}
	case ".bz2":
		content = bzip2.NewReader(fp)
	default:
		f.logger.Error().Msg("unknown file type found: " + fileType)
		return
	}

	mrtReader := mrt.NewReader(content)
	for {
		rec, err := mrtReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			f.logger.Error().Err(err).Msg("reading MRT entry failed")
			continue
		}

		switch rec.Type() {
		case mrt.TYPE_TABLE_DUMP_V2:
			switch rec.Subtype() {
			case mrt.TABLE_DUMP_V2_SUBTYPE_PEER_INDEX_TABLE:
				f.addPeerInformation(rec.(*mrt.TableDumpV2PeerIndexTable))
			case mrt.TABLE_DUMP_V2_SUBTYPE_RIB_IPv4_UNICAST,
				mrt.TABLE_DUMP_V2_SUBTYPE_RIB_IPv6_UNICAST,
				mrt.TABLE_DUMP_V2_SUBTYPE_RIB_IPv4_UNICAST_ADDPATH,
				mrt.TABLE_DUMP_V2_SUBTYPE_RIB_IPv6_UNICAST_ADDPATH:
				f.processMRTEntry(rec.(*mrt.TableDumpV2RIB))
			default:
				f.logger.Trace().Msgf("unknown mrt entry subtype: '%v'", rec.Subtype())
			}
		default:
			f.logger.Trace().Msgf("unknown mrt entry type: '%v'", rec.Type())
		}
	}
}

func (f *mrtFile) addPeerInformation(peers *mrt.TableDumpV2PeerIndexTable) {
	for _, peer := range peers.PeerEntries {
		f.peers = append(f.peers, routes.Peer{
			AS: peer.PeerAS.String(),
			IP: peer.PeerIPAddress.String(),
		})
	}
	f.channels.Peers <- f.peers
}

func (f *mrtFile) processMRTEntry(mrtEntry *mrt.TableDumpV2RIB) {
	if mrtEntry.Prefix != nil {
		err := filterPrefix(*mrtEntry.Prefix)
		if err != nil {
			f.logger.Trace().Err(err).Str("prefix", mrtEntry.Prefix.String()).Msg("invalid prefix")
			return
		}

		for _, ribEntry := range mrtEntry.RIBEntries {
			f.processRIBEntry(ribEntry, *mrtEntry.Prefix)
		}
	}
}

func (f *mrtFile) processRIBEntry(ribEntry *mrt.TableDumpV2RIBEntry, prefix net.IPNet) {
	for _, attribute := range ribEntry.BGPAttributes {
		switch asPath := attribute.Value.(type) {
		case mrt.BGPPathAttributeASPath:
			if len(asPath) == 0 {
				f.logger.Trace().Str("prefix", prefix.String()).Msg("AS path is empty")
				return
			}
			lastASPathEntry := asPath[len(asPath)-1]
			var originAS string

			switch lastASPathEntry.Type {
			case mrt.BGPASPathSegmentTypeASSequence:
				if len(lastASPathEntry.Value) == 0 {
					f.logger.Trace().Str("prefix", prefix.String()).Msg("last AS path entry is empty")
					return
				}
				originAS = lastASPathEntry.Value[len(lastASPathEntry.Value)-1].String()
				asnParsed, err := strconv.Atoi(originAS)
				if err != nil {
					f.logger.Trace().Err(err).Str("prefix", prefix.String()).Str("asn", originAS).Msg("ASN is not a number")
					return
				}
				err = filterASN(asnParsed)
				if err != nil {
					f.logger.Trace().Err(err).Str("prefix", prefix.String()).Str("asn", originAS).Msg("invalid ASN")
					return
				}
			case mrt.BGPASPathSegmentTypeASSet:
				var validASes []int
				for _, asn := range lastASPathEntry.Value {
					asnParsed, err := strconv.Atoi(asn.String())
					if err != nil {
						f.logger.Trace().Err(err).Str("prefix", prefix.String()).Str("asn", asn.String()).Msg("ASN is not a number")
						return
					}
					err = filterASN(asnParsed)
					if err != nil {
						f.logger.Trace().Err(err).Str("prefix", prefix.String()).Str("asn", asn.String()).Msg("invalid ASN")
					} else {
						validASes = append(validASes, asnParsed)
					}
				}

				if len(validASes) == 0 {
					originAS = "{"
					for index, asn := range lastASPathEntry.Value {
						originAS += asn.String()
						if index != len(lastASPathEntry.Value)-1 {
							originAS += ","
						}
					}
					originAS += "}"
					f.logger.Trace().Str("prefix", prefix.String()).Str("as_set", originAS).Msg("invalid AS set")
					return
				} else if len(validASes) == 1 {
					originAS = strconv.Itoa(validASes[0])
				} else {
					sort.Ints(validASes)
					originAS = "{"
					for index, asn := range validASes {
						originAS += strconv.Itoa(asn)
						if index != len(validASes)-1 {
							originAS += ","
						}
					}
					originAS += "}"
				}
			}

			if prefix.IP.To4() != nil {
				f.channels.IPv4 <- routes.RouteAnnouncement{
					Prefix:     prefix,
					OriginAS:   originAS,
					ReceivedBy: f.peers[ribEntry.PeerIndex],
				}
			} else {
				f.channels.IPv6 <- routes.RouteAnnouncement{
					Prefix:     prefix,
					OriginAS:   originAS,
					ReceivedBy: f.peers[ribEntry.PeerIndex],
				}
			}
			return
		}
	}
}
