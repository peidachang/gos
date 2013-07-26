package log

import (
	"fmt"
	"io"
	golog "log"
	"os"
	"runtime/debug"
)

const (
	LOG_EMERG   = 0 //* system is unusable */
	LOG_ALERT   = 1 //* action must be taken immediately */
	LOG_CRIT    = 2 //* critical conditions */
	LOG_ERR     = 3 //* error conditions */
	LOG_WARNING = 4 //* warning conditions */
	LOG_NOTICE  = 5 //* normal but significant condition */
	LOG_INFO    = 6 //* informational */
	LOG_DEBUG   = 7 //* debug-level messages */
	LOG_SQL     = 8 //* debug-level messages */
)

type Logger struct {
	*golog.Logger
}

func (this *Logger) Init(file string) {
	var out io.Writer
	if RunMode == "dev" {
		fmt.Println("[development]", file+" log is write in Stdout")
		out = os.Stdout
	} else {
		fmt.Println("[production]", file+" log is write in .log file")
		fmt.Println("[production]", "log level is:", Level)
		var err error
		out, err = os.OpenFile(file, os.O_APPEND|os.O_CREATE, 0)
		if err != nil {
			golog.Fatalln(err)
		}
	}
	this.Logger = golog.New(out, "", 0)
	this.Logger.SetFlags(golog.Ldate | golog.Ltime) // | golog.Lshortfile

}

func (this *Logger) SetPrefix(prefix string) {
	this.Logger.SetPrefix(prefix)
}

func (this *Logger) Write(m ...interface{}) {
	this.Logger.Println(m...)
}

func (this *Logger) Stack() {
	this.Logger.SetPrefix("[stack]")
	this.Write(string(debug.Stack()))
}

func (this *Logger) Emerg(m ...interface{}) {
	if Level < LOG_EMERG {
		return
	}
	this.Logger.SetPrefix("[emerg]")
	this.Write(m...)
}

func (this *Logger) Alert(m ...interface{}) {
	if Level < LOG_ALERT {
		return
	}
	this.Logger.SetPrefix("[alert]")
	this.Write(m...)
}

func (this *Logger) Crit(m ...interface{}) {
	if Level < LOG_CRIT {
		return
	}
	this.Logger.SetPrefix("[crit]")
	this.Write(m...)
}
func (this *Logger) Err(m ...interface{}) {
	this.Error(m...)
}
func (this *Logger) Error(m ...interface{}) {
	if Level < LOG_ERR {
		return
	}
	this.Logger.SetPrefix("[err]")
	this.Write(m...)
}
func (this *Logger) Warn(m ...interface{}) {
	if Level < LOG_WARNING {
		return
	}
	this.Logger.SetPrefix("[warn]")
	this.Write(m...)
}
func (this *Logger) Notice(m ...interface{}) {
	if Level < LOG_NOTICE {
		return
	}
	this.Logger.SetPrefix("[notice]")
	this.Write(m...)
}
func (this *Logger) Info(m ...interface{}) {
	if Level < LOG_INFO {
		return
	}
	this.Logger.SetPrefix("[info]")
	this.Write(m...)
}
func (this *Logger) Debug(m ...interface{}) {
	if Level < LOG_DEBUG {
		return
	}
	this.Logger.SetPrefix("[debug]")
	this.Write(m...)
}
func (this *Logger) Sql(m ...interface{}) {
	if Level < LOG_SQL {
		return
	}
	this.Logger.SetPrefix("[sql]")
	this.Write(m...)
}
