package main

import (
	"flag"
	"github.com/TheFireMike/moasDetector/parser"
	"github.com/TheFireMike/moasDetector/routes"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var dir = flag.String("dir", "", "input file directory (required)")
var output = flag.String("output", ".", "output directory")
var peers = flag.String("peers", "", "peers to process announcements from (comma separated list of ASNs) (default all)")
var maxCPUs = flag.Int("max-cpus", 0, "limit the number of used CPUs (default 0 => no limit)")
var ignore = flag.String("ignore", "", "ignore files whose path matches this regex")

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	flag.Parse()
	if dir == nil || *dir == "" {
		log.Fatal().Msg("flag 'dir' is missing")
	}

	err := os.MkdirAll(*output, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create directory")
	}
	logfile, err := os.OpenFile(filepath.Join(*output, "log.txt"), os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open logfile")
	}

	log.Logger = zerolog.New(zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr}, logfile)).With().Timestamp().Logger()

	if *maxCPUs != 0 {
		runtime.GOMAXPROCS(*maxCPUs)
	}
}

func main() {
	channels := routes.NewChannels()

	var p []string
	if *peers != "" {
		p = strings.Split(*peers, ",")
	}

	var i *string
	if *ignore != "" {
		i = ignore
	}

	go parser.ProcessFiles(*dir, channels, p, i)

	r := routes.NewRoutes()
	err := r.HandleAnnouncements(channels)
	if err != nil {
		log.Fatal().Err(err).Msg("handling route announcements failed")
	}

	err = r.PrintMOASPrefixes(*output)
	if err != nil {
		log.Fatal().Err(err).Msg("printing moas failed")
	}
}
