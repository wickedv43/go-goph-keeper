// Package kv provides a RoseDB-backed key-value storage used by the client.
package kv

import (
	"github.com/pkg/errors"
	"github.com/rosedblabs/rosedb/v2"
	"github.com/samber/do/v2"
	"github.com/wickedv43/go-goph-keeper/internal/config"
	"github.com/wickedv43/go-goph-keeper/internal/logger"
	"go.uber.org/zap"
)

// KV represents the RoseDB-based key-value storage layer with logging.
type KV struct {
	db *rosedb.DB

	log *zap.SugaredLogger
}

// NewRoseDB initializes and returns a KV instance using the configured RoseDB directory.
func NewRoseDB(i do.Injector) (*KV, error) {
	cfg := do.MustInvoke[*config.Config](i)
	log := do.MustInvoke[*logger.Logger](i)

	kv, err := do.InvokeStruct[KV](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	options := rosedb.DefaultOptions
	options.DirPath = cfg.KV.DirPath

	// open a database
	db, err := rosedb.Open(options)
	if err != nil {
		return nil, errors.Wrap(err, "init db")
	}

	kv.db = db
	kv.log = log.Named("kv")

	return kv, nil
}

// Shutdown closes the underlying RoseDB instance.
func (s *KV) Shutdown() error {
	//defer func() {
	//	_ = os.RemoveAll("/tmp/rosedb_basic")
	//}()
	return s.db.Close()
}
