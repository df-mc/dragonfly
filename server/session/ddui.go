package session

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/player/ddui"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// DDUIFormHandler holds shared state for all active data-driven UI forms.
type DDUIFormHandler struct {
	mu             sync.Mutex
	forms          map[uint32]*activeDDUIForm
	nextFormID     atomic.Uint32
	nextInstanceID atomic.Uint32
}

// activeDDUIForm tracks a single open DDUI form.
type activeDDUIForm struct {
	form                ddui.Form
	formID              uint32
	instanceID          uint32
	property            string
	propertyUpdateCount uint32
}

// SendDDUIForm sends f to the client via s, registering it as an active form.
func (h *DDUIFormHandler) SendDDUIForm(f ddui.Form, s *Session) {
	instanceID := h.nextInstanceID.Add(1)
	formID := h.nextFormID.Add(1)

	property := deriveProperty(f.ScreenID(), instanceID)

	af := &activeDDUIForm{
		form:                f,
		formID:              formID,
		instanceID:          instanceID,
		property:            property,
		propertyUpdateCount: 1,
	}

	h.mu.Lock()
	h.forms[instanceID] = af
	h.mu.Unlock()

	f.BindSend(func(_ ddui.UpdateNotification) {
		h.mu.Lock()
		if _, active := h.forms[instanceID]; !active {
			h.mu.Unlock()
			return
		}
		af.propertyUpdateCount++
		count := af.propertyUpdateCount
		h.mu.Unlock()

		s.writePacket(&packet.ClientBoundDataStore{
			Updates: []protocol.DataStoreChangeEntry{
				{
					ChangeType: protocol.DataStoreChangeTypeChange,
					Change: protocol.DataStoreChange{
						DataStoreName: "minecraft",
						Property:      property,
						UpdateCount:   count,
						NewValue:      serializeForm(f.Describe()),
					},
				},
			},
		})
	})

	s.writePacket(&packet.ClientBoundDataStore{
		Updates: []protocol.DataStoreChangeEntry{
			{
				ChangeType: protocol.DataStoreChangeTypeChange,
				Change: protocol.DataStoreChange{
					DataStoreName: "minecraft",
					Property:      property,
					UpdateCount:   1,
					NewValue:      serializeForm(f.Describe()),
				},
			},
		},
	})
	s.writePacket(&packet.ClientBoundDataDrivenUIShowScreen{
		ScreenID:       f.ScreenID(),
		FormID:         formID,
		DataInstanceID: protocol.Option(instanceID),
	})
}

// CloseDDUIForms closes all active DDUI forms, calling OnClose on each.
func (h *DDUIFormHandler) CloseDDUIForms(s *Session) {
	h.mu.Lock()
	active := make([]*activeDDUIForm, 0, len(h.forms))
	for _, af := range h.forms {
		active = append(active, af)
	}
	h.forms = make(map[uint32]*activeDDUIForm)
	h.mu.Unlock()

	if len(active) == 0 {
		return
	}

	s.writePacket(&packet.ClientBoundDataDrivenUICloseScreen{})

	for _, af := range active {
		af.form.OnClose(ddui.CloseReasonProgrammaticAll)
		sendDataStoreCleanup(s, af)
	}
}

// sendDataStoreCleanup nulls the data store property for af.
func sendDataStoreCleanup(s *Session, af *activeDDUIForm) {
	s.writePacket(&packet.ClientBoundDataStore{
		Updates: []protocol.DataStoreChangeEntry{
			{
				ChangeType: protocol.DataStoreChangeTypeChange,
				Change: protocol.DataStoreChange{
					DataStoreName: "minecraft",
					Property:      af.property,
					UpdateCount:   af.propertyUpdateCount + 2,
					NewValue:      protocol.DataStorePropertyValue{Type: protocol.DataStorePropertyTypeNone},
				},
			},
		},
	})
}

// deriveProperty converts a DDUI screen ID and instance ID to a data store property name.
func deriveProperty(screenID string, instanceID uint32) string {
	base := strings.TrimPrefix(screenID, "minecraft:")
	base = strings.ReplaceAll(base, ":", "_")
	return base + "_data_" + strconv.FormatUint(uint64(instanceID), 10)
}

// serializeForm converts a ddui.FormDescriptor to a DataStorePropertyValue map.
func serializeForm(desc ddui.FormDescriptor) protocol.DataStorePropertyValue {
	if desc.Body != "" || desc.Button1.Label != "" || desc.Button2.Label != "" {
		return serializeMessageBox(desc)
	}
	return serializeCustomForm(desc)
}

// serializeCustomForm builds the nested data store map for a CustomForm.
func serializeCustomForm(desc ddui.FormDescriptor) protocol.DataStorePropertyValue {
	entries := make([]protocol.DataStoreMapEntry, 0, 3)

	if desc.HasCloseButton {
		entries = append(entries, dsEntry("closeButton", dsMap(
			dsEntry("button_visible", dsBool(true)),
			dsEntry("label", dsStr("Close")),
			dsEntry("onClick", dsInt(0)),
		)))
	}

	layoutEntries := make([]protocol.DataStoreMapEntry, 0, len(desc.Elements)+1)
	for i, elem := range desc.Elements {
		layoutEntries = append(layoutEntries, dsEntry(strconv.Itoa(i), serializeElement(elem)))
	}
	layoutEntries = append(layoutEntries, dsEntry("length", dsInt(int64(len(desc.Elements)))))

	entries = append(entries,
		dsEntry("layout", dsMap(layoutEntries...)),
		dsEntry("title", dsStr(desc.Title)),
	)
	return dsMap(entries...)
}

