package clierrors

type Interrupted struct{}

func (e Interrupted) Error() string {
	return "interrupted"
}

type UnknownCommand struct {
	Command string
}

func (e UnknownCommand) Error() string {
	return "unknown command: " + e.Command
}
