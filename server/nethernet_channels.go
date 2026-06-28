package server

import (
	"bytes"
	"context"
	"net"

	"github.com/df-mc/go-nethernet"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

const maxNetherNetUnreliableBatchSize = 30000

type netherNetChannelNetwork struct {
	minecraft.NetherNet
	unreliableMovement  bool
	unreliableEphemeral bool
}

func (n netherNetChannelNetwork) Listen(address string) (minecraft.NetworkListener, error) {
	l, err := n.NetherNet.Listen(address)
	if err != nil {
		return nil, err
	}
	return netherNetChannelListener{
		NetworkListener:     l,
		unreliableMovement:  n.unreliableMovement,
		unreliableEphemeral: n.unreliableEphemeral,
	}, nil
}

type netherNetChannelListener struct {
	minecraft.NetworkListener
	unreliableMovement  bool
	unreliableEphemeral bool
}

func (l netherNetChannelListener) Accept() (net.Conn, error) {
	conn, err := l.NetworkListener.Accept()
	if err != nil {
		return nil, err
	}
	nnConn, ok := conn.(*nethernet.Conn)
	if !ok || (!l.unreliableMovement && !l.unreliableEphemeral) {
		return conn, nil
	}
	return netherNetChannelConn{
		Conn:                nnConn,
		unreliableMovement:  l.unreliableMovement,
		unreliableEphemeral: l.unreliableEphemeral,
	}, nil
}

type netherNetChannelConn struct {
	*nethernet.Conn
	unreliableMovement  bool
	unreliableEphemeral bool
}

func (conn netherNetChannelConn) Write(b []byte) (int, error) {
	if len(b) <= maxNetherNetUnreliableBatchSize && netherNetUnreliableBatch(b, conn.unreliableMovement, conn.unreliableEphemeral) {
		if n, err := conn.Send(b, nethernet.MessageReliabilityUnreliable); err == nil {
			return n, nil
		}
	}
	return conn.Conn.Write(b)
}

func (conn netherNetChannelConn) Context() context.Context {
	return conn.Conn.Context()
}

func netherNetUnreliableBatch(data []byte, movement, ephemeral bool) bool {
	if payload, ok := netherNetCompressedPayload(data); ok && netherNetUnreliablePayload(payload, movement, ephemeral) {
		return true
	}
	return netherNetUnreliablePayload(data, movement, ephemeral)
}

func netherNetCompressedPayload(data []byte) ([]byte, bool) {
	if len(data) == 0 {
		return nil, false
	}
	if data[0] == byte(packet.NopCompression.EncodeCompression()) {
		return data[1:], true
	}
	if compression, ok := packet.CompressionByID(uint16(data[0])); ok {
		payload, err := compression.Decompress(data[1:], 16*1024*1024)
		return payload, err == nil
	}
	return nil, false
}

func netherNetUnreliablePayload(payload []byte, movement, ephemeral bool) bool {
	buf := bytes.NewBuffer(payload)
	seen := false
	for buf.Len() != 0 {
		var length uint32
		if err := protocol.Varuint32(buf, &length); err != nil || length == 0 || length > uint32(buf.Len()) {
			return false
		}
		pk := bytes.NewBuffer(buf.Next(int(length)))
		var header packet.Header
		if err := header.Read(pk); err != nil || !netherNetUnreliablePacket(header.PacketID, movement, ephemeral) {
			return false
		}
		seen = true
	}
	return seen
}

func netherNetUnreliablePacket(id uint32, movement, ephemeral bool) bool {
	if movement {
		switch id {
		case packet.IDMoveActorAbsolute, packet.IDMoveActorDelta, packet.IDMovePlayer:
			return true
		}
	}
	if ephemeral {
		switch id {
		case packet.IDSetActorMotion,
			packet.IDSetTime,
			packet.IDAnimate,
			packet.IDAnimateEntity,
			packet.IDActorEvent,
			packet.IDLevelEvent,
			packet.IDLevelEventGeneric,
			packet.IDCameraShake,
			packet.IDSpawnParticleEffect,
			packet.IDOnScreenTextureAnimation,
			packet.IDEmote,
			packet.IDMotionPredictionHints:
			return true
		}
	}
	return false
}
