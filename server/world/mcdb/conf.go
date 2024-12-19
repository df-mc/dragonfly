package mcdb

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/world/mcdb/leveldat"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/opt"
	"log/slog"
	"os"
	"path/filepath"
)

// Config holds the optional parameters of a DB.
type Config struct {
	// Log is the Logger that will be used to log errors and debug messages to.
	// If set to nil, Log is set to slog.Default().
	Log *slog.Logger
	// LDBOptions holds LevelDB specific default options, such as the block size
	// or compression used in the database.
	LDBOptions *opt.Options
}

// Open creates a new DB reading and writing from/to files under the path
// passed. If a world is present at the path, Open will parse its data and
// initialise the world with it. If the data cannot be parsed, an error is
// returned.
func (conf Config) Open(dir string) (*DB, error) {
	if conf.Log == nil {
		conf.Log = slog.Default()
	}
	conf.Log = conf.Log.With("provider", "mcdb")
	if conf.LDBOptions == nil {
		conf.LDBOptions = new(opt.Options)
	}
	if conf.LDBOptions.BlockSize == 0 {
		conf.LDBOptions.BlockSize = 16 * opt.KiB
	}
	_ = os.MkdirAll(filepath.Join(dir, "db"), 0777)

	db := &DB{conf: conf, dir: dir, ldat: &leveldat.Data{}}
	if _, err := os.Stat(filepath.Join(dir, "level.dat")); os.IsNotExist(err) {
		// A level.dat was not currently present for the world.
		db.ldat.FillDefault()
	} else {
		ldat, err := leveldat.ReadFile(filepath.Join(dir, "level.dat"))
		if err != nil {
			return nil, fmt.Errorf("open db: read level.dat: %w", err)
		}
		ver := ldat.Ver()
		if ver != leveldat.Version && ver >= 10 {
			return nil, fmt.Errorf("open db: level.dat version %v is unsupported", ver)
		}
		if err = ldat.Unmarshal(db.ldat); err != nil {
			return nil, fmt.Errorf("open db: unmarshal level.dat: %w", err)
		}
	}
	db.set = db.ldat.Settings()
	ldb, err := leveldb.OpenFile(filepath.Join(dir, "db"), conf.LDBOptions)
	if err != nil {
		return nil, fmt.Errorf("open db: leveldb: %w", err)
	}
	db.ldb = ldb
	return db, nil
}
