package clierrors

type Interrupted struct{}

func (e Interrupted) Error() string {
	return "interrupted"
}
