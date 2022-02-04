package main

import (
	"flag"
	"github.com/TheFireMike/moasDetector/parser"
	"github.com/TheFireMike/moasDetector/routes"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

var dir = flag.String("dir", "", "input file directory")
var output = flag.String("output", ".", "output directory")

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	flag.Parse()
	if dir == nil || *dir == "" {
		log.Fatal().Msg("flag 'dir' is missing")
	}
	if output == nil || *output == "" {
		log.Fatal().Msg("flag 'output' is missing")
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
}

func main() {
	channels := routes.NewChannels()

	go parser.ProcessFiles(*dir, channels)

	r := routes.NewRoutes()
	err := r.HandleAnnouncements(channels)
	if err != nil {
		log.Fatal().Err(err).Msg("handling route announcements failed")
	}

	err = r.PrintMOAS(*output)
	if err != nil {
		log.Fatal().Err(err).Msg("printing moas failed")
	}
}
