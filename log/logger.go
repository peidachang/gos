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
	if file == "" || RunMode == "dev" {
		fmt.Println("[development] Logger is write in Stdout ")
		out = os.Stdout
	} else {
		fmt.Println("Logger is write in .log file: " + file)
		fmt.Println("log level is:", Level)
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

func (this *Logger) Write(m ...interface{}) *Logger {
	this.Logger.Println(m...)
	return this
}
func (this *Logger) writelog(loglevel int, m ...interface{}) *Logger {
	if Level < loglevel {
		return this
	}
	this.Write(m...)
	if PrintStackLevel >= loglevel {
		this.Stack()
	}
	return this
}

func (this *Logger) Stack() *Logger {
	this.Logger.SetPrefix("[stack]")
	this.Write(string(debug.Stack()))
	return this
}

func (this *Logger) Emerg(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[emerg]")
	this.writelog(LOG_EMERG, m...)
	this.Write("emergency exit")
	os.Exit(0)
	return this
}

func (this *Logger) Alert(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[alert]")
	return this.writelog(LOG_ALERT, m...)
}

func (this *Logger) Crit(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[crit]")
	return this.writelog(LOG_CRIT, m...)
}
func (this *Logger) Error(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[err]")
	return this.writelog(LOG_ERR, m...)
}
func (this *Logger) Warn(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[warn]")
	return this.writelog(LOG_WARNING, m...)
}
func (this *Logger) Notice(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[notice]")
	return this.writelog(LOG_NOTICE, m...)
}
func (this *Logger) Info(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[info]")
	return this.writelog(LOG_INFO, m...)
}
func (this *Logger) Debug(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[debug]")
	return this.writelog(LOG_DEBUG, m...)
}
func (this *Logger) Sql(m ...interface{}) *Logger {
	this.Logger.SetPrefix("[sql]")
	return this.writelog(LOG_SQL, m...)
}
