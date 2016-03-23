package conf
import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"path"
	"math"
)

type XConfig struct {
	XMLName    xml.Name    `xml:"config"`
	Server     XServer     `xml:"server"`
	Auth       XAuth       `xml:"auth"`
	Users      []XUser     `xml:"user"`
	Commands   []XCommands `xml:"commands"`
	Files      []XFiles    `xml:"files"`
	Databases  []XDatabase `xml:"database"`
}

type XServer struct {
	Listen string      `xml:"listen"`
}

type XAuth struct {
	Enabled       bool     `xml:"enabled,attr"`
	MaxTimeDelta  uint32   `xml:"maxTimeDelta"`
}

type XUser struct {
	Name      string           `xml:"id,attr"`
	Hosts     []string         `xml:"host"`
	Key       string           `xml:"key"`
	Files     []XUserFiles     `xml:"files"`
	Commands  []XUserCommands  `xml:"commands"`
}

type XCommands struct {
	Name     string      `xml:"id,attr"`
	Commands []XCommand  `xml:"command"`
}

type XCommand struct {
	Name         string  `xml:"id,attr"`
	Lang         string	 `xml:"lang,attr"`
	Code         string  `xml:"code"`
	Timeout      uint32  `xml:"timeout,attr"`
	User         string  `xml:"runas,attr"`
	Lock         XLock   `xml:"lock"`
}

type XDatabase struct {
	Name    string    `xml:"id,attr"`
	Driver  string    `xml:"driver,attr"`
	Dsn     string    `xml:"dsn,attr"`
	Queries []XQuery  `xml:"query"`
}

type XQuery struct {
	Name    string `xml:"id,attr"`
	Sql     string `xml:",chardata"`
}

type XLock struct {
	Name     string  `xml:"id,attr"`
	Timeout  uint    `xml:"timeout,attr"`
	Wait     bool    `xml:"wait,attr"`
}

type XFiles struct {
	Name   string       `xml:"id,attr"`
	Dirs   []XDir       `xml:"dir"`
}

type XDir struct {
	Name      string    `xml:"id,attr"`
	Root      string    `xml:"root"`
	Allows    []string  `xml:"allow"`
	Patterns  []string  `xml:"pattern"`
}

type XUserFiles struct {
	Name   string   `xml:"id,attr"`
}

type XUserCommands struct {
	Name   string   `xml:"id,attr"`
}

func XConfigFromData(data []byte) (*XConfig, error) {
	ret := XConfig{}
	err := xml.Unmarshal(data, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func XConfigFromReader(reader io.Reader) (*XConfig, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return XConfigFromData(data)
}

func XConfigFromFile(path string) (*XConfig, error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return XConfigFromReader(reader)
}

func (conf *XConfig) ToConfig() *Config {
	ret := Config{}
	ret.Server = Server {
		Listen: conf.Server.Listen,
	}
	ret.Auth = Auth {
		Enabled:      conf.Auth.Enabled,
		MaxTimeDelta: conf.Auth.MaxTimeDelta,
	}
	ret.Files = make(map[string]*Files)
	for _, file := range(conf.Files) {
		fname := file.Name
		ret.Files[fname] = &Files{
			Dirs: make(map[string]*Dir),
		}
		for _, dir := range(file.Dirs) {
			dname := dir.Name
			dir := &Dir{
				Root: path.Clean(strings.TrimSpace(dir.Root)),
				Allows: make([]string, 0, 4),
				Patterns: make([]string, 0, 4),
			}
			for _, method := range(dir.Allows) {
				dir.Allows = append(dir.Allows, strings.ToUpper(strings.TrimSpace(method)))
			}
			for _, pattern := range(dir.Patterns) {
				dir.Patterns = append(dir.Patterns, strings.TrimSpace(pattern))
			}
			ret.Files[fname].Dirs[dname] = dir
		}
	}
	ret.Commands = make(map[string]*Commands)
	for _, commands := range(conf.Commands) {
		csname := commands.Name
		ret.Commands[csname] = &Commands{
			Commands: make(map[string]*Command),
		}
		for _, command := range(commands.Commands) {
			cname := command.Name
			if command.Timeout == 0 {
				command.Timeout = math.MaxUint32
			}
			if command.Lock.Timeout == 0 {
				command.Lock.Timeout = math.MaxUint32
			}
			ret.Commands[csname].Commands[cname] = &Command{
				Code: strings.TrimSpace(command.Code),
				Lang: command.Lang,
				User: command.User,
				Timeout: command.Timeout,
				Lock: Lock {
					Name: strings.TrimSpace(command.Lock.Name),
					Timeout: command.Lock.Timeout,
					Wait: command.Lock.Wait,
				},
			}
		}
	}
	ret.Databases = make(map[string]*Database)
	for _, database := range(conf.Databases) {
		dname := database.Name
		ret.Databases[dname] = &Database{
			Dsn: database.Dsn,
			Driver: database.Driver,
			Queries: make(map[string]*Query),
		}
		for _, query := range(database.Queries) {
			ret.Databases[dname].Queries[query.Name] = &Query{ Sql: query.Sql }
		}
	}
	ret.Users = make(map[string]*User)
	for _, user := range(conf.Users) {
		uname := user.Name
		u := &User{
			Key: strings.TrimSpace(user.Key),
			Hosts: make([]string, len(user.Hosts)),
		}
		for j := range(user.Hosts) {
			u.Hosts[j] = strings.TrimSpace(user.Hosts[j])
		}
		u.Commands = make([]string, 0, 2)
		u.Files = make([]string, 0, 2)
		for _, command := range(user.Commands) {
			u.Commands = append(u.Commands, command.Name)
		}
		for _, file := range(user.Files) {
			u.Files = append(u.Files, file.Name)
		}
		ret.Users[uname] = u
	}
	return &ret
}

