package cli

import "strings"

// The standard flag package stops at the first positional argument. CLI users
// reasonably write both `build -o x .` and `build . -o x`, so normalize known
// value flags before parsing.
func normalizeValueFlags(args []string, valueFlags map[string]bool) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		if valueFlags[args[i]] && i+1 < len(args) {
			flags = append(flags, args[i], args[i+1])
			i++
			continue
		}
		matched := false
		for name := range valueFlags {
			if strings.HasPrefix(args[i], name+"=") {
				flags = append(flags, args[i])
				matched = true
				break
			}
		}
		if !matched {
			positional = append(positional, args[i])
		}
	}
	return append(flags, positional...)
}