// serializeMessageBox builds the nested data store map for a MessageBox.
func serializeMessageBox(desc ddui.FormDescriptor) protocol.DataStorePropertyValue {
	btn1 := []protocol.DataStoreMapEntry{
		dsEntry("label", dsStr(desc.Button1.Label)),
		dsEntry("onClick", dsInt(0)),
	}
	if desc.Button1.Tooltip != "" {
		btn1 = append(btn1, dsEntry("tooltip", dsStr(desc.Button1.Tooltip)))
	}

	btn2 := []protocol.DataStoreMapEntry{
		dsEntry("label", dsStr(desc.Button2.Label)),
		dsEntry("onClick", dsInt(0)),
	}
	if desc.Button2.Tooltip != "" {
		btn2 = append(btn2, dsEntry("tooltip", dsStr(desc.Button2.Tooltip)))
	}

	return dsMap(
		dsEntry("body", dsStr(desc.Body)),
		dsEntry("button1", dsMap(btn1...)),
		dsEntry("button2", dsMap(btn2...)),
		dsEntry("title", dsStr(desc.Title)),
	)
}

// serializeElement converts a single ElementDescriptor to its data store map representation.
func serializeElement(e ddui.ElementDescriptor) protocol.DataStorePropertyValue {
	switch e.Kind {
	case ddui.ElementSpacer:
		return dsMap(
			dsEntry("spacer_visible", dsBool(true)),
			dsEntry("visible", dsBool(true)),
		)
	case ddui.ElementDivider:
		return dsMap(
			dsEntry("divider_visible", dsBool(true)),
			dsEntry("visible", dsBool(true)),
		)
	case ddui.ElementLabel:
		return dsMap(
			dsEntry("label_visible", dsBool(true)),
			dsEntry("text", dsStr(e.StringValue)),
			dsEntry("visible", dsBool(true)),
		)
	case ddui.ElementTextField:
		return dsMap(
			dsEntry("visible", dsBool(true)),
			dsEntry("description", dsStr(e.Description)),
			dsEntry("label", dsStr(e.Label)),
			dsEntry("text", dsStr(e.StringValue)),
			dsEntry("textfield_visible", dsBool(true)),
		)
	case ddui.ElementDropdown:
		return dsMap(
			dsEntry("visible", dsBool(true)),
			dsEntry("dropdown_visible", dsBool(true)),
			dsEntry("items", serializeDropdownItems(e.Options)),
			dsEntry("label", dsStr(e.Label)),
			dsEntry("value", dsInt(int64(e.IntValue))),
		)
	case ddui.ElementToggle:
		return dsMap(
			dsEntry("label", dsStr(e.Label)),
			dsEntry("toggle_visible", dsBool(true)),
			dsEntry("toggled", dsBool(e.BoolValue)),
			dsEntry("visible", dsBool(true)),
		)
	case ddui.ElementSlider:
		return dsMap(
			dsEntry("visible", dsBool(true)),
			dsEntry("description", dsStr(e.Description)),
			dsEntry("slider_visible", dsBool(true)),
			dsEntry("label", dsStr(e.Label)),
			dsEntry("maxValue", dsInt(int64(e.Max))),
			dsEntry("minValue", dsInt(int64(e.Min))),
			dsEntry("step", dsInt(int64(e.Step))),
			dsEntry("value", dsInt(int64(e.FloatValue))),
		)
	case ddui.ElementButton:
		return dsMap(
			dsEntry("button_visible", dsBool(true)),
			dsEntry("label", dsStr(e.Label)),
			dsEntry("onClick", dsInt(0)),
			dsEntry("visible", dsBool(true)),
		)
	}
	return dsMap()
}

// serializeDropdownItems encodes dropdown options as a length-keyed data store map.
func serializeDropdownItems(opts []ddui.DropdownOption) protocol.DataStorePropertyValue {
	entries := make([]protocol.DataStoreMapEntry, 0, len(opts)+1)
	for i, opt := range opts {
		entries = append(entries, dsEntry(strconv.Itoa(i), dsMap(
			dsEntry("label", dsStr(opt.Label)),
			dsEntry("value", dsInt(int64(opt.Value))),
		)))
	}
	entries = append(entries, dsEntry("length", dsInt(int64(len(opts)))))
	return dsMap(entries...)
}

// dsMap returns a map-typed DataStorePropertyValue from the given entries.
func dsMap(entries ...protocol.DataStoreMapEntry) protocol.DataStorePropertyValue {
	return protocol.DataStorePropertyValue{
		Type:     protocol.DataStorePropertyTypeMap,
		MapValue: entries,
	}
}

// dsBool returns a bool-typed DataStorePropertyValue.
func dsBool(v bool) protocol.DataStorePropertyValue {
	return protocol.DataStorePropertyValue{Type: protocol.DataStorePropertyTypeBool, BoolValue: v}
}

// dsInt returns an int64-typed DataStorePropertyValue.
func dsInt(v int64) protocol.DataStorePropertyValue {
	return protocol.DataStorePropertyValue{Type: protocol.DataStorePropertyTypeInt64, Int64Value: v}
}

// dsStr returns a string-typed DataStorePropertyValue.
func dsStr(v string) protocol.DataStorePropertyValue {
	return protocol.DataStorePropertyValue{Type: protocol.DataStorePropertyTypeString, StringValue: v}
}

// dsEntry returns a DataStoreMapEntry with the given key and value.
func dsEntry(key string, value protocol.DataStorePropertyValue) protocol.DataStoreMapEntry {
	return protocol.DataStoreMapEntry{Key: key, Value: value}
}
