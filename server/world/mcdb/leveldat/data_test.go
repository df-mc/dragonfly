package leveldat

import "testing"

func TestRespawnBlocksExplodeSetting(t *testing.T) {
	d := &Data{RespawnBlocksExplode: true}
	settings := d.Settings()
	if !settings.RespawnBlocksExplode {
		t.Fatal("Settings().RespawnBlocksExplode = false, want true")
	}
	settings.RespawnBlocksExplode = false
	d.PutSettings(settings)
	if d.RespawnBlocksExplode {
		t.Fatal("PutSettings() left RespawnBlocksExplode enabled")
	}
}
