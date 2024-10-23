package hasher

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

	err = f.generateDuplicateFiles(f.FileHashDuplicates)
	if err != nil {
		return "", err
	}

	for k, m := range f.ActualDuplicates {
		fmt.Printf("for hash %s, duplicate files: \n", k)
		for _, v := range m {
			fmt.Printf("%v\n", v.FilePath)
		}
	}

	return "", nil
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
				fmt.Printf("out %s is equal to in %s \n", out.MD5Hash, in.MD5Hash)
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

func (f *FileManager) generateDuplicateFiles(hashDupes map[string][]*File) error {
	for _, list := range hashDupes {
		for i := 0; i < len(list); i++ {
			for k := i + 1; k < len(list); k++ {
				out := list[i]
				in := list[k]
				isDup, err := IsFileDuplicate(out.FilePath, in.FilePath)
				if err != nil {
					return fmt.Errorf("error generating duplicate files %w", err)
				}
				if isDup {
					dupeList := f.ActualDuplicates[out.MD5Hash]
					dupeList = append(dupeList, out)
					if len(dupeList) == 1 {
						dupeList = append(dupeList, in)
					}
					f.ActualDuplicates[out.MD5Hash] = dupeList
				}
			}
		}
	}
	return nil
}

func IsFileDuplicate(filePathOne, filePathTwo string) (bool, error) {
	fileOne, err := os.Open(filePathOne)
	if err != nil {
		return false, fmt.Errorf("error opening %s: %w", filePathOne, err)
	}
	defer fileOne.Close()
	fileTwo, err := os.Open(filePathTwo)
	if err != nil {
		return false, fmt.Errorf("error opening %s: %w", filePathTwo, err)
	}
	defer fileTwo.Close()

	bArrOne := make([]byte, 1024)
	bArrTwo := make([]byte, 1024)

	var readOneErr, readTwoErr error
	var nOne, nTwo int
	for !errors.Is(readOneErr, io.EOF) && !errors.Is(readTwoErr, io.EOF) {
		nOne, readOneErr = fileOne.Read(bArrOne)
		nTwo, readTwoErr = fileTwo.Read(bArrTwo)
		if nOne != nTwo || !bytes.Equal(bArrOne, bArrTwo) {
			return false, nil
		}
	}
	if errors.Is(readOneErr, io.EOF) && errors.Is(readTwoErr, io.EOF) {
		return true, nil
	}

	return false, nil
}

func (f *FileManager) addFileToDuplicates(h *File) {
	f.fileLock.Lock()
	defer f.fileLock.Unlock()
	f.ActualDuplicates[h.MD5Hash] = append(f.ActualDuplicates[h.MD5Hash], h)
}
