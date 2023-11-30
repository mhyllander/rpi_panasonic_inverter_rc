package codec

const (
	Power_Off = 0
	Power_On  = 1
)

type Config struct {
	Power uint8
	Raw   bool
	Clean bool
	Trace bool
	Diff  bool
	Param bool
}
