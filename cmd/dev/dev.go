// ⚠ Spaghetti warning ⚠

package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

func run(ctx context.Context, argv []string) {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	_, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed creating stdin pipe for %s: %v\n", argv[0], err)
	}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
	}
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command %s failed: %v\n", argv[0], err)
	}
}

func webCmd(ctx context.Context) func() {
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/web")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Starting web process failed: %v\n", err)
	}
	return func() {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGINT)
		cmd.Wait()
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tailwind := "tailwindcss"
	if t, ok := os.LookupEnv("TAILWIND_PATH"); ok {
		tailwind = t
	}
	for _, cmd := range [][]string{
		{tailwind, "-i", "style.css", "-o", "./services/static/public/style.dist.css"},
	} {
		go run(ctx, cmd)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	restartMain := make(chan struct{}, 32)
	go func() {
		cancelMain := webCmd(ctx)
		for {
			select {
			case <-restartMain:
				cancelMain()
				time.Sleep(100 * time.Millisecond)
				cleared := false
				for !cleared {
					select {
					case _ = <-restartMain:
					default:
						cleared = true
					}
				}
				cancelMain = webCmd(ctx)
				time.Sleep(time.Second)
			case <-ctx.Done():
				cancelMain()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// We care about all events except this one:
				if event.Op & ^fsnotify.Chmod == 0 {
					break
				}
				ending := event.Name[strings.LastIndexByte(event.Name, '.')+1:]
				switch ending {
				case "templ", "sql":
					run(ctx, []string{"go", "generate", "./..."})
				case "go", "css":
					restartMain <- struct{}{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				panic(err)
			}
		}
	}()

	watchDirs := make(map[string]struct{})
	if err := fs.WalkDir(
		os.DirFS("."),
		".",
		func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				return nil
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				panic(err)
			}
			for _, entry := range entries {
				name := entry.Name()
				ending := name[strings.LastIndexByte(name, '.')+1:]
				if slices.Contains([]string{"go", "templ", "sql"}, ending) {
					watchDirs[path] = struct{}{}
				}
			}
			return nil
		},
	); err != nil {
		panic(err)
	}
	fmt.Println("Watching:")
	var watchDirsSlice []string
	for dir := range watchDirs {
		watchDirsSlice = append(watchDirsSlice, dir)
	}
	slices.Sort(watchDirsSlice)
	for _, dir := range watchDirsSlice {
		fmt.Println("  ", dir)
		err := watcher.Add(dir)
		if err != nil {
			panic(err)
		}
	}

	<-ctx.Done()
}
