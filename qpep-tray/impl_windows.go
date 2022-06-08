package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	BASEDIR_ENVIRONMENTVAR = "APPDATA"
	EXENAME                = "qpep.exe"
)

func getClientCommand() *exec.Cmd {
	exeFile := filepath.Join(ExeDir, EXENAME)
	handle, _ := syscall.GetCurrentProcess()

	attr := &syscall.SysProcAttr{
		HideWindow: true,
		CmdLine: fmt.Sprintf("--client "+
			"--acks %d "+
			"--ackDelay %d "+
			"--congestion %d "+
			"--decimate %d "+
			"--minBeforeDecimation %d "+
			"--gateway \"%s\" "+
			"--port %d "+
			"--multistream %v "+
			"--verbose %v "+
			"--varAckDelay %d "+
			"--threads %d ",
			qpepConfig.Acks,
			qpepConfig.AckDelay,
			qpepConfig.Congestion,
			qpepConfig.Decimate,
			qpepConfig.DelayDecimate,
			qpepConfig.GatewayHost,
			qpepConfig.GatewayPort,
			qpepConfig.MultiStream,
			qpepConfig.Verbose,
			qpepConfig.VarAckDelay,
			qpepConfig.WinDivertThreads),

		ParentProcess: handle,
	}

	cmd := exec.Command(exeFile)
	if cmd == nil {
		ErrorMsg("Could not create client command")
		return nil
	}
	cmd.Dir, _ = filepath.Abs(ExeDir)
	cmd.SysProcAttr = attr

	log.Println(cmd.Path)
	log.Println(cmd.Dir)
	log.Println(cmd.SysProcAttr.CmdLine)
	return cmd
}

func getServerCommand() *exec.Cmd {
	exeFile := filepath.Join(ExeDir, EXENAME)
	handle, _ := syscall.GetCurrentProcess()

	attr := &syscall.SysProcAttr{
		HideWindow: true,
		CmdLine: fmt.Sprintf("--acks %d "+
			"--ackDelay %d "+
			"--congestion %d "+
			"--decimate %d "+
			"--minBeforeDecimation %d "+
			"--gateway \"%s\" "+
			"--port %d "+
			"--multistream %v "+
			"--verbose %v "+
			"--varAckDelay %d "+
			"--threads %d ",
			qpepConfig.Acks,
			qpepConfig.AckDelay,
			qpepConfig.Congestion,
			qpepConfig.Decimate,
			qpepConfig.DelayDecimate,
			qpepConfig.GatewayHost,
			qpepConfig.GatewayPort,
			qpepConfig.MultiStream,
			qpepConfig.Verbose,
			qpepConfig.VarAckDelay,
			qpepConfig.WinDivertThreads),

		ParentProcess: handle,
	}

	cmd := exec.Command(exeFile)
	if cmd == nil {
		ErrorMsg("Could not create client command")
		return nil
	}
	cmd.Dir, _ = filepath.Abs(ExeDir)
	cmd.SysProcAttr = attr

	log.Println(cmd.Path)
	log.Println(cmd.Dir)
	log.Println(cmd.SysProcAttr.CmdLine)
	return cmd
}

func stopClientProcess() error {
	return stopProcess(clientCmd.Process.Pid)
}
func stopServerProcess() error {
	return stopProcess(serverCmd.Process.Pid)
}

func stopProcess(pid int) error {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		return ErrFailed
	}
	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		return ErrFailed
	}
	r, _, e := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r == 0 {
		return ErrFailed
	}

	return nil
}
