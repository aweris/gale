package fs

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Exists checks if the given path exists.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)

	// Path exists
	if err == nil {
		return true, nil
	}

	// Path does not exist
	if os.IsNotExist(err) {
		return false, nil
	}

	// An error occurred while checking the path
	return false, err
}

// EnsureDir ensures that the given directory exists under the ghx data home directory.
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}

	return nil
}

// EnsureFile ensures that the given file exists under the ghx data home directory. If the file does not exist, it will
// be created.
//
// To ensure everything is working as expected, the method will also ensure that the directory of the file
// exists.
func EnsureFile(file string) error {
	// just to be sure that the directory exists
	if err := EnsureDir(filepath.Dir(file)); err != nil {
		return err
	}

	if _, err := os.Stat(file); os.IsNotExist(err) {
		// Create the file if it doesn't exist
		file, err := os.Create(file)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	return nil
}

// WriteFile ensures that the given file exists under the ghx data home directory. If the file does not
// exist, it will be created and the given content will be written to it.
//
// To ensure everything is working as expected, the method will also ensure that the directory of the file
// exists.
func WriteFile(file string, content []byte, permissions os.FileMode) error {
	if err := EnsureFile(file); err != nil {
		return err
	}

	if permissions == 0 {
		permissions = 0600
	}

	return os.WriteFile(file, content, permissions)
}

// ReadJSONFile reads the given path as a JSON file under the ghx data home directory and unmarshal it into the given value.
func ReadJSONFile[T any](file string, val *T) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// if the file is empty, we don't need to do anything
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, val)
}

// WriteJSONFile writes the given value to the given path as a JSON file under the ghx data home directory.
func WriteJSONFile(file string, val interface{}) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return WriteFile(file, data, 0600)
}

// ReadYAMLFile reads the given path as a YAML file.
func ReadYAMLFile[T any](file string, val *T) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// if the file is empty, we don't need to do anything
	if len(data) == 0 {
		return nil
	}

	return yaml.Unmarshal(data, val)
}

// CopyFile copies the given file from src to dst.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	s, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.Chmod(dst, s.Mode())
	if err != nil {
		return err
	}

	return nil
}
