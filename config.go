package xd

import (
    "fmt"
    "log"
    "os"
    "os/user"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v3"
)

const baseyaml = `
- name: System
  commands:
    - name: Reboot
      cmd: reboot
    - name: Shutdown
      cmd: shutdown now
    - name: Suspend
      cmd: systemctl suspend
`

// LoadConfig reads configuration files from a provided path or from the default location.
// It attempts to load configuration file in the following order:
// 1. From the path provided through the command line flag -config.
// 2. From the directory specified in XDG_CONFIG_HOME environment variable.
// If the directory does not exist or is empty, it creates a default configuration file.
// The function also takes an optional `command` parameter which allows selecting a specific
// sub-command directly from the configuration.
func LoadConfig(configPath, command string) ([]Command, error) {
    var commands []Command

    // Check if a config file path was provided through the flag
    if configPath != "" {
        return ReadConfig(configPath, command)
    }

    // Get the XDG_CONFIG_HOME directory, default to ~/.config if it's not set
    configDir := os.Getenv("XDG_CONFIG_HOME")
    if configDir == "" {
        usr, err := user.Current()
        if err != nil {
            return nil, fmt.Errorf("failed to get home directory. err: %q", err)
        }
        configDir = filepath.Join(usr.HomeDir, ".config")
    }

    configDir = filepath.Join(configDir, "xd")

    // If the config directory does not exist, create it and write a default config file
    if _, err := os.Stat(configDir); os.IsNotExist(err) {
        if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
            return nil, fmt.Errorf("failed to create config dir. err: %q", err)
        }
        if err := os.WriteFile(filepath.Join(configDir, "xd.yaml"), []byte(baseyaml), os.ModePerm); err != nil {
            return nil, fmt.Errorf("failed to write default config file err: %q", err)
        }
    }

    // Read all yaml files in the config directory and merge them into the commands
    files, err := os.ReadDir(configDir)
    if err != nil {
        return nil, fmt.Errorf("could not read config dir: %s err:%q", configDir, err)
    }

    for _, file := range files {
        if !file.IsDir() && filepath.Ext(file.Name()) == ".yaml" {
            fileCommands, err := ReadConfig(filepath.Join(configDir, file.Name()), command)
            if err != nil {
                log.Printf("Warning: could not read configuration file %s: %q", file.Name(), err)
                continue
            }
            commands = append(commands, fileCommands...)
        }
    }

    return commands, nil
}

// ReadConfig reads a configuration file and unmarshal it into a slice of Command structs.
// It takes a `command` parameter which, if provided, will filter and return only the sub-commands
// of the matching command. If `command` is an empty string, it will return all commands from the
// configuration file.
func ReadConfig(path string, command string) ([]Command, error) {
    var commands []Command
    config, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to load config. err: %q", err)
    }
    if err := yaml.Unmarshal(config, &commands); err != nil {
        return nil, fmt.Errorf("failed to parse config file. err:  %q", err)
    }
    if command != "" {
        for _, c := range commands {
            if strings.ToLower(c.Name) == strings.ToLower(command) && len(c.Commands) > 0 {
                return c.Commands, nil
            }
        }
        return nil, nil
    }
    return commands, nil
}
