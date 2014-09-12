package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/pressly/qmd"
)

var (
	flagSet = flag.NewFlagSet("qmd-worker", flag.ExitOnError)

	config = flagSet.String("config", "", "path to config file")

	// Basic Options
	name       = flagSet.String("name", "", "name of the worker, autogenerated if none given")
	throughput = flagSet.Int("throughput", 10, "max number of concurrent worker threads")
	topic      = flagSet.String("topic", "job", "topic the worker looks for, defaults to 'job'")
	scriptDir  = flagSet.String("script-dir", "", "path to directory of scripts")
	workingDir = flagSet.String("working-dir", "", "the temporary working directory of the worker")
	storeDir   = flagSet.String("store-dir", "", "location for worker to persist items")
	whitelist  = flagSet.String("whitelist", "", "path to whitelist file")
	keepTemp   = flagSet.Bool("keep-temp", false, "keep temporary files in working dir")

	// Queue Options
	hostNSQD     = flagSet.String("host-nsqd", "0.0.0.0:4150", "<addr>:<port> to local NSQD node")
	nsqdAddrs    = qmd.StringFlagArray{}
	lookupdAddrs = qmd.StringFlagArray{}

	// Logging Options
	logLevel    = flagSet.String("log-level", "INFO", "level of logging") // DEBUG > INFO > NOTICE > WARNING > ERROR > CRITICAL
	logBackends = qmd.StringFlagArray{}                                   // "STDOUT", "syslog", or "/file/path"
)

func init() {
	flagSet.Var(&nsqdAddrs, "nsqd-addresses", "nsqd address for consumption (may be given multiple times)")
	flagSet.Var(&lookupdAddrs, "lookupd-addresses", "lookupd address for consumption, takes precedence over nsqd (may be given multiple times)")
	flagSet.Var(&logBackends, "log-backends", "log output location, defaults to STDOUT (may be given multiple times)")
}

func main() {
	var err error

	flagSet.Parse(os.Args[1:])

	var wc qmd.WorkerConfig
	if *config != "" {
		// Use toml file settings
		if _, err := toml.DecodeFile(*config, &wc); err != nil {
			log.Fatalf("Couldn't parse config file at %s", *config)
		}
	} else {
		// Use flag settings
		wc.Name = *name
		wc.Throughput = *throughput
		wc.Topic = *topic
		wc.ScriptDir = *scriptDir
		wc.WorkingDir = *workingDir
		wc.StoreDir = *storeDir
		wc.Whitelist = *whitelist
		wc.KeepTemp = *keepTemp
		wc.Queue = &qmd.QueueConfig{
			HostNSQDAddr: *hostNSQD,
			NSQDAddrs:    nsqdAddrs,
			LookupdAddrs: lookupdAddrs,
		}
		wc.Logging = &qmd.LoggingConfig{
			LogLevel:    *logLevel,
			LogBackends: logBackends,
		}
	}
	if err = wc.Clean(); err != nil {
		log.Fatal(err.Error())
	}

	exitChan := make(chan int)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		exitChan <- 1
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	worker, err := qmd.NewWorker(&wc)
	if err != nil {
		log.Fatalf("Worker couldn't created: %s", err.Error())
	}

	err = worker.Run()
	if err != nil {
		log.Fatalf("Can't start worker: %s", err.Error())
	}
	<-exitChan
	worker.Exit()
}
