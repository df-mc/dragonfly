package mcdb

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb/leveldat"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/opt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

// Logger is a logger implementation that may be passed to the Log field of Config. World will send errors and debug
// messages to this Logger when appropriate.
type Logger interface {
	Errorf(format string, a ...any)
	Debugf(format string, a ...any)
}

// Config holds the optional parameters of a DB.
type Config struct {
	// Log is the Logger that will be used to log errors and debug messages to.
	// If set to nil, a Logrus logger will be used.
	Log Logger
	// Compression specifies the compression to use for compressing new data in
	// the database. Decompression of the database will happen based on IDs
	// found in the compressed blocks and is therefore uninfluenced by this
	// field. If left empty, Compression will default to opt.FlateCompression.
	Compression opt.Compression
	// BlockSize specifies the size of blocks to be compressed. The default
	// value, when left empty, is 16KiB (16 * opt.KiB). Higher values generally
	// lead to better compression ratios at the expense of slightly higher
	// memory usage while (de)compressing.
	BlockSize int
	// ReadOnly opens the DB in read-only mode. This will leave the data in the
	// database unedited.
	ReadOnly bool

	// Entities is an EntityRegistry with all entity types registered that may
	// be read from the DB. Entities will default to entity.DefaultRegistry.
	Entities world.EntityRegistry
}

// Open creates a new DB reading and writing from/to files under the path
// passed. If a world is present at the path, Open will parse its data and
// initialise the world with it. If the data cannot be parsed, an error is
// returned.
func (conf Config) Open(dir string) (*DB, error) {
	if conf.Log == nil {
		conf.Log = logrus.New()
	}
	if conf.BlockSize == 0 {
		conf.BlockSize = 16 * opt.KiB
	}
	if len(conf.Entities.Types()) == 0 {
		conf.Entities = entity.DefaultRegistry
	}
	_ = os.MkdirAll(filepath.Join(dir, "db"), 0777)

	db := &DB{conf: conf, dir: dir, ldat: &leveldat.Data{}}
	if _, err := os.Stat(filepath.Join(dir, "level.dat")); os.IsNotExist(err) {
		// A level.dat was not currently present for the world.
		db.ldat.FillDefault()
	} else {
		ldat, err := leveldat.ReadFile(filepath.Join(dir, "level.dat"))
		if err != nil {
			return nil, fmt.Errorf("open db: %w", err)
		}

		// TODO: Perform proper conversion here. Dragonfly stored 3 for a long
		//  time even though the fields were up to date, so we have to accept
		//  older ones no matter what.
		ver := ldat.Ver()
		if ver != leveldat.Version && ver >= 10 {
			return nil, fmt.Errorf("open db: level.dat version %v is unsupported", ver)
		}
		if err = ldat.Unmarshal(db.ldat); err != nil {
			return nil, fmt.Errorf("open db: %w", err)
		}
	}
	db.set = db.ldat.Settings()
	ldb, err := leveldb.OpenFile(filepath.Join(dir, "db"), &opt.Options{
		Compression: conf.Compression,
		BlockSize:   conf.BlockSize,
		ReadOnly:    conf.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("error opening leveldb database: %w", err)
	}
	db.ldb = ldb
	return db, nil
}
