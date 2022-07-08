package main

import (
	"fmt"
	"grepclone/worker"
	"grepclone/worklist"
	"os"
	"path/filepath"
	"sync"
)

func getAllFiles(wl *worklist.Worklist, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("readdir error:", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			nextPath := filepath.Join(path, entry.Name())
			getAllFiles(wl, nextPath)
		} else {
			wl.Add(worklist.NewJob(filepath.Join(path, entry.Name())))
		}
	}
}

func main() {

	var workersWg sync.WaitGroup

	var mu sync.Mutex

	wl := worklist.New(100)

	var results []worker.Result

	numWorkers := 10

	workersWg.Add(1)
	go func() {
		defer workersWg.Done()
		getAllFiles(&wl, os.Args[2])
		wl.Finalize(numWorkers)
	}()

	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			for {
				workEntry := wl.Next()
				if workEntry.Path != "" {
					workerResult := worker.FindInFile(workEntry.Path, os.Args[1])
					if workerResult != nil {
						mu.Lock()
						results = append(results, workerResult.Inner...)
						mu.Unlock()
					}
				} else {
					// When the path is empty, this indicates that there are no more jobs available,
					// so we quit.
					return
				}
			}
		}()
	}

	workersWg.Wait()
	for _, r := range results {
		fmt.Printf("%v[%v]:%v\n", r.Path, r.LineNum, r.Line)
	}
}
