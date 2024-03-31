package lsportal

import (
	"io"
	"os/exec"
)

func StartLanguageServer(command string, args []string) (io.ReadWriteCloser, error) {
	// Create a new command instance
	cmd := exec.Command(command, args...)

	// Create pipes for stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Create a custom ReadWriteCloser that combines stdin and stdout
	rwc := &cmdReadWriteCloser{
		stdin:  stdin,
		stdout: stdout,
		cmd:    cmd,
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	//make readwriteCloser

	return rwc, nil
}

type cmdReadWriteCloser struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cmd    *exec.Cmd
}

func (rwc *cmdReadWriteCloser) Read(p []byte) (n int, err error) {
	return rwc.stdout.Read(p)
}

func (rwc *cmdReadWriteCloser) Write(p []byte) (n int, err error) {
	return rwc.stdin.Write(p)
}

func (rwc *cmdReadWriteCloser) Close() error {
	err := rwc.stdin.Close()
	if err != nil {
		return err
	}
	return rwc.cmd.Wait()
}
