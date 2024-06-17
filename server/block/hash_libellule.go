package block

const (
	hashButton = iota + 200
	hashDropper
	hashHopper
	hashLever
	hashMoving
	hashObserver
	hashPiston
	hashPistonArmCollision
	hashPressurePlate
	hashRedstoneBlock
	hashRedstoneComparator
	hashRedstoneLamp
	hashRedstoneRepeater
	hashRedstoneTorch
	hashRedstoneWire
	hashSlime
	hashIronDoor
)

func (b Button) Hash() uint64 {
	return hashButton | uint64(b.Type.Uint8())<<8 | uint64(b.Facing)<<14 | uint64(boolByte(b.Pressed))<<17
}

func (d Dropper) Hash() uint64 {
	return hashDropper | uint64(d.Facing)<<8 | uint64(boolByte(d.Powered))<<11
}

func (h Hopper) Hash() uint64 {
	return hashHopper | uint64(h.Facing)<<8 | uint64(boolByte(h.Powered))<<11
}

// Hash ...
func (d IronDoor) Hash() uint64 {
	return hashIronDoor | uint64(d.Facing)<<8 | uint64(boolByte(d.Open))<<10 | uint64(boolByte(d.Top))<<11 | uint64(boolByte(d.Right))<<12
}

func (l Lever) Hash() uint64 {
	return hashLever | uint64(boolByte(l.Powered))<<8 | uint64(l.Facing)<<9 | uint64(l.Direction)<<12
}

func (Moving) Hash() uint64 {
	return hashMoving
}

func (o Observer) Hash() uint64 {
	return hashObserver | uint64(o.Facing)<<8 | uint64(boolByte(o.Powered))<<11
}

func (p Piston) Hash() uint64 {
	return hashPiston | uint64(p.Facing)<<8 | uint64(boolByte(p.Sticky))<<11
}

func (c PistonArmCollision) Hash() uint64 {
	return hashPistonArmCollision | uint64(c.Facing)<<8
}
func (RedstoneBlock) Hash() uint64 {
	return hashRedstoneBlock
}

func (r RedstoneComparator) Hash() uint64 {
	return hashRedstoneComparator | uint64(r.Facing)<<8 | uint64(boolByte(r.Subtract))<<10 | uint64(boolByte(r.Powered))<<11 | uint64(r.Power)<<12
}

func (l RedstoneLamp) Hash() uint64 {
	return hashRedstoneLamp | uint64(boolByte(l.Lit))<<8
}

func (r RedstoneRepeater) Hash() uint64 {
	return hashRedstoneRepeater | uint64(r.Facing)<<8 | uint64(boolByte(r.Powered))<<10 | uint64(r.Delay)<<11
}

func (t RedstoneTorch) Hash() uint64 {
	return hashRedstoneTorch | uint64(t.Facing)<<8 | uint64(boolByte(t.Lit))<<11
}

func (r RedstoneWire) Hash() uint64 {
	return hashRedstoneWire | uint64(r.Power)<<8
}

func (Slime) Hash() uint64 {
	return hashSlime
}

func (p PressurePlate) Hash() uint64 {
	return hashPressurePlate | uint64(boolByte(p.Powered))<<8
}
