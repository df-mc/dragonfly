package item

type LapisLazuli struct{}

func (LapisLazuli) EncodeItem() (id int32, meta int16) {
	return 351, 4
}