package pkg

import "github.com/fsnotify/fsnotify"

// Op is an alias for fsnotify.Op
type Op = fsnotify.Op

// Write is the file write event operation
const Write = fsnotify.Write

// WatchFile creates a file watcher for the specified file path
func WatchFile(filename string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(filename); err != nil {
		watcher.Close()
		return nil, err
	}

	return watcher, nil
}
