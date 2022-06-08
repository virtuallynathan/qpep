package main

const string (
	BASEDIR_ENVIRONMENTVAR = "HOME"
	EXENAME                = "qpep"
)

func stopClientProcess() error {
	return stopProcess(clientCmd.Process.Pid)
}
func stopServerProcess() error {
	return stopProcess(serverCmd.Process.Pid)
}

func stopProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		ErrorMsg("Could not terminate client process: %v", err)
		return ErrFailed
	}

	log.Println("Waiting for client exe to terminate")
	if err := proc.Signal(syscall.SIGINT); err != nil {
		ErrorMsg("Could not terminate client process: %v", err)
		return ErrFailed
	}

	return nil
}
