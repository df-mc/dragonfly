package provider

import (
	"encoding/json"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/goleveldb/leveldb"
	"github.com/google/uuid"
	"os"
)

// DBProvider ...
type DBProvider struct {
	db *leveldb.DB
}

// NewDBProvider creates a new player data provider that saves and loads data using
// a LevelDB database.
func NewDBProvider(path string) (*DBProvider, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.Mkdir(path, 0777)
	}
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &DBProvider{db: db}, nil
}

// Save ...
func (p *DBProvider) Save(d player.Data) {
	data := toJson(d)
	jsondata, err := json.Marshal(data)
	if err != nil {
		return
	}
	_ = p.db.Put([]byte(data.UUID), jsondata, nil)
}

// Load ...
func (p *DBProvider) Load(UUID uuid.UUID) (player.Data, bool) {
	jsondata, err := p.db.Get([]byte(UUID.String()), nil)
	if err != nil {
		return player.Data{}, false
	}
	d := jsonData{}
	err = json.Unmarshal(jsondata, &d)
	if err != nil {
		return player.Data{}, false
	}

	return fromJson(d), true
}

// Close ...
func (p *DBProvider) Close() {
	_ = p.db.Close()
}
