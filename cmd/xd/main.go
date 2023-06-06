package main

import (
    "flag"
    "log"

    "github.com/forscht/xd"
)

// Define a command line flag for specifying the configuration file path.
var configPath = flag.String("config", "", "Path to the configuration file")
var command = flag.String("command", "", "Execute specific command from config")

func main() {
    flag.Parse()

    commands, err := xd.LoadConfig(*configPath, *command)
    if err != nil {
        log.Fatalf("Could not read configuration: %v", err)
    }
    if len(commands) > 0 {
        xd.Navigate(commands, "xd")
    }
}
