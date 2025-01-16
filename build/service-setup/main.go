// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Simple service that only works by printing a log message every few seconds.
package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/kardianos/service"
)

// Config is the runner app config structure.

var logger service.Logger

type program struct {
	exit    chan struct{}
	service service.Service
	root    string

	cmd *exec.Cmd
}

// const WIN_PATH = "%ProgramFiles%\\bkup"
// const LINUX_PATH = "/usr/local/bkup" //"/home/shruti/Project/BTechProj-dedup/client-background" //

func (p *program) Start(s service.Service) error {

	p.cmd = exec.Command(filepath.Join(p.root, "client-background"))
	p.cmd.Env = append(os.Environ(),
		"ROOT_PATH="+p.root,
	)
	go p.run()
	return nil
}
func (p *program) run() {
	logger.Info("Starting ")
	defer func() {
		if service.Interactive() {
			p.Stop(p.service)
		} else {
			p.service.Stop()
		}
	}()

	// if p.Stderr != "" {
	// 	f, err := os.OpenFile(p.Stderr, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	// 	if err != nil {
	// 		logger.Warningf("Failed to open std err %q: %v", p.Stderr, err)
	// 		return
	// 	}
	// 	defer f.Close()
	// 	p.cmd.Stderr = f
	// }
	// if p.Stdout != "" {
	// 	f, err := os.OpenFile(p.Stdout, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	// 	if err != nil {
	// 		logger.Warningf("Failed to open std out %q: %v", p.Stdout, err)
	// 		return
	// 	}
	// 	defer f.Close()
	// 	p.cmd.Stdout = f
	// }

	err := p.cmd.Run()
	if err != nil {
		logger.Warningf("Error running: %v", err)
	}

}
func (p *program) Stop(s service.Service) error {
	close(p.exit)
	logger.Info("Stopping ")
	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func main() {
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()

	svcConfig := &service.Config{
		Name:        "client-background",
		DisplayName: "client-background",
		Description: "Backup background process",
	}

	prg := &program{
		exit: make(chan struct{}),
	}

	if runtime.GOOS == "windows" {
		prg.root = filepath.Join(os.Getenv("ProgramFiles"), "backup")
	} else if runtime.GOOS == "linux" {
		prg.root = filepath.Join(os.Getenv("HOME"), ".backup")
		svcConfig.UserName = os.Getenv("USR")
		// fmt.Println(prg.root)
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	prg.service = s

	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
