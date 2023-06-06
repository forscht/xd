package xd

import (
    "flag"
    "log"
    "os"
    "os/exec"
    "strings"
)

// Command is a struct that defines the structure of command configurations.
type Command struct {
    // Name represents the option that the user will see in the dmenu.
    Name string `yaml:"name"`
    // Cmd defines the actual command that will be executed when the user selects this option.
    Cmd string `yaml:"cmd"`

    // List represents a command that will be executed to generate dynamic options.
    // These dynamic options will be shown to the user and the selected option will replace "$selected"
    // in the Cmd or in the Cmd of nested Commands. This is useful for listing available resources (like Bluetooth connections)
    // and taking actions based on the user's selection.
    List string `yaml:"list"`

    // Prompt is used to ask the user for input without showing a list of options.
    // The user's input replaces "$selected" in Cmd or in the Cmd of nested Commands.
    Prompt string `yaml:"prompt"`

    // Commands hold any sub-commands nested under this command.
    // These commands will be presented as additional options to the user when the parent command is selected.
    Commands []Command `yaml:"commands"`
}

// Show function displays a dmenu with the given prompt and options, and waits for user input.
func Show(prompt string, options []string) string {
    cmd := exec.Command("dmenu", flag.Args()...)
    if len(options) > 0 {
        cmd.Stdin = strings.NewReader(strings.Join(options, "\n"))
    }
    cmd.Args = append(cmd.Args, "-i", "-p", prompt)
    output, err := cmd.Output()
    if err != nil {
        log.Fatalf("dmenu show error: %q prompt: %s options: %v", err, prompt, options)
    }
    return strings.TrimSpace(string(output))
}

// Execute function runs a bash command with the given command string and returns its output as a string.
// If there's an error running the command, it displays an error message and an exit option using Show.
func Execute(cmdStr string) string {
    cmd := exec.Command("bash", "-c", cmdStr)
    log.Printf("executing command -> %s", cmdStr)
    output, err := cmd.CombinedOutput()
    if err != nil {
        e := string(output)
        if e == "" {
            e = err.Error()
        }
        Show("Error", []string{"Error: " + e, "Exit"})
        os.Exit(1)
    }
    return string(output)
}

// Navigate function allows the user to navigate through a slice of commands using dmenu.
// For each command, it checks whether it has a List, nested Commands, or a Cmd.
// If a command has a List, it calls NavigateList to let the user select an item from the list.
// If it has nested Commands, it calls Navigate recursively with the nested commands.
// If it has a Cmd, it executes the command using Execute.
// The function always includes an "Exit" option in the dmenu for the user to exit the program.
func Navigate(commands []Command, prompt string) {
    // Prepare dmenu options
    names := make([]string, len(commands)+1)
    cmdMap := make(map[string]Command)
    for i, cmd := range commands {
        names[i] = cmd.Name
        cmdMap[cmd.Name] = cmd
    }
    // Add an "Exit" option
    names[len(commands)] = "Exit"

    // Show dmenu and get selected cmd
    selected := Show(prompt, names)
    selectedCmd := cmdMap[selected]

    // Navigate based on selected cmd
    if selectedCmd.Prompt != "" {
        NavigatePrompt(selectedCmd, prompt)
    } else if selectedCmd.List != "" {
        NavigateList(selectedCmd, prompt)
    } else if len(selectedCmd.Commands) > 0 {
        Navigate(selectedCmd.Commands, prompt+" > "+selectedCmd.Name)
    } else if selectedCmd.Cmd != "" {
        Execute(selectedCmd.Cmd)
    } else {
        return // No command or nested commands found
    }
}

func NavigatePrompt(cmd Command, prompt string) {
    input := Show(prompt+" > "+cmd.Name+" > "+cmd.Prompt, []string{})
    if input == "" {
        return
    }

    if len(cmd.Commands) > 0 {
        // Navigate deeper if there are nested commands
        for i := range cmd.Commands {
            cmd.Commands[i].List = strings.ReplaceAll(cmd.Commands[i].Cmd, "$selected", input)
            cmd.Commands[i].Cmd = strings.ReplaceAll(cmd.Commands[i].Cmd, "$selected", input)
        }
        Navigate(cmd.Commands, prompt+" > "+cmd.Name)
    } else if cmd.Cmd != "" {
        // Execute the command if there is one
        execCommand := strings.ReplaceAll(cmd.Cmd, "$selected", input)
        Execute(execCommand)
    } else {
        return // No command or nested commands found
    }
}

// NavigateList function is used to handle commands that have a List command.
// It first executes the List command to get a list of items, and shows them to the user using dmenu.
// The user can then select an item from the list.
// If the command has nested Commands, it replaces the placeholder "$selected" in the commands with the selected item,
// and calls Navigate with the nested commands.
// If the command has a Cmd, it replaces the placeholder "$selected" in the command with the selected item,
// and executes the command using Execute.
func NavigateList(cmd Command, prompt string) {

    // Prepare command options based on output of list command
    output := Execute(cmd.List)
    items := make([]string, 0)
    if output != "" {
        items = append(items, strings.Split(output, "\n")...)
    }
    items = append(items, "Exit")

    // Show dmenu and get selected item.
    selectedItem := Show(prompt+" > "+cmd.Name, items)
    if selectedItem == "Exit" {
        return
    }

    if len(cmd.Commands) > 0 {
        // Navigate deeper if there are nested commands
        for i := range cmd.Commands {
            cmd.Commands[i].List = strings.ReplaceAll(cmd.Commands[i].Cmd, "$selected", selectedItem)
            cmd.Commands[i].Cmd = strings.ReplaceAll(cmd.Commands[i].Cmd, "$selected", selectedItem)
        }
        Navigate(cmd.Commands, prompt+" > "+cmd.Name)
    } else if cmd.Cmd != "" {
        // Execute the command if there is one
        execCommand := strings.ReplaceAll(cmd.Cmd, "$selected", selectedItem)
        Execute(execCommand)
    } else {
        return // No command or nested commands found
    }
}
