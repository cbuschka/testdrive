package internal

func Run() (int, error) {
	session := NewSession()
	return session.Run()
}
