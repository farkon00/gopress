package shared

type Mode int

const (
	ENCODE_MODE Mode = iota
	DECODE_MODE Mode = iota
)

type Config struct {
	Filename string
	Output   string
	Mode     Mode
}
