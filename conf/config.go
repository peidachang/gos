package conf

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	config   Configuration
	bComment = []byte{'#'}
	bEmpty   = []byte{}
	bEqual   = []byte{'='}
	bDQuote  = []byte{'"'}
	bBracket = []byte{'['}
)

func Load(f string) Configuration {
	c, err := loadConfig(f)
	if err != nil {
		println("can not found config file: " + f + "app.conf")
		os.Exit(1)
	}
	config = c
	return c
}

type Configuration map[string]Conf
type Conf map[string]string

func loadConfig(name string) (Configuration, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := Configuration{}

	buf := bufio.NewReader(file)
	var group string

	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if bytes.Equal(line, bEmpty) {
			continue
		}

		if bytes.HasPrefix(line, bBracket) {
			group = string(line[1 : len(line)-1])
			continue
		}

		if bytes.HasPrefix(line, bComment) {
			continue
		}

		val := bytes.SplitN(line, bEqual, 2)

		if bytes.HasPrefix(val[1], bDQuote) {
			val[1] = bytes.Trim(val[1], `"`)
		}

		key := strings.TrimSpace(string(val[0]))

		if cfg[group] == nil {
			cfg[group] = make(map[string]string)
		}

		cfg[group][key] = strings.TrimSpace(string(val[1]))
	}

	return cfg, nil
}

func (this Configuration) IsSet(key string) bool {
	_, ok := this[key]
	if ok && len(this[key]) > 0 {
		return true
	}
	return false
}

func (this Conf) IsSet(key string) bool {
	_, ok := this[key]
	return ok
}

// Bool returns the boolean value for a given key.
func (this Conf) GetBool(key string) bool {
	value, _ := strconv.ParseBool(this[key])
	return value
}

// Int returns the integer value for a given key.
func (this Conf) GetInt(key string) int {
	value, _ := strconv.Atoi(this[key])
	return value
}

// Float returns the float value for a given key.
func (this Conf) GetFloat(key string) float64 {
	value, _ := strconv.ParseFloat(this[key], 64)
	return value
}

// String returns the string value for a given key.
func (this Conf) GetString(key string) string {
	return this.Get(key)
}

// String returns the string value for a given key.
func (this Conf) Get(key string) string {
	if this.IsSet(key) {
		return this[key]
	} else {
		return ""
	}
}

func (this Configuration) GetRunMode() string {
	var m string
	if this.IsSet("app") && this.IsSet("mode") {
		m = this["app"]["mode"]
		switch m {
		case "production":
			m = "pro"
		case "development":
			m = "dev"
		case "pro":
		case "dev":
		default:
			m = "dev"
		}
	} else {
		m = "dev"
	}

	return m
}
