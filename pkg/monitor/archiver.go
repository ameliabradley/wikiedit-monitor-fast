package monitor

import (
	"io/ioutil"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Archiver archives the given revision to a folder
type Archiver interface {
	Archive(revision int, diff []byte)
}

// fileArchive is an implementation of Archiver
type fileArchive struct {
	folder string
	logger *logrus.Logger
}

// NewFileArchiver creates a new instance of an Archiver
func NewFileArchiver(logger *logrus.Logger, folder string) Archiver {
	return fileArchive{
		folder: folder,
		logger: logger,
	}
}

// Archives archives the given revision to a folder
func (a fileArchive) Archive(revision int, diff []byte) {
	path := a.folder + "/" + strconv.Itoa(revision)
	a.logger.WithFields(logrus.Fields{
		"file": path,
	}).Info("Archiving revision")

	err := ioutil.WriteFile(path, diff, 0644)
	if err != nil {
		a.logger.WithError(err).Error("Could not write file")
	}
}
