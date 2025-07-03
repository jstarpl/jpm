package processrunner

import (
	"errors"
	"os/exec"
	"strconv"
)

type Status int

const (
	Starting Status = 1
	Running  Status = 2
	Respawn  Status = 3
	Stopped  Status = 0
	Failed   Status = -1
)

type Process struct {
	Id           string
	Name         string
	Exec         string
	ExitCode     int
	Arg          []string
	Env          []string
	Status       Status
	Cmd          *exec.Cmd
	RespawnDelay int
	FailCount    int
}

var processes map[string]Process

func ListProcesses() *map[string]Process {
	return &processes
}

func getNextProcessId() string {
	biggestId := 0
	for k, _ := range processes {
		thisId, err := strconv.Atoi(example)
		if err != nil {
			continue
		}
		if biggestId < thisId {
			biggestId = thisId
		}
	}

	return strconv.FormatInt(biggestId+1, 10)
}

func StartProcess(Name string, Exec string, Arg []string, Env []string) (*Process, error) {
	cmd := exec.Command(Exec, Arg...)
	cmd.Env = Env

	newId := getNextProcessId()

	proc := Process{Id: newId, Name: Name, Exec: Exec, Arg: Arg, Env: Env, Status: Starting, Cmd: cmd, ExitCode: 0, RespawnDelay: 0, FailCount: 0}
	processes[newId] = proc

	err := cmd.Start()

	if err != nil {
		return nil, err
	}

	go (func() {
		err := cmd.Wait()

		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				proc.ExitCode = exiterr.ExitCode()
			} else {
				proc.ExitCode = 0
			}
		}
	})()

	return &proc, nil
}

func DeleteProcess(Id string) error {
	proc, ok := processes[Id]
	if !ok {
		return errors.New("Process Id not found")
	}

	err := StopProcess(proc.Id)
	if err != nil {
		return err
	}

	delete(processes, Id)

	return nil
}

func StopProcess(Id string) error {
	return nil
}
