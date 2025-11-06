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

	"github.com/teivah/broadcast"
)

const (
	stopTimeoutTime = 3 * time.Second
)

type Process struct {
	Id           string
	Name         string
	Namespace    string
	Exec         string
	Dir          string
	ExitCode     int
	Arg          []string
	Env          []string
	Status       api.Status
	Cmd          *exec.Cmd
	LastStarted  time.Time
	StartCount   int
	RespawnDelay int
	FailCount    int
	StdOutErr    broadcast.Relay[api.StdStreamMessage]
	StdIn        broadcast.Relay[api.StdStreamMessage]
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
		var uptime int
		if proc.Status == api.Running {
			uptime = int(time.Since(proc.LastStarted).Milliseconds())
		} else {
			uptime = 0
		}

		result[i] = api.Process{
			Id:         id,
			Name:       proc.Name,
			Namespace:  proc.Namespace,
			Exec:       proc.Exec,
			Arg:        proc.Arg,
			Env:        proc.Env,
			Dir:        proc.Dir,
			Uptime:     uptime,
			StartCount: proc.StartCount,
			Status:     proc.Status,
			ExitCode:   proc.ExitCode,
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

func StartProcess(Name string, Namespace string, Exec string, Arg []string, Dir string, Env []string) (*Process, error) {
	newId := getNextProcessId()

	stdOutErrRelay := broadcast.NewRelay[api.StdStreamMessage]()
	stdInRelay := broadcast.NewRelay[api.StdStreamMessage]()
	proc := Process{Id: newId, Name: Name, Namespace: Namespace, Exec: Exec, Dir: Dir, Arg: Arg, Env: Env, Status: api.Starting, Cmd: nil, ExitCode: 0, RespawnDelay: 0, FailCount: 0, StdOutErr: *stdOutErrRelay, StdIn: *stdInRelay}
	processes[newId] = &proc

	cmd := exec.Command(Exec, Arg...)
	cmd.Dir = Dir
	cmd.Env = Env

	stdout, err := cmd.StdoutPipe()
	checkError(err)
	stderr, err := cmd.StderrPipe()
	checkError(err)
	_, err = cmd.StdinPipe()
	checkError(err)

	proc.Cmd = cmd
	proc.StartCount = 1
	proc.LastStarted = time.Now()

	err = cmd.Start()

	logger.Printf("Starting %s as %s...", proc.Exec, newId)

	if err != nil {
		logger.Printf("Failed to start %s: %v", newId, err)
		proc.Status = api.Failed
		return nil, err
	}

	proc.Status = api.Running

	go readerCopyToRelay(stdOutErrRelay, stdout, api.Stdout)
	go readerCopyToRelay(stdOutErrRelay, stderr, api.Stderr)

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

func GetProcessStdInStreamRelay(Id string) (*broadcast.Relay[api.StdStreamMessage], error) {
	processes := processes[Id]
	if processes == nil {
		return nil, errors.New("Process not found")
	}

	return &processes.StdIn, nil
}

func GetProcessStdStreamRelay(Id string) (*broadcast.Relay[api.StdStreamMessage], error) {
	process := processes[Id]
	if process == nil {
		return nil, errors.New("Process not found")
	}

	return &process.StdOutErr, nil
}

func readerCopyToRelay(dst *broadcast.Relay[api.StdStreamMessage], src io.Reader, streamType api.StreamType) {
	for {
		buf := make([]byte, 1024)
		read, err := src.Read(buf)

		dst.Broadcast(api.StdStreamMessage{
			StreamType: streamType,
			Data:       buf[0:read],
		})

		if err == io.EOF {
			return
		} else if err != nil {
			logger.Printf("Error while reading from stream %v", err)
		}
	}
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

	proc.StdOutErr.Close()
	proc.StdIn.Close()

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
