package executor

import (
	"errors"
	"io"
	"io/fs"
	"jstarpl/jpm/api"
	"jstarpl/jpm/service/logger"
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
	StdOutErr    *broadcast.Relay[api.StdStreamMessage]
	StdIn        *broadcast.Relay[api.StdStreamMessage]
	Logger       *logger.ProcessLogger
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
var execLog *log.Logger

var logsDir string
var logRetentionDays = logger.DefaultRetentionDays

// SetLogConfig configures the directory for process log files and the number
// of days to retain them. Call before starting any processes.
func SetLogConfig(dir string, retentionDays int) {
	logsDir = dir
	logRetentionDays = retentionDays
}

const logProps = log.Lmicroseconds | log.Ltime | log.Ldate | log.LUTC

func init() {
	processes = make(map[string]*Process)
	execLog = log.New(log.Default().Writer(), "executor: ", logProps)
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
	proc := Process{Id: newId, Name: Name, Namespace: Namespace, Exec: Exec, Dir: Dir, Arg: Arg, Env: Env, Status: api.Starting, Cmd: nil, ExitCode: 0, RespawnDelay: 0, FailCount: 0, StdOutErr: stdOutErrRelay, StdIn: stdInRelay}
	processes[newId] = &proc

	if logsDir != "" {
		pl, err := logger.NewProcessLogger(logsDir, newId, Name, logRetentionDays)
		if err != nil {
			execLog.Printf("Warning: could not create process logger for %s: %v", newId, err)
		} else {
			proc.Logger = pl
			go logRelayToFile(proc.StdOutErr, pl)
		}
	}

	err := startProcess(&proc)
	if err != nil {
		return nil, err
	}

	return &proc, nil
}

// logRelayToFile subscribes to a process stdout/stderr relay and writes all
// messages to the given ProcessLogger. It closes the logger when the relay closes.
func logRelayToFile(relay *broadcast.Relay[api.StdStreamMessage], pl *logger.ProcessLogger) {
	l := relay.Listener(1)
	for msg := range l.Ch() {
		if err := pl.Write(msg); err != nil {
			execLog.Printf("Error writing to process log: %v", err)
		}
	}
	if err := pl.Close(); err != nil {
		execLog.Printf("Error closing process log: %v", err)
	}
}

func startProcess(proc *Process) error {
	if proc == nil {
		return errors.New("Process is nil")
	}

	cmd := exec.Command(proc.Exec, proc.Arg...)
	cmd.Dir = proc.Dir
	cmd.Env = proc.Env

	stdout, err := cmd.StdoutPipe()
	checkError(err)
	stderr, err := cmd.StderrPipe()
	checkError(err)
	stdin, err := cmd.StdinPipe()
	checkError(err)

	proc.Cmd = cmd
	proc.StartCount++
	proc.LastStarted = time.Now()
	proc.Status = api.Starting

	err = cmd.Start()

	execLog.Printf("Starting %s as %s...", proc.Exec, proc.Id)

	if err != nil {
		execLog.Printf("Failed to start %s: %v", proc.Id, err)
		proc.Status = api.Failed
		return err
	}

	proc.Status = api.Running

	go readerCopyToRelay(proc.StdOutErr, stdout, api.Stdout)
	go readerCopyToRelay(proc.StdOutErr, stderr, api.Stderr)
	go relayCopyToWriter(proc.StdIn, stdin)

	go (func() {
		err := cmd.Wait()

		execLog.Printf("%s finished", proc.Id)

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

	return nil
}

func relayCopyToWriter(src *broadcast.Relay[api.StdStreamMessage], dst io.Writer) {
	l := src.Listener(1)
	defer l.Close()

	for msg := range l.Ch() {
		if len(msg.Data) == 0 {
			continue
		}

		_, err := dst.Write(msg.Data)
		if err != nil {
			execLog.Printf("Error while writing to stdin stream %v", err)
			return
		}
	}
}

func GetProcessStdInStreamRelay(Id string) (*broadcast.Relay[api.StdStreamMessage], error) {
	processes := processes[Id]
	if processes == nil {
		return nil, errors.New("Process not found")
	}

	return processes.StdIn, nil
}

func GetProcessStdStreamRelay(Id string) (*broadcast.Relay[api.StdStreamMessage], error) {
	process := processes[Id]
	if process == nil {
		return nil, errors.New("Process not found")
	}

	return process.StdOutErr, nil
}

func RestartProcess(Id string) error {
	proc, ok := processes[Id]
	if !ok {
		return errors.New("Process Id not found")
	}

	if proc.Status > api.Stopped {
		err := StopProcess(Id)
		if err != nil {
			return err
		}
	}

	return startProcess(proc)
}

func readerCopyToRelay(dst *broadcast.Relay[api.StdStreamMessage], src io.Reader, streamType api.StreamType) {
	for {
		buf := make([]byte, 1024)
		read, err := src.Read(buf)

		dst.Broadcast(api.StdStreamMessage{
			StreamType: streamType,
			Data:       buf[0:read],
		})

		if (errors.Is(err, io.EOF)) || (errors.Is(err, io.ErrClosedPipe)) || (errors.Is(err, fs.ErrClosed)) {
			return
		} else if err != nil {
			execLog.Printf("Error while reading from stream %v", err)
			return
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

	if proc.StdOutErr != nil {
		proc.StdOutErr.Close()
	}

	if proc.StdIn != nil {
		proc.StdIn.Close()
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
		execLog.Printf("Shutting down %s with os.Kill", proc.Exec)
		cmd.Process.Signal(os.Kill)
	} else {
		execLog.Printf("Shutting down %s with os.Interrupt", proc.Exec)
		cmd.Process.Signal(os.Interrupt)
		startTime := time.Now()
		for time.Since(startTime) < stopTimeoutTime {
			if cmd.ProcessState != nil {
				break
			}
			execLog.Printf("Waiting for shutdown...")
			time.Sleep(100 * time.Millisecond)
		}
		if cmd.ProcessState == nil {
			execLog.Printf("Forcing %s with Process.Kill()", proc.Exec)
			cmd.Process.Kill()
		}
	}

	proc.Status = api.Stopped

	return nil
}
