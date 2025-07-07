package executor

import (
	"errors"
	"jstarpl/jpm/api"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

const stopTimeoutTime = 3 * time.Second

type Process struct {
	Id           string
	Name         string
	Exec         string
	Dir          string
	ExitCode     int
	Arg          []string
	Env          []string
	Status       api.Status
	Cmd          *exec.Cmd
	RespawnDelay int
	FailCount    int
}

var processes map[string]Process

func ListProcesses() *[]api.Process {
	result := make([]api.Process, len(processes))
	i := 0
	for id, proc := range processes {
		result[i] = api.Process{
			Id:   id,
			Name: proc.Name,
			Exec: proc.Exec,
			Arg:  proc.Arg,
			Env:  proc.Env,
			Dir:  proc.Dir,
		}
		i++
	}

	return &result
}

func getNextProcessId() string {
	biggestId := 1
	for k := range processes {
		thisId, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		if biggestId < thisId {
			biggestId = thisId
		}
	}

	return strconv.FormatInt(int64(biggestId+1), 10)
}

func StartProcess(Name string, Exec string, Arg []string, Dir string, Env []string) (*Process, error) {
	newId := getNextProcessId()

	proc := Process{Id: newId, Name: Name, Exec: Exec, Dir: Dir, Arg: Arg, Env: Env, Status: api.Starting, Cmd: nil, ExitCode: 0, RespawnDelay: 0, FailCount: 0}
	processes[newId] = proc

	cmd := exec.Command(Exec, Arg...)
	cmd.Dir = Dir
	cmd.Env = Env

	proc.Cmd = cmd

	err := cmd.Start()

	log.Default().Printf("Starting %s as %s...", proc.Exec, newId)

	if err != nil {
		proc.Status = api.Failed
		return nil, err
	}

	proc.Status = api.Running

	go (func() {
		err := cmd.Wait()

		log.Default().Printf("%s finished", newId)

		proc.Status = api.Respawn

		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				proc.ExitCode = exiterr.ExitCode()
			} else {
				proc.ExitCode = 0
			}
		} else {
			proc.ExitCode = 0
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
	proc, ok := processes[Id]
	if !ok {
		return errors.New("Process Id not found")
	}

	proc.Status = api.Stopping

	cmd := proc.Cmd

	if cmd == nil {
		return errors.New("Process is not attached")
	}

	if runtime.GOOS == "windows" {
		log.Default().Printf("Shutting down %s with os.Kill", proc.Exec)
		cmd.Process.Signal(os.Kill)
	} else {
		log.Default().Printf("Shutting down %s with os.Interrupt", proc.Exec)
		cmd.Process.Signal(os.Interrupt)
		startTime := time.Now()
		for time.Since(startTime) < stopTimeoutTime {
			if cmd.ProcessState != nil {
				break
			}
			log.Default().Printf("Waiting for shutdown...")
			time.Sleep(100 * time.Millisecond)
		}
		if cmd.ProcessState == nil {
			log.Default().Printf("Forcing %s with Process.Kill()", proc.Exec)
			cmd.Process.Kill()
		}
	}

	proc.Status = api.Stopped

	return nil
}
