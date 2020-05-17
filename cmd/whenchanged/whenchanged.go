// Program whenchanged executes a command when a given file changes.
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var (
	ignoreErrors = flag.Bool("ignore_errors", true, "If false, exits when the command fails.")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var filesToWatch, cmd []string
	i := 0
	var args = flag.Args()
	for ; i < len(args); i++ {
		if args[i] == "--" {
			i++
			break
		}
		filesToWatch = append(filesToWatch, args[i])
	}
	for ; i < len(args); i++ {
		cmd = append(cmd, args[i])
	}

	execCmd := func() error {
		log.Printf("executing %s\n", strings.Join(cmd, " "))
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdin = os.Stdin
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		return c.Run()
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	for _, f := range filesToWatch {
		watcher.Add(f)
	}
	if err := execCmd(); err != nil && !*ignoreErrors {
		return err
	}
	for {
		select {
		case <-watcher.Events:
			if err := execCmd(); err != nil && !*ignoreErrors {
				return err
			}
		case err := <-watcher.Errors:
			return err
		}
	}
}
