package main

import (
	"bufio"
	"flag"
	"fmt"
	gut "github.com/panyam/goutils/utils"
	"github.com/panyam/slicer/cmd/harness"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	log_file     = flag.String("log_file", "/tmp/harness.log", "Logfile to redirect all logs to.")
	control_addr = flag.String("control_addr", ":7000", "Address of control service.")
	db_endpoint  = flag.String("db_endpoint", "postgres://postgres:docker@localhost:5432/slicerdb", "Endpoint of DB backing slicer shard targets.  Supported - sqlite eg (sqlite://~/.slicer/sqlite.db) or postgres eg (postgres://user:pass@localhost:5432/dbname)")
)

func main() {
	flag.Parse()

	logfile, err := os.OpenFile(*log_file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logfile.Close()

	var ctrl *harness.Controller
	clients := make(map[string]*harness.WebClient)
	producers := make(map[string]*harness.Producer)

	reader := bufio.NewReader(os.Stdin)
	spacesRe := regexp.MustCompile("\\s+")

	startController := func() {
		if ctrl != nil {
			log.Println("Controller already running.")
		} else {
			ctrl = harness.NewController(*control_addr, *db_endpoint, logfile)
			go ctrl.Start()
		}
	}
	stopController := func() {
		if ctrl != nil {
			ctrl.Stop()
			ctrl = nil
		} else {
			log.Println("Controller already stopped.")
		}
	}

	startClient := func(addr string) {
		if _, ok := clients[addr]; ok {
			log.Println("Client already running: ", addr)
		} else {
			clients[addr] = harness.NewWebClient(addr, logfile)
			go clients[addr].Start()
		}
	}
	stopClient := func(addr string) {
	}

	startProducer := func(addr string, prefix string) {
		if _, ok := producers[addr]; ok {
			log.Println("Producer already running: ", addr)
		} else {
			producers[addr] = harness.NewProducer(prefix, addr, *control_addr, logfile)
			go producers[addr].Start()
		}
	}
	stopProducer := func(addr string) {
	}

	for {
		fmt.Print(":-> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Read Error: ", err)
		} else {
			parts := spacesRe.Split(line, -1)
			parts = gut.Map(parts, func(s string) string { return strings.Trim(s, " ") })
			parts = gut.Filter(parts, func(s string) bool { return len(s) > 0 })
			if len(parts) < 2 {
				continue
			}
			log.Println("Parts: ", parts)
			if parts[0] == "ctrl" {
				if parts[1] == "start" {
					startController()
				} else {
					stopController()
				}
			} else if parts[0] == "client" {
				if parts[1] == "start" {
					startClient(parts[2])
				} else {
					stopClient(parts[2])
				}
			} else if parts[0] == "prod" {
				if parts[1] == "start" {
					startProducer(parts[2], parts[3])
				} else {
					stopProducer(parts[2])
				}
				/*
					} else if parts[0] == "get" {
						addr := parts[1]
						shard := parts[2]
						key := parts[3]
						client, ok := clients[addr]
				*/
			}
		}
	}
}
