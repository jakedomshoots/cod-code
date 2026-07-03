package cli

import "fmt"

func parseDelimitedCommand(args []string, start int, flag string) ([]string, int, error) {
	command := []string{}
	for index := start; index < len(args); index++ {
		if args[index] == "--" {
			return command, index, nil
		}
		command = append(command, args[index])
	}
	if len(command) == 0 {
		return nil, start, fmt.Errorf("%s requires a command", flag)
	}
	return command, len(args), nil
}
