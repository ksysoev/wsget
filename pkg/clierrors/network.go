package clierrors

type ConnectionClosed struct{}

func (e ConnectionClosed) Error() string {
	return "connection closed"
}

type Timeout struct{}

func (e Timeout) Error() string {
	return "timeout"
}
