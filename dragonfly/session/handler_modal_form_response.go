package session

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/player/form"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"go.uber.org/atomic"
	"sync"
)

// ModalFormResponseHandler handles the ModalFormResponse packet.
type ModalFormResponseHandler struct {
	mu        sync.Mutex
	forms     map[uint32]form.Form
	currentID atomic.Uint32
}

// nullBytes contains the word 'null' converted to a byte slice.
var nullBytes = []byte("null")

// Handle ...
func (h *ModalFormResponseHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.ModalFormResponse)

	h.mu.Lock()
	f, ok := h.forms[pk.FormID]
	delete(h.forms, pk.FormID)
	h.mu.Unlock()

	if !ok {
		return fmt.Errorf("no form with ID %v currently opened", pk.FormID)
	}
	if bytes.Equal(pk.ResponseData, nullBytes) {
		// The form was cancelled: The cross in the top right corner was clicked.
		return nil
	}
	if err := f.SubmitJSON(pk.ResponseData, s.c); err != nil {
		return fmt.Errorf("error submitting form data: %w", err)
	}
	return nil
}
