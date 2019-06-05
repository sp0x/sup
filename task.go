package sup

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
)

// Task represents a set of commands to be run.
type Task struct {
	Run     string
	Input   io.Reader
	Output  io.Writer
	Clients []Client
	TTY     bool
}

func (t *Task) String() string {
	return fmt.Sprintf("%v", t.Run)
}

func (sup *Stackup) createTasks(cmd *Command, clients []Client, env string) ([]*Task, error) {
	var tasks []*Task

	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "resolving CWD failed")
	}

	// Anything to upload?
	for _, upload := range cmd.Upload {
		uploadFile, err := ResolveLocalPath(cwd, upload.Src, env)
		if err != nil {
			return nil, errors.Wrap(err, "upload: "+upload.Src)
		}
		log.Println(fmt.Sprintf("Tarring source: %v %v", cwd, uploadFile))
		uploadTarReader, err := NewTarStreamReader(cwd, uploadFile, upload.Exc)

		if err != nil {
			return nil, errors.Wrap(err, "upload: "+upload.Src)
		}

		task := Task{
			Run:   RemoteTarCommand(upload.Dst),
			Input: uploadTarReader,
			TTY:   false,
		}
		log.Println(fmt.Sprintf("Tarring: %v", task))
		addTask(task, cmd, clients, &tasks)
	}

	for _, download := range cmd.Download {
		dst, err := ResolveLocalPath(cwd, download.Dst, env)
		if err != nil {
			return nil, errors.Wrap(err, "download: "+download.Dst)
		}
		tarWriter, err := NewTarStreamWriter(dst)
		if err != nil {
			return nil, errors.Wrap(err, "download: "+download.Dst)
		}

		// todo: support download.Exclude
		task := Task{
			Run:    RemoteTarCreateCommand(download.SrcFolder, download.Src),
			Output: tarWriter,
			TTY:    false,
		}

		addTask(task, cmd, clients, &tasks)
	}

	// Script. Read the file as a multiline input command.
	if cmd.Script != "" {
		f, err := os.Open(cmd.Script)
		if err != nil {
			return nil, errors.Wrap(err, "can't open script")
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, errors.Wrap(err, "can't read script")
		}

		task := Task{
			Run: string(data),
			TTY: true,
		}
		if sup.debug {
			task.Run = "set -x;" + task.Run
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}
		addTask(task, cmd, clients, &tasks)
	}

	// Local command.
	if cmd.Local != "" {
		local := &LocalhostClient{
			env: env + `export SUP_HOST="localhost";`,
		}
		local.Connect("localhost")
		task := &Task{
			Run:     cmd.Local,
			Clients: []Client{local},
			TTY:     true,
		}
		if sup.debug {
			task.Run = "set -x;" + task.Run
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}
		tasks = append(tasks, task)
	}

	// Remote command.
	if cmd.Run != "" {
		task := Task{
			Run: cmd.Run,
			TTY: true,
		}
		if cmd.Chdir != "" {
			chdir := cmd.Chdir // Maybe resolve it?
			task.Run = fmt.Sprintf("cd %v; ", chdir) + task.Run
		}
		if sup.debug {
			task.Run = "set -x;" + task.Run
		}
		if cmd.Stdin {
			task.Input = os.Stdin
		}
		addTask(task, cmd, clients, &tasks)
	}

	return tasks, nil
}

func addTask(task Task, cmd *Command, clients []Client, tasks *[]*Task) {
	if cmd.Once {
		task.Clients = []Client{clients[0]}
		*tasks = append(*tasks, &task)
	} else if cmd.Serial > 0 {
		// Each "serial" task client group is executed sequentially.
		for i := 0; i < len(clients); i += cmd.Serial {
			j := i + cmd.Serial
			if j > len(clients) {
				j = len(clients)
			}
			copy := task
			copy.Clients = clients[i:j]
			*tasks = append(*tasks, &copy)
		}
	} else {
		task.Clients = clients
		*tasks = append(*tasks, &task)
	}
}

type ErrTask struct {
	Task   *Task
	Reason string
}

func (e ErrTask) Error() string {
	return fmt.Sprintf(`Run("%v"): %v`, e.Task, e.Reason)
}
