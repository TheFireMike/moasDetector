# MOAS Detector
[![Go Report Card](https://goreportcard.com/badge/github.com/thefiremike/moasdetector)](https://goreportcard.com/report/github.com/thefiremike/moasdetector)
[![GitHub license](https://img.shields.io/badge/license-BSD-blue.svg)](https://github.com/thefiremike/moasdetector/blob/main/LICENSE)
[![GoDoc doc](https://img.shields.io/badge/godoc-reference-blue)](https://godoc.org/github.com/thefiremike/moasdetector)

A tool which detects multiple origin AS (MOAS) prefixes based on BGP routing information in the MRT format.

## Data Source

The tool processes BGP routing information encoded in the [MRT format](https://datatracker.ietf.org/doc/html/rfc6396).

Public routing information encoded in this format is provided for example by the [RIPE Routing Information Service (RIS)](https://www.ripe.net/analyse/internet-measurements/routing-information-service-ris/) or the [RouteViews Project](http://www.routeviews.org/).
They feature the routing data from a large number of ASes which peer with their route collectors distributed around the globe.

Both of these projects feature two different types of files:
* **Table Dumps**, which contain the full BGP routing table of the route collector at the time of the snapshot.
* **Updates**, which contain every BGP update received by the route collector in the respective time period.

**Note**: Currently only Table Dumps are supported!

The MOAS Detector can directly consume the compressed (gzip or bzip2) MRT files, so no decompression or parsing of the files is required beforehand.

## Installation

You can download the latest version of the MOAS Detector under the `Releases` tab or build it yourself:

```
$ git clone https://github.com/TheFireMike/moasDetector.git
...
$ cd moasDetector && go build
$ ./moasDetector -h
Usage of ./moasDetector:
  -dir string
    	input file directory (required)
  -ignore string
    	ignore files whose path matches this regex
  -max-cpus int
    	limit the number of used CPUs (default 0 => no limit)
  -output string
    	output directory (default ".")
  -peers string
    	peers to process announcements from (comma seperated list of ASNs) (default all)
```

## Usage

First, download and place all uncompressed MRT files which you want to use for the computation into one folder, e.g.:
```
$ tree mrt_files
mrt_files
├── rrc00
│   └── bview.gz
├── rrc01
│   └── bview.gz
└── rrc02
    └── bview.gz

3 directories, 3 files
```

You can then invoke the MOAS Detector and point it to this directory:

```
$ ./moasDetector -dir mrt_files
```

After processing is finished, you will find the detected MOAS prefixes in the output directory.
Next to the `moasIPv4.json` and `moasIPv6.json` file you will also find the `statistics.json` file which contains information about the processed data.
