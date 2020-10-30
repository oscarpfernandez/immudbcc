package server

import (
	"log"
	"os"
	"path/filepath"

	immulogger "github.com/codenotary/immudb/pkg/logger"
	immuserver "github.com/codenotary/immudb/pkg/server"
)

type Config struct {
	AuthEnabled   bool
	LogFile       string
	serverOptions immuserver.Options
}

type Server struct {
	conf   Config
	server immuserver.ImmuServerIf
	logger immulogger.Logger
	file   *os.File
}

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

func (s *Server) Start() {
	go func() {
		if err := s.server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
}

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
