package order

type status int8

func (s status) String() string {
	switch s {
	case StatusNew:
		return "NEW"
	case StatusProcessing:
		return "PROCESSING"
	case StatusInvalid:
		return "INVALID"
	case StatusProcessed:
		return "PROCESSED"
	}
	return ""
}

const (
	StatusNew status = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
)
