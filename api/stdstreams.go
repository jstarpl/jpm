package api

type StreamType string

const (
	Stdout StreamType = "stdout"
	Stderr StreamType = "stderr"
	Stdin  StreamType = "stdin"
)

type StdStreamMessage struct {
	StreamType StreamType
	Data       []byte
}
