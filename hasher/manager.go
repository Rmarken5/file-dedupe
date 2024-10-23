package hasher

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

type (
	Manager interface {
		Run(filePath string) (string, error)
	}
	FileManager struct {
		Hashes             map[string]*File
		FileHashDuplicates map[string][]*File
		ActualDuplicates   map[string][]*File
		hashLock           sync.RWMutex
		duplicateLock      sync.RWMutex
		fileLock           sync.RWMutex
	}
)

func NewManager() *FileManager {
	return &FileManager{
		Hashes:             make(map[string]*File),
		FileHashDuplicates: make(map[string][]*File),
		ActualDuplicates:   make(map[string][]*File),
		hashLock:           sync.RWMutex{},
		duplicateLock:      sync.RWMutex{},
		fileLock:           sync.RWMutex{},
	}
}

func (f *FileManager) Run(filePath string) (string, error) {

	err := f.getHashesForDirectory(filePath)
	if err != nil {
		return "", err
	}
	f.generateHashDuplicates()

	for k, v := range f.FileHashDuplicates {
		fmt.Printf("for %s\n", k)
		for _, f := range v {
			fmt.Println(f.FilePath)
		}
	}
	return "", err
}

func (f *FileManager) getHashesForDirectory(directoryPath string) error {
	fs, err := os.Open(directoryPath)
	if err != nil {
		return err
	}
	defer fs.Close()

	dir, err := fs.ReadDir(0)
	if err != nil {
		return err
	}

	var errs []error
	for _, d := range dir {
		if d.IsDir() {
			err := f.getHashesForDirectory(fmt.Sprintf("%s/%s", directoryPath, d.Name()))
			if err != nil {
				errs = append(errs, err)
			}
			continue
		}
		file, err := New(fmt.Sprintf("%s/%s", directoryPath, d.Name()))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		f.addHashToMap(file)
	}
	return errors.Join(errs...)
}

func (f *FileManager) addHashToMap(h *File) {
	f.hashLock.Lock()
	defer f.hashLock.Unlock()
	f.Hashes[h.FilePath] = h
}

func (f *FileManager) generateHashDuplicates() {
	hashes := make([]*File, len(f.Hashes))
	i := 0
	for _, file := range f.Hashes {
		hashes[i] = file
		i++
	}

	for i := 0; i < len(hashes); i++ {
		for k := i + 1; k < len(hashes); k++ {
			out := hashes[i]
			in := hashes[k]
			if out.FilePath != in.FilePath && out.MD5Hash == in.MD5Hash {
				fmt.Printf("out %s is equal to in %s \n", out.MD5Hash, in.FilePath)
				if _, ok := f.FileHashDuplicates[out.MD5Hash]; !ok {
					f.addFileHashDuplicates(out)
				}
				f.addFileHashDuplicates(in)
			}
		}
	}
}

func (f *FileManager) addFileHashDuplicates(h *File) {
	f.duplicateLock.Lock()
	defer f.duplicateLock.Unlock()
	f.FileHashDuplicates[h.MD5Hash] = append(f.FileHashDuplicates[h.MD5Hash], h)
}

func (f *FileManager) generateDuplicateFiles() {

}

func (f *FileManager) addFileToDuplicates(h *File) {
	f.fileLock.Lock()
	defer f.fileLock.Unlock()
	f.ActualDuplicates[h.MD5Hash] = append(f.ActualDuplicates[h.MD5Hash], h)
}
