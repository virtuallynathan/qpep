package main

const string (
	BASEDIR_ENVIRONMENTVAR = "HOME"
	EXENAME                = "qpep"
)

func getSysAttributes() *exec.SysAttributes {
	exeFile := filepath.Join(ExeDir, EXENAME)
	handle, _ := syscall.GetCurrentProcess()
	
	return nil
}

func stopClientProcess() error {
	proc, err := os.FindProcess(clientCmd.Process.Pid)
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
