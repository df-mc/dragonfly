package session

import (
	"strconv"
	"strings"

	"github.com/df-mc/dragonfly/server/player/ddui"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type ServerBoundDataStoreHandler struct {
	h *DDUIFormHandler
}

func (d *ServerBoundDataStoreHandler) Handle(p packet.Packet, s *Session, _ *world.Tx, _ Controllable) error {
	pk := p.(*packet.ServerBoundDataStore)

	property := pk.Update.Property
	if !strings.HasPrefix(property, "custom_form_data_") && !strings.HasPrefix(property, "message_box_data_") {
		return nil
	}

	const sep = "_data_"
	i := strings.LastIndex(property, sep)
	if i < 0 {
		return nil
	}
	id, err := strconv.ParseUint(property[i+len(sep):], 10, 32)
	if err != nil {
		return nil
	}

	d.h.mu.Lock()
	af := d.h.forms[uint32(id)]
	d.h.mu.Unlock()

	if af == nil {
		return nil
	}

	if af.form.HandleUpdate(pk.Update.Path, dataStoreControlToUpdateValue(pk.Update)) {
		d.h.mu.Lock()
		delete(d.h.forms, af.instanceID)
		d.h.mu.Unlock()
		af.form.OnClose(ddui.Closed)
		s.writePacket(&packet.ClientBoundDataDrivenUICloseScreen{
			FormID: protocol.Option(af.formID),
		})
		sendDataStoreCleanup(s, af)
	}
	return nil
}

func dataStoreControlToUpdateValue(u protocol.DataStoreUpdate) ddui.UpdateValue {
	switch u.ControlType {
	case protocol.DataStoreControlDouble:
		return ddui.UpdateValue{Kind: ddui.UpdateKindFloat, Float: u.DoubleValue}
	case protocol.DataStoreControlBoolean:
		return ddui.UpdateValue{Kind: ddui.UpdateKindBool, Bool: u.BoolValue}
	default:
		return ddui.UpdateValue{Kind: ddui.UpdateKindString, String: u.StringValue}
	}
}
