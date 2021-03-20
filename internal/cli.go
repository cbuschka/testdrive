package internal

import (
	"github.com/jpillora/opts"
	"os"
)

type CommandConfig struct {
	Verbose bool `opts:"help=Turn on verbose logging."`
}

func Run() (int, error) {

	args := os.Args

	config := CommandConfig{}
	err := opts.New(&config).ParseArgs(args).Run()
	if err != nil {
		return -1, err
	} else {
		return 0, nil
	}
}

func (config *CommandConfig) Run() error {
	session, err := NewSession()
	if err != nil {
		return err
	}

	setVerbose(config.Verbose)

	err = session.LoadConfig("testdrive.yaml")
	if err != nil {
		return err
	}

	return session.Run()
}
