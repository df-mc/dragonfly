// This program generates the items and blocks list in the StartGame packet. It does this by connecting to a
// partnered server using a gophertunnel client and reading the items and blocks sent by that server, then
// encodes them as JSON.
// Note that an XBOX Live password and email must be set in the constants for this to work.
package main

import (
	"encoding/json"
	"github.com/sandertv/gophertunnel/minecraft"
	"io/ioutil"
	"log"
)

const (
	email    = ""
	password = ""
)

func main() {
	conn, err := minecraft.Dialer{
		Email:    email,
		Password: password,
	}.Dial("raknet", "mco.mineplex.com:19132")
	if err != nil {
		log.Fatalln("Failed connecting to server:", err)
	}

	data, _ := json.Marshal(conn.GameData().Blocks)
	if err := ioutil.WriteFile("blocks.json", data, 0644); err != nil {
		log.Fatalln("Error writing blocks to file:", err)
	}
	log.Println("Blocks written to blocks.json")

	_ = conn.Close()
}
