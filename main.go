package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/sirnewton01/mdns"
)

var (
	searchVar bool
	bcastVar  bool
)

func init() {
	flag.BoolVar(&searchVar, "search", false, "Search for other 9P services")
	flag.BoolVar(&bcastVar, "bcast", false, "Broadcast 9P service at this address")
	flag.Parse()
}

func main() {
	if searchVar {
		search()
	} else if bcastVar {
		bcast()
	} else {
		fmt.Errorf("No mode specified, use either -search or -bcast\n")
		flag.PrintDefaults()
	}
}

func search() {
	log.SetOutput(ioutil.Discard)

	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			fmt.Printf("%v\n", entry.AddrV4)
		}
	}()

	// Start the lookup
	query := mdns.DefaultParams("_9p._tcp")
	query.Timeout = 1 * time.Second
	query.Entries = entriesCh
	err := mdns.Query(query)
	if err != nil {
		panic(err)
	}
	close(entriesCh)
}

func bcast() {
	host, _ := os.Hostname()
	info := []string{"9P"}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}

	ips := make([]net.IP, 0)

	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok {
			if ip.IP.IsLinkLocalUnicast() {
				ips = append(ips, ip.IP)
			}
		}
	}

	service, err := mdns.NewMDNSService(host, "_9p._tcp", "", "", 564, ips, info)
	if err != nil {
		panic(err)
	}
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		panic(err)
	}
	defer server.Shutdown()

	select {}
}
