package provider

import (
	"encoding/json"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/goleveldb/leveldb"
	"os"
)

type DBProvider struct {
	db *leveldb.DB
}

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
	err = p.db.Put([]byte(d.XUID), jsondata, nil)
}

// Load ...
func (p *DBProvider) Load(XUID string) (player.Data, bool) {
	jsondata, err := p.db.Get([]byte(XUID), nil)
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
