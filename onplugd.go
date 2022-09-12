package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"onplugd/actionregistry"
	"onplugd/confmonitor"
	"onplugd/devicemonitor"
	"onplugd/engine"
	"onplugd/executor"
	"onplugd/messagepipe"
	"onplugd/utils"
)

// MainLoop is a type that describes a main loop: a function that starts the
// main loop of the program, and returns a stopper function that can stop the
// main loop.
type MainLoop func() (func() error, error)

// RunWithSignals wraps a main loop in a handler that can deal with SIGINT and
// SIGHUP signals.
func RunWithSignals(loop MainLoop) error {

	for done := false; !done; {
		stop, err := loop()
		if err != nil {
			return err
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT)

		s := <-sig
		switch s {
		case syscall.SIGHUP:
			log.Println("SIGHUP received, reloading...")
		case syscall.SIGINT:
			log.Println("SIGINT received, quitting...")
			done = true
		default:
			log.Println("Unexpected signal received:", s)
			done = true
		}

		signal.Stop(sig)
		err = stop()
		if err != nil {
			return err
		}
	}

	return nil
}

func mainLoop(configDir string, debug bool) (func() error, error) {

	messagePipe := messagepipe.New(debug)
	deviceMonitor := devicemonitor.New(&messagePipe)
	executor, cleanup := executor.New(&messagePipe)
	actionRegistry := actionregistry.New(&messagePipe, executor)
	confMonitor := confmonitor.New(configDir, &messagePipe)

	e := engine.New(&deviceMonitor, &confMonitor, actionRegistry, &messagePipe)
	e.AddCleanupCallback(cleanup)

	err := e.Start()
	if err != nil {
		return nil, err
	}

	return e.Stop, nil
}

func main() {

	configDirFlag := flag.String("config_dir", "~/.config/onplugd.d/",
		"The directory where configs are stored")
	debug := flag.Bool("debug", false, "Log more verbosely")
	flag.Parse()

	configDir := utils.Expand(*configDirFlag)

	if *debug {
		log.Println("Debug on.")
		log.Println("Config directory:", configDir)
	}

	log.Println("Started with PID", os.Getpid())

	err := RunWithSignals(func() (func() error, error) {
		return mainLoop(configDir, *debug)
	})
	if err != nil {
		log.Fatal(err)
	}

}
