package item

// Leather is Dropped by Cows! Need to add Mobs too but still Leather is now here
type Leather struct{}

// EncodeItem ...
func (Leather) EncodeItem() (id int32, meta int16) {
    return 334, 0
}


