package backgroundtasks

func Init() error {
	go statsRoutine()
	go certificatesRoutine()
	go cleaningRoutine()

	return nil
}
