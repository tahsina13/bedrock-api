package models

import "encoding/json"

// Packet represents a collection of events to be sent together.
type Packet struct {
	Flag   byte    `json:"flag"`
	Events []Event `json:"events"`
}

// NewPacket creates a new Packet instance and adds the provided events to it.
func NewPacket(flag byte, events ...Event) *Packet {
	instance := Packet{
		Flag:   flag,
		Events: make([]Event, 0),
	}

	instance.AddEvents(events...)

	return &instance
}

// AddEvents adds one or more events to the packet and updates the event count accordingly.
func (p *Packet) AddEvents(events ...Event) {
	for _, event := range events {
		p.Events = append(p.Events, event)
	}
}

// ToBytes converts the Packet struct to a byte slice.
func (p Packet) ToBytes() []byte {
	b, _ := json.Marshal(p)
	return b
}

// PacketFromBytes converts a byte slice to a Packet struct.
func PacketFromBytes(bytes []byte) (*Packet, error) {
	var packet Packet

	err := json.Unmarshal(bytes, &packet)
	if err != nil {
		return nil, err
	}

	return &packet, nil
}
