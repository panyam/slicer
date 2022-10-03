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
		if _, ok := clients[addr]; !ok {
			log.Println("Client already stopped: ", addr)
		} else {
			delete(clients, addr)
			clients[addr].Stop()
		}
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
		if _, ok := producers[addr]; !ok {
			log.Println("Producer already stopped: ", addr)
		} else {
			delete(producers, addr)
			producers[addr].Stop()
		}
	}

	showHelp := func() {
		log.Println("Usage: ")
		log.Println("ctrl start|stop")
		log.Println("prod start|stop host:addr")
		log.Println("client start|stop host:addr")
		log.Println("help")
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
			log.Println("Parts: ", parts)
			if len(parts) < 1 {
				continue
			} else if parts[0] == "help" {
				showHelp()
			} else if parts[0] == "exit" {
				os.Exit(0)
			} else if len(parts) < 2 {
				log.Println("This command needs at least 2 arguments")
				continue
			} else if parts[0] == "ctrl" {
				if parts[1] == "start" {
					startController()
				} else {
					stopController()
				}
			} else if len(parts) < 3 {
				log.Println("This command needs at least 3 arguments")
				continue
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
			} else {
				log.Println("Invalid command")
			}
		}
	}
}
