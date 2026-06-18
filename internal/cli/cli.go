// Package cli implements mctop's headless subcommands. Each returns a process
// exit code; main dispatches to them.
package cli

import "time"

// dialTimeout bounds connecting to and querying a server, so a stuck server
// never hangs the command.
const dialTimeout = 30 * time.Second
