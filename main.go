package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	// VERSION is a version of this app
	VERSION = "0.0.1"
)

type config struct {
	Aliases map[string]string `toml:"aliases"`
}

func configDir() string {
	var dir string

	if runtime.GOOS == "windows" {
		dir = os.Getenv("APPDATA")
		if dir == "" {
			dir = filepath.Join(os.Getenv("USERPROFILE"), "Application Data", "goto")
		}
		dir = filepath.Join(dir, "goto")
	} else {
		dir = filepath.Join(os.Getenv("HOME"), ".config", "goto")
	}
	return dir
}

func msg(err error) int {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		return 1
	}
	return 0
}

func (cfg *config) load() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("cannot create directory: %v", err)
	}
	file := filepath.Join(dir, "config.toml")

	_, err := os.Stat(file)
	if err == nil {
		_, err := toml.DecodeFile(file, cfg)
		if err != nil {
			return err
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (cfg *config) save() error {
	dir := configDir()
	file := filepath.Join(dir, "config.toml")

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	return toml.NewEncoder(f).Encode(cfg)
}

func cmdAdd(alias string) error {
	var cfg config
	err := cfg.load()
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
		directory, _ = os.Getwd()
	}

	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{alias: directory}
	} else {
		cfg.Aliases[alias] = directory
	}
	return cfg.save()
}

func cmdDelete(alias string) error {
	var cfg config
	err := cfg.load()
	if err != nil {
		return err
	}

	delete(cfg.Aliases, alias)
	return cfg.save()
}

func cmdShowAll() error {
	var cfg config
	err := cfg.load()
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
	err := cfg.load()
	if err != nil {
		return err
	}

	directory := detectDirectory(alias, cfg.Aliases)
	if len(directory) == 0 {
		return fmt.Errorf("'%s' is not registered", alias)
	}

	fmt.Fprintf(os.Stdout, "%s", directory)
	return nil
}

func detectDirectory(alias string, aliases map[string]string) string {
	directory := aliases[alias]

	if len(directory) != 0 {
		return directory
	}

	for k, v := range aliases {
		if strings.HasPrefix(k, alias) {
			directory = v
			break
		}
	}

	return directory
}

func run() int {
	var showVersion bool
	var showAliases bool
	var addAlias string
	var deleteAlias string

	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.BoolVar(&showAliases, "s", false, "show all aliases")
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

	return 0
}

func main() {
	os.Exit(run())
}
