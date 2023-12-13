package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/y-yagi/configure"
	"github.com/y-yagi/goext/arr"
)

const (
	// VERSION is a version of this app
	VERSION = "0.0.1"
)

type config struct {
	Aliases map[string]string `toml:"aliases"`
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		return 1
	}
	return 0
}

func cmdAdd(alias string) error {
	var cfg config
	err := configure.Load("goto", &cfg)
	if err != nil {
		return err
	}

	var directory string

	// TODO: Check on the existence of alias.

	fmt.Print("Directory: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return errors.New("canceled")
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	directory = scanner.Text()

	if len(directory) == 0 {
		wd, err := os.Getwd()
		if err == nil {
			directory = wd
		}
	}

	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{alias: directory}
	} else {
		cfg.Aliases[alias] = directory
	}

	return configure.Save("goto", cfg)
}

func cmdEdit() error {
	editor := os.Getenv("EDITOR")
	if len(editor) == 0 {
		editor = "vim"
	}

	return configure.Edit("goto", editor)
}

func cmdDelete(alias string) error {
	var cfg config
	err := configure.Load("goto", &cfg)
	if err != nil {
		return err
	}

	delete(cfg.Aliases, alias)
	return configure.Save("goto", cfg)
}

func cmdShowAll() error {
	var cfg config
	err := configure.Load("goto", &cfg)
	if err != nil {
		return err
	}

	for key, value := range cfg.Aliases {
		fmt.Printf("%s: %s\n", key, value)
	}
	return nil
}

func cmdGoto(alias string) error {
	var cfg config
	err := configure.Load("goto", &cfg)
	if err != nil {
		return err
	}

	dir := cfg.Aliases[alias]
	if len(dir) != 0 {
		fmt.Fprintf(os.Stdout, "%s", dir)
		return nil
	}

	var maybe []string
	for key := range cfg.Aliases {
		if strings.HasPrefix(key, alias) {
			maybe = append(maybe, key)
		}
	}

	if len(maybe) == 1 {
		fmt.Fprintf(os.Stdout, "%s", cfg.Aliases[maybe[0]])
		return nil
	} else if len(maybe) > 1 {
		return fmt.Errorf("Did you mean '%s'?", arr.Join(maybe, ", "))
	}

	return fmt.Errorf("'%s' is not registered", alias)
}

func run() int {
	var showVersion bool
	var showAliases bool
	var editAliases bool
	var addAlias string
	var deleteAlias string

	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showAliases, "s", false, "show all aliases")
	flag.BoolVar(&editAliases, "c", false, "edit aliases")
	flag.StringVar(&addAlias, "a", "", "add alias")
	flag.StringVar(&deleteAlias, "d", "", "delete alias")
	flag.Parse()

	if showVersion {
		fmt.Println("version:", VERSION)
		return 0
	}

	if showAliases {
		return msg(cmdShowAll())
	}

	if editAliases {
		return msg(cmdEdit())
	}

	if len(addAlias) > 0 {
		return msg(cmdAdd(addAlias))
	}

	if len(deleteAlias) > 0 {
		return msg(cmdDelete(deleteAlias))
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Please specify alias.")
		return 0
	}

	return msg(cmdGoto(flag.Args()[0]))
}

func main() {
	os.Exit(run())
}
