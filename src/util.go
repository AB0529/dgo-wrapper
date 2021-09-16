package dgowrapper

import (
	"strings"
)

// SelectPrefix selects the prefix from list of prefixes from message content
func SelectPrefix(content string) (string, error) {
	// Get the prefix
	for _, prefix := range Bot.Prefixes {
		if strings.HasPrefix(strings.ToLower(content), prefix) {
			return prefix, nil
		}
	}

	return "", ErrNoPrefixFound
}

// FindCommandOrAlias finds the command from message content
func FindCommandOrAlias(prefix string, content string) (*Command, error) {
	msg := strings.Split(
		strings.Replace(strings.ToLower(content), strings.ToLower(prefix), "", 1), " ")

	cmd, ok := Commands[msg[0]]

	// Can't find command, try for alias
	if !ok {
		alias, ok2 := Aliases[msg[0]]
		// No command or alias found
		if !ok2 {
			return nil, ErrNoCommandOrAliasFound
		}
		return alias, nil
	}

	return cmd, nil
}
