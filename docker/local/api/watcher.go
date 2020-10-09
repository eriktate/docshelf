package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func contains(strings []string, el string) bool {
	for _, str := range strings {
		if str == el {
			return true
		}
	}

	return false
}

func findDirectories(basePath string) ([]string, error) {
	cmd := exec.Command("find", basePath, "-name", "*.go") // nolint: gosec
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	rawFiles := out.String()
	lines := strings.Split(rawFiles, "\n")
	directories := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.Split(line, "/")
		dir := strings.Join(parts[:len(parts)-1], "/")
		if !contains(directories, dir) {
			directories = append(directories, dir)
		}
	}

	return directories, nil
}

func build(mainFile string, reload chan bool) error {
	log.Println("building API")
	cmd := exec.Command("go", "build", "-o", "/opt/app/docshelf", mainFile) // nolint: gosec
	if err := cmd.Run(); err != nil {
		return err
	}

	if reload != nil {
		reload <- true
	}

	return nil
}

func runAPI() (chan bool, error) {
	log.Println("running API")
	reload := make(chan bool)
	cmd := exec.Command("/opt/app/docshelf")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		for {
			<-reload
			log.Println("reloading API")
			if err := cmd.Process.Kill(); err != nil {
				log.Printf("could not kill process: %s", err)
			}
			_ = cmd.Wait()
			cmd = exec.Command("/opt/app/docshelf")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Start(); err != nil {
				log.Printf("could not restart API: %s", err)
			}
		}
	}()

	return reload, nil
}

func run(mainFile string) error {
	if err := build(mainFile, nil); err != nil {
		return fmt.Errorf("failed to complete initial build: %w", err)
	}

	reload, err := runAPI()
	if err != nil {
		return fmt.Errorf("failed to run API: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op == fsnotify.Write {
					// only consider go files
					if strings.HasSuffix(event.Name, ".go") {
						// ignore temporary files
						if event.Name[len(event.Name)-1] != '~' {
							log.Printf("rebuilding from: %v", event)
							if err := build(mainFile, reload); err != nil {
								log.Printf("failed to build: %s", err)
							}
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Error: %s", err)
			}
		}
	}()

	dirs, err := findDirectories("./")
	if err != nil {
		return fmt.Errorf("failed to find directories: %w", err)
	}

	for _, dir := range dirs {
		if err := watcher.Add(dir); err != nil {
			return err
		}
	}
	<-done
	return nil
}

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("no main.go given to build")
	}

	if err := run(args[1]); err != nil {
		log.Fatal(err)
	}
}
