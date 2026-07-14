package entity

// BaseBehaviour provides shared runtime state for Ent behaviours. Embed it
// to inherit common functionality, or forward methods to another instance.
type BaseBehaviour struct {
	portalTravel *PortalTravelComputer
}

// NewBaseBehaviour returns a BaseBehaviour initialised with the default Ent runtime behaviour.
func NewBaseBehaviour() BaseBehaviour {
	return BaseBehaviour{portalTravel: NewPortalTravelComputer()}
}

// PortalTravelComputer returns the portal travel state for a behaviour.
func (b *BaseBehaviour) PortalTravelComputer() *PortalTravelComputer {
	if b.portalTravel == nil {
		b.portalTravel = NewPortalTravelComputer()
	}
	return b.portalTravel
}
