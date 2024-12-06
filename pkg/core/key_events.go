package core

type Key uint16

type KeyEvent struct {
	Key  Key
	Rune rune
}

const (
	KeyEsc          Key = 27
	KeyCtrlC        Key = 3
	KeyCtrlD        Key = 4
	KeyEnter        Key = 13
	KeyCtrlS        Key = 19
	KeyCtrlU        Key = 21
	KeySpace        Key = 32
	KeyAltBackspace Key = 23
	KeyBackspace    Key = 8
	MacBackspace2   Key = 127
	KeyDelete       Key = 65522
	KeyArrowLeft    Key = 65515
	KeyArrowRight   Key = 65514
	KeyArrowUp      Key = 65517
	KeyArrowDown    Key = 65516
	KeyTab          Key = 9
	KeyHome         Key = 65505
	KeyEnd          Key = 65507
)
