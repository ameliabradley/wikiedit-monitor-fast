package monitor

import (
	"io/ioutil"
	"strconv"

	"github.com/sirupsen/logrus"
)

type Archiver interface {
	Archive(revision int, diff []byte)
}

type FileArchive struct {
	folder string
	logger *logrus.Logger
}

func NewFileArchiver(logger *logrus.Logger, folder string) Archiver {
	return FileArchive{
		folder: folder,
		logger: logger,
	}
}

func (a FileArchive) Archive(revision int, diff []byte) {
	path := a.folder + "/" + strconv.Itoa(revision)
	a.logger.WithFields(logrus.Fields{
		"file": path,
	}).Info("Archiving revision")
	ioutil.WriteFile(path, diff, 0644)
}
