package log

import (
	"os"
)

var (
	Level           int
	PrintStackLevel int
	App             *Logger
	RunMode         string
	Folder          string
	pool            loggers
)

type loggers map[string]*Logger

func init() {
	pool = make(map[string]*Logger)
	PrintStackLevel = 4
}

func Init(folder string, names []string, runmode string) {
	Folder = folder
	RunMode = runmode
	if _, err := os.Stat(Folder); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(Folder, os.ModeDir)
		}
	}

	App = Add("app")

	for _, name := range names {
		if name == "app" {
			continue
		}
		Add(name)
	}
}

func Get(name string) *Logger {
	return pool[name]
}
func Add(name string) *Logger {
	l := &Logger{}
	pool[name] = l

	l.Init(Folder + name + ".log")
	return l
}
