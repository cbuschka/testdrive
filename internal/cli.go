package internal

func Run() (int, error) {

	session, err := NewSession()
	if err != nil {
		return -1, err
	}

	err = session.LoadConfig("testdrive.yaml")
	if err != nil {
		return -1, err
	}

	return session.Run()
}
