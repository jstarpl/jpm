package executor

import (
	"errors"
	"io"
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

type ProcessStatusChangeEvent struct {
	Process Process
}

type ProcessDeletedEvent struct {
	DeletedProcessId string
}

type Event struct {
	*ProcessStatusChangeEvent
	*ProcessDeletedEvent
}

var processes map[string]*Process
var logger *log.Logger

const logProps = log.Lmicroseconds | log.Ltime | log.Ldate | log.LUTC

func init() {
	processes = make(map[string]*Process)
	logger = log.New(log.Default().Writer(), "executor: ", logProps)
}

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

func ListProcesses() *[]api.Process {
	result := make([]api.Process, len(processes))
	i := 0
	for id, proc := range processes {
		result[i] = api.Process{
			Id:       id,
			Name:     proc.Name,
			Exec:     proc.Exec,
			Arg:      proc.Arg,
			Env:      proc.Env,
			Dir:      proc.Dir,
			Status:   proc.Status,
			ExitCode: proc.ExitCode,
		}
		i++
	}

	return &result
}

func getNextProcessId() string {
	biggestId := -1
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
	processes[newId] = &proc

	cmd := exec.Command(Exec, Arg...)
	cmd.Dir = Dir
	cmd.Env = Env

	stdout, err := cmd.StdoutPipe()
	checkError(err)
	stderr, err := cmd.StderrPipe()
	checkError(err)

	proc.Cmd = cmd

	err = cmd.Start()

	logger.Printf("Starting %s as %s...", proc.Exec, newId)

	if err != nil {
		logger.Printf("Failed to start %s: %v", newId, err)
		proc.Status = api.Failed
		return nil, err
	}

	proc.Status = api.Running

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	go (func() {
		err := cmd.Wait()

		logger.Printf("%s finished", newId)

		if proc.Status != api.Stopped && proc.Status != api.Stopping {
			proc.Status = api.Respawn
		}

		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				proc.ExitCode = exiterr.ExitCode()
			} else {
				proc.ExitCode = 0
			}
		} else {
			proc.ExitCode = 0
		}

		proc.Cmd = nil
	})()

	return &proc, nil
}

func DeleteProcess(Id string) error {
	proc, ok := processes[Id]
	if !ok {
		return errors.New("Process Id not found")
	}

	if proc.Status > 0 {
		err := StopProcess(proc.Id)
		if err != nil {
			return err
		}
	}

	delete(processes, Id)

	return nil
}

func StopProcess(Id string) error {
	proc, ok := processes[Id]
	if !ok {
		return errors.New("Process Id not found")
	}

	cmd := proc.Cmd

	proc.Status = api.Stopping

	if cmd == nil {
		return errors.New("Process is not attached")
	}

	if runtime.GOOS == "windows" {
		logger.Printf("Shutting down %s with os.Kill", proc.Exec)
		cmd.Process.Signal(os.Kill)
	} else {
		logger.Printf("Shutting down %s with os.Interrupt", proc.Exec)
		cmd.Process.Signal(os.Interrupt)
		startTime := time.Now()
		for time.Since(startTime) < stopTimeoutTime {
			if cmd.ProcessState != nil {
				break
			}
			logger.Printf("Waiting for shutdown...")
			time.Sleep(100 * time.Millisecond)
		}
		if cmd.ProcessState == nil {
			logger.Printf("Forcing %s with Process.Kill()", proc.Exec)
			cmd.Process.Kill()
		}
	}

	proc.Status = api.Stopped

	return nil
}
