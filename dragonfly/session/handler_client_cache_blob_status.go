package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ClientCacheBlobStatusHandler handles the ClientCacheBlobStatus packet.
type ClientCacheBlobStatusHandler struct {
}

// Handle ...
func (c *ClientCacheBlobStatusHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.ClientCacheBlobStatus)

	resp := &packet.ClientCacheMissResponse{Blobs: make([]protocol.CacheBlob, 0, len(pk.MissHashes))}

	s.blobMu.Lock()
	for _, hit := range pk.HitHashes {
		delete(s.blobs, hit)
	}
	for _, miss := range pk.MissHashes {
		blob, ok := s.blobs[miss]
		if !ok {
			s.log.Debugf("missing blob hash could not be recovered: %v", miss)
			continue
		}
		resp.Blobs = append(resp.Blobs, protocol.CacheBlob{Hash: miss, Payload: blob})
	}
	s.blobMu.Unlock()

	if len(resp.Blobs) > 0 {
		s.writePacket(resp)
	}
	return nil
}
