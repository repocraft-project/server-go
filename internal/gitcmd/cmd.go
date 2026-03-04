package gitcmd

import (
	"bytes"
	"os/exec"
	"strings"
)

func Exec(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", &GitError{
				args:   args,
				err:    err,
				stderr: string(ee.Stderr),
			}
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

type GitError struct {
	args   []string
	err    error
	stderr string
}

func (e *GitError) Error() string {
	return e.err.Error()
}

func HashObject(data []byte) (string, error) {
	cmd := exec.Command("git", "hash-object", "-t", "blob", "-w", "--stdin")
	cmd.Stdin = bytes.NewReader(data)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func CatFile(hash string) (string, error) {
	return Exec("cat-file", "-p", hash)
}

func UpdateRef(name, hash string) error {
	_, err := Exec("update-ref", name, hash)
	return err
}

func DeleteRef(name string) error {
	_, err := Exec("update-ref", "-d", name)
	return err
}

func RevParse(ref string) (string, error) {
	return Exec("rev-parse", ref)
}

func RevParseWithDir(dir, ref string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", ref)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func ShowRef() (string, error) {
	return Exec("show-ref")
}

func Init(dir string) error {
	cmd := exec.Command("git", "init", dir)
	return cmd.Run()
}
