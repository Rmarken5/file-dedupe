package hasher

import (
	"bytes"
	"context"
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

func (f *FileManager) Run(ctx context.Context, filePath string) (string, error) {

	err := f.getHashesForDirectory(ctx, filePath)
	if err != nil {
		return "", err
	}
	f.generateHashDuplicates(ctx)

	err = f.generateDuplicateFiles(ctx, f.FileHashDuplicates)
	if err != nil {
		return "", err
	}

	err = f.displayDuplicates(ctx)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (f *FileManager) getHashesForDirectory(ctx context.Context, directoryPath string) error {
	fs, err := os.Open(directoryPath)
	if err != nil {
		return err
	}

	defer func() {
		if tmpErr := fs.Close(); tmpErr != nil {
			err = tmpErr
		}
	}()

	dir, err := fs.ReadDir(0)
	if err != nil {
		return err
	}

	var errs []error
	for _, d := range dir {
		if d.IsDir() {
			err := f.getHashesForDirectory(ctx, fmt.Sprintf("%s/%s", directoryPath, d.Name()))
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

func (f *FileManager) generateHashDuplicates(ctx context.Context) {
	hashes := make([]*File, len(f.Hashes))
	i := 0
	for _, file := range f.Hashes {
		hashes[i] = file
		i++
	}

	var decrementer int
	for i := len(hashes) - 1; i >= 1; i -= decrementer {
		decrementer = 1
		out := hashes[i]
		for k := i - 1; k >= 0; k-- {
			in := hashes[k]
			if out.FilePath != in.FilePath && out.MD5Hash == in.MD5Hash {
				fmt.Printf("out %s is equal to in %s \n", out.MD5Hash, in.MD5Hash)
				if _, ok := f.FileHashDuplicates[out.MD5Hash]; !ok {
					f.addFileHashDuplicates(out)
				}
				f.addFileHashDuplicates(in)
				hashes = removeElementFromSlice(k, hashes)
				decrementer++
			}
		}
	}
}

func removeElementFromSlice[S ~[]*E, E any](i int, s S) []*E {
	return append(s[:i], s[i+1:]...)
}

func (f *FileManager) addFileHashDuplicates(h *File) {
	f.duplicateLock.Lock()
	defer f.duplicateLock.Unlock()
	f.FileHashDuplicates[h.MD5Hash] = append(f.FileHashDuplicates[h.MD5Hash], h)
}

func (f *FileManager) generateDuplicateFiles(ctx context.Context, hashDupes map[string][]*File) error {
	for _, list := range hashDupes {
		var decrementer int
		for i := len(list) - 1; i > 0; i -= decrementer {
			decrementer = 1
			out := list[i]
			for k := i - 1; k >= 0; k-- {
				in := list[k]
				isDup, err := IsFileDuplicate(out.FilePath, in.FilePath)
				if err != nil {
					return fmt.Errorf("error generating duplicate files %w", err)
				}
				if isDup {
					dupeList := f.ActualDuplicates[out.MD5Hash]
					if len(dupeList) == 0 {
						dupeList = append(dupeList, out)
					}
					dupeList = append(dupeList, in)
					f.ActualDuplicates[out.MD5Hash] = dupeList
					list = removeElementFromSlice(k, list)
					decrementer++
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

func (f *FileManager) displayDuplicates(ctx context.Context) error {
	for key, list := range f.ActualDuplicates {
		fmt.Println("The following files are duplicates.")
		for i, file := range list {
			fmt.Printf("%d. %s\n", i+1, file.FilePath)
		}
		inFlag := true
		for inFlag {
			fmt.Println("The choose a file to keep.")
			var n int
			_, err := fmt.Scanf("%d", &n)
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			if n > len(list) || n < 1 {
				fmt.Println("Choose from a number on the list.")
				continue
			}
			cDupes := f.ActualDuplicates[key]
			n--
			fList := append(cDupes[:n], cDupes[n+1:]...)
			var multErr []error
			for _, file := range fList {
				err := file.Delete(ctx)
				if err != nil {
					multErr = append(multErr, err)
				}
			}
			if multErr != nil {
				return fmt.Errorf("error deleting duplicate files: %w", err)
			}
			inFlag = false
		}
	}
	return nil
}
