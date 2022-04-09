package plugins

import (
	"fmt"
	"github.com/5HT2/taro-bot/bot"
	"io/ioutil"
	"log"
	"path/filepath"
	"plugin"
	"strings"
)

type PluginInit struct {
}

type Plugin struct {
	Name        string             // Name of the plugin to display to users
	Description string             // Description of what the plugin does
	Version     string             // Version in semver, e.g.., 1.1.0
	Commands    []bot.CommandInfo  // Commands to register, could be none
	Responses   []bot.ResponseInfo // Responses to register, could be none
	Jobs        []bot.JobInfo      // Jobs to register, could be none
}

func (p Plugin) String() string {
	return fmt.Sprintf("[%s, %s, %v, %s, %s, %s]", p.Name, p.Description, p.Version, p.Commands, p.Responses, p.Jobs)
}

func (p *Plugin) Register() {
	bot.Commands = append(bot.Commands, p.Commands...)
	bot.Responses = append(bot.Responses, p.Responses...)
	bot.Jobs = append(bot.Jobs, p.Jobs...) // these need to have RegisterJobs called in order to function
}

func Load(dir string) {
	d, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("plugin loading failed: couldn't load dir: %s\n", err)
		return
	}

	pluginInit := &PluginInit{}

	for _, entry := range d {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".so") {
			pluginPath := filepath.Join(dir, entry.Name())
			log.Printf("plugin found: %s\n", entry.Name())

			p, err := plugin.Open(pluginPath)
			if err != nil {
				log.Printf("plugin load failed: couldn't open plugin: %s (%s)\n", entry.Name(), err)
				continue
			}

			fn, err := p.Lookup("InitPlugin")
			if err != nil {
				log.Printf("plugin load failed: couldn't lookup symbols: %s (%s)\n", entry.Name(), err)
				continue
			}

			initFn := fn.(func(manager *PluginInit) *Plugin)
			if p := initFn(pluginInit); p != nil {
				p.Register()
				log.Printf("plugin registered: %s\n", p)
			} else {
				log.Printf("plugin load failed: %s (nil)\n", entry.Name())
			}
		}
	}
}
