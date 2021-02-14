package server

// This package serves the purpose of facilitating the testing of the project
// by using an embedded Server version of ImmuDB. This makes the infrastructure
// perspective of the project much more manageable.

import (
	"log"
	"os"
	"path/filepath"
	"time"

	immulogger "github.com/codenotary/immudb/pkg/logger"
	immuserver "github.com/codenotary/immudb/pkg/server"
)

// Config represents the basic configuration parameters of the server.
type Config struct {
	AuthEnabled   bool
	LogFile       string
	serverOptions immuserver.Options
}

// Server represents the a server instance and its state.
type Server struct {
	conf   Config
	server immuserver.ImmuServerIf
	logger immulogger.Logger
	file   *os.File
}

// New creates a new Server instance.
func New(c Config) (*Server, error) {
	flogger, file, err := immulogger.NewFileLogger("immuserver ", c.LogFile)
	if err != nil {
		return nil, err
	}

	c.serverOptions = immuserver.DefaultOptions().WithLogfile(c.LogFile).WithAuth(c.AuthEnabled)
	server := immuserver.DefaultServer().WithOptions(c.serverOptions).WithLogger(flogger)

	return &Server{
		server: server,
		logger: flogger,
		file:   file,
	}, nil
}

// Start launches a server instance in a go-routine.
func (s *Server) Start() {
	go func() {
		if err := s.server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	time.Sleep(50 * time.Millisecond)
}

// Stop immediately stops the server instance and cleans up.
func (s *Server) Stop() error {
	defer func() {
		s.cleanupServerFiles()
		s.file.Close()
	}()

	if err := s.server.Stop(); err != nil {
		return err
	}

	return nil
}

// cleanupServerFiles does some house keeping chores.
func (s *Server) cleanupServerFiles() {
	os.RemoveAll(s.conf.serverOptions.Dir)  // remove db
	os.Remove(s.conf.serverOptions.Logfile) // remove log file

	// remove root
	files, err := filepath.Glob("./\\.root*")
	if err == nil {
		for _, f := range files {
			os.Remove(f)
		}
	}
}
