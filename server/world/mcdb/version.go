package mcdb

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"strconv"
	"strings"
)

// minimumCompatibleClientVersion is the minimum compatible client version, required by the latest Minecraft data provider.
var minimumCompatibleClientVersion []int32

// init initializes the minimum compatible client version.
func init() {
	fullVersion := append(strings.Split(protocol.CurrentVersion, "."), "0", "0")
	for _, v := range fullVersion {
		i, _ := strconv.Atoi(v)
		minimumCompatibleClientVersion = append(minimumCompatibleClientVersion, int32(i))
	}
}
