package order

type status int8

const (
	StatusNew status = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
)

func (s status) String() string {
	return [...]string{"NEW", "PROCESSING", "INVALID", "PROCESSED"}[s]
}
