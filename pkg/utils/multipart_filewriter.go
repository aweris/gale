package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// MultipartFileWriter allows writing a file in chunks first and then merging the chunks into a single file.
// This is useful for uploading large files in chunks while ensuring files are not corrupted in case of network errors
// or not ordered chunk uploads.
type MultipartFileWriter struct {
	root string
}

// NewMultipartFileWriter creates a new MultipartFileWriter that writes files to the specified root directory.
func NewMultipartFileWriter(root string) (*MultipartFileWriter, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}

	return &MultipartFileWriter{root: root}, nil
}

// Write saves the uploaded part data as separate files in the same directory as the specified file.
func (w *MultipartFileWriter) Write(filePath string, offset int, data io.Reader) error {
	// Ensure absolute path for the file directory exists.
	fileDir := filepath.Join(w.root, filepath.Dir(filePath))
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return err
	}

	// Construct the part file name with offset
	partFileName := fmt.Sprintf("%s.part.%d", filepath.Base(filePath), offset)

	// Create the part file path in the same directory
	partFilePath := filepath.Join(fileDir, partFileName)

	// Write the part data to the part file
	partFile, err := os.Create(partFilePath)
	if err != nil {
		return err
	}
	defer partFile.Close()

	_, err = io.Copy(partFile, data)
	if err != nil {
		return err
	}

	return nil
}

// Merge combines the uploaded part files into separate intended files and removes the part files.
func (w *MultipartFileWriter) Merge() error {
	// Find all the part files in the root directory
	parts, err := findPartialFiles(w.root)
	if err != nil {
		return err
	}

	for target, parts := range parts {
		// Ensure that the part files are sorted by offset
		sort.Slice(parts, func(i, j int) bool { return parseOffset(parts[i]) < parseOffset(parts[j]) })

		// Create the target file
		targetFile, err := os.Create(target)
		if err != nil {
			return err
		}

		// Merge the part files into the target file
		for _, part := range parts {
			file, err := os.Open(part)
			if err != nil {
				return err
			}
			defer file.Close()

			// Copy the part file data to the target file
			_, err = io.Copy(targetFile, file)
			if err != nil {
				return err
			}

			// Remove the part file
			os.Remove(part)
		}
	}
	return err
}

// findPartialFiles finds all the part files in the root directory and returns a map of target file paths and their part file paths.
func findPartialFiles(root string) (map[string][]string, error) {
	parts := make(map[string][]string)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == root {
			return nil
		}

		// Check if the directory contains .part files
		if !d.IsDir() && strings.Contains(path, ".part.") {
			// Get the directory and filename from the provided file path
			dir := filepath.Dir(path)
			base := filepath.Base(path)

			// Get the original filename without the .part extension
			filename := strings.TrimSuffix(base, fmt.Sprintf(".part%s", filepath.Ext(base)))

			targetPath := filepath.Join(dir, filename)

			if _, ok := parts[targetPath]; !ok {
				parts[targetPath] = []string{}
			}

			// append the part file to the list
			parts[targetPath] = append(parts[targetPath], path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return parts, nil
}

// parseOffset parses the offset from the part file name. The part file name is of the format <filename>.<offset>.part
// and this function returns the offset.
func parseOffset(filePath string) int64 {
	ext := filepath.Ext(filePath)
	offset, _ := strconv.ParseInt(strings.TrimPrefix(ext, "."), 10, 64)
	return offset
}
