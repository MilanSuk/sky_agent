package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func _sandbox_isPathValid(name string) bool {
	curr, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	path, err := filepath.Abs(name)
	if err != nil {
		log.Fatal(err)
	}

	valid := strings.HasPrefix(path, curr)
	if !valid {
		SDK_Sandbox_violation(fmt.Errorf("path '%s' is outside of program '%s' folder", path, curr))
	}
	return valid
}

func exec_Command(name string, arg ...string) *exec.Cmd {
	SDK_Sandbox_violation(fmt.Errorf("command '%s' blocked", name))
	//exec.Command()
	return nil
}

func exec_CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	//exec.CommandContext()
	return nil
}

func exec_StartProcess(name string, argv []string, attr *os.ProcAttr) (*os.Process, error) {
	//os.StartProcess()
	return nil, fmt.Errorf("StartProcess(%s, %v, %v) was blocked", name, argv, attr)
}

func os_WriteFile(name string, data []byte, perm os.FileMode) error {
	if !_sandbox_isPathValid(name) {
		return fmt.Errorf("WriteFile(%s) outside program folder", name)
	}
	return os.WriteFile(name, data, perm)
}

func os_Mkdir(path string, perm os.FileMode) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Mkdir(%s) outside program folder", path)
	}
	return os.Mkdir(path, perm)
}
func os_MkdirAll(path string, perm os.FileMode) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("MkdirAll(%s) outside program folder", path)
	}
	return os.MkdirAll(path, perm)
}

func os_Remove(path string) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Remove(%s) outside program folder", path)
	}
	return os.Remove(path)
}

func os_RemoveAll(path string) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("RemoveAll(%s) outside program folder", path)
	}
	return os.RemoveAll(path)
}

func os_Rename(oldpath, newpath string) error {
	if !_sandbox_isPathValid(oldpath) || !_sandbox_isPathValid(newpath) {
		return fmt.Errorf("Rename(%s, %s) outside program folder", oldpath, newpath)
	}
	return os.Rename(oldpath, newpath)
}

func os_Chmod(path string, mode fs.FileMode) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf(" Chmod(%s) outside program folder", path)
	}
	return os.Chmod(path, mode)
}
func os_Chdir(path string) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Chdir(%s) outside program folder", path)
	}
	return os.Chdir(path)
}

func os_Create(path string) (*os.File, error) {
	if !_sandbox_isPathValid(path) {
		return nil, fmt.Errorf("Create(%s) outside program folder", path)
	}
	return os.Create(path)
}

func os_OpenFile(path string, flag int, perm fs.FileMode) (*os.File, error) {
	if !_sandbox_isPathValid(path) && flag != os.O_RDONLY {
		return nil, fmt.Errorf("OpenFile(%s) outside program folder", path)
	}
	return os.OpenFile(path, flag, perm)
}

func os_Lchown(path string, uid, gid int) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Lchown(%s) outside program folder", path)
	}
	return os.Lchown(path, uid, gid)
}

func os_Truncate(path string, size int64) error {
	if !_sandbox_isPathValid(path) {
		return fmt.Errorf("Truncate(%s) outside program folder", path)
	}
	return os.Truncate(path, size)
}

func os_Link(oldpath, newpath string) error {
	if !_sandbox_isPathValid(oldpath) || !_sandbox_isPathValid(newpath) {
		return fmt.Errorf("Link(%s, %s) outside program folder", oldpath, newpath)
	}
	return os.Link(oldpath, newpath)
}

func os_Symlink(oldpath, newpath string) error {
	if !_sandbox_isPathValid(oldpath) || !_sandbox_isPathValid(newpath) {
		return fmt.Errorf("Symlink(%s, %s) outside program folder", oldpath, newpath)
	}
	return os.Symlink(oldpath, newpath)
}

func os_NewFile(fd uintptr, path string) *os.File {
	if !_sandbox_isPathValid(path) {
		//fmt.Errorf("NewFile(%s) outside program folder", path)
		return nil
	}
	return os.NewFile(fd, path)
}
