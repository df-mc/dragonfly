package cmd

import "sync"

// commands holds a list of registered commands indexed by their name.
var commands sync.Map

// Register registers a command with its name and all aliases that it has. Any command with the same name or
// aliases will be overwritten.
func Register(command Command) {
	commands.Store(command.name, command)
	for _, alias := range command.aliases {
		commands.Store(alias, command)
	}
}

// ByAlias looks up a command by an alias. If found, the command and true are returned. If not, the returned
// command is nil and the bool is false.
func ByAlias(alias string) (Command, bool) {
	command, ok := commands.Load(alias)
	if !ok {
		return Command{}, false
	}
	return command.(Command), ok
}

// Commands returns a map of all registered commands indexed by the alias they were registered with.
func Commands() map[string]Command {
	cmd := make(map[string]Command)
	commands.Range(func(key, value any) bool {
		cmd[key.(string)] = value.(Command)
		return true
	})
	return cmd
}
