package namedpipe

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"syscall"
	"time"
)

func GenerateRandomPipeName() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	gen := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, 16)
	for i := range b {
		b[i] = letters[gen.Intn(len(letters))]
	}
	return string(b)
}

func CreatePipe(path string) error {
	var err error
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = syscall.Mkfifo(path, 0600); err != nil {
			return fmt.Errorf("failed to create pipe %s: %v", path, err)
		}
	}
	return err
}

func OpenReadPipe(pipeName string) (*os.File, error) {
	r, err := os.OpenFile(pipeName, os.O_RDONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("could not open pipe %s for reading: %v", pipeName, err)
	}

	return r, nil
}

func OpenWritePipe(pipeName string) (*os.File, error) {
	w, err := os.OpenFile(pipeName, os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("could not open pipe %s for writing: %v", pipeName, err)
	}

	return w, nil
}

func GetTmpDir(serviceName string) string {
	return fmt.Sprintf("/tmp/Blueprint%s", serviceName)
}

func GetClientsList(serviceName string) ([]string, error) {
	clients := []string{}

	entries, err := os.ReadDir(GetTmpDir(serviceName))
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %v", GetTmpDir(serviceName), err)
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "args_pipe_") {
			uniqueSuffix := strings.TrimPrefix(e.Name(), "args_pipe_")
			clients = append(clients, uniqueSuffix)
		}
	}
	return clients, nil
}

func GetArgsPipeName(serviceName string, uniqueSuffix string) string {
	return fmt.Sprintf("%s/args_pipe_%s", GetTmpDir(serviceName), uniqueSuffix)
}

func GetResultPipeName(serviceName string, uniqueSuffix string) string {
	return fmt.Sprintf("%s/result_pipe_%s", GetTmpDir(serviceName), uniqueSuffix)
}
