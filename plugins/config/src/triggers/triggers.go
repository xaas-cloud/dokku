package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/config"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	global := flag.Bool("global", false, "--global: Whether global or app-specific")
	flag.Parse()

	var err error
	switch trigger {
	case "config-export":
		appName := flag.Arg(0)
		global := flag.Arg(1)
		merged := flag.Arg(2)
		format := flag.Arg(3)
		err = config.TriggerConfigExport(appName, global, merged, format)
	case "config-get":
		appName := flag.Arg(0)
		key := flag.Arg(1)
		if *global {
			appName = "--global"
			key = flag.Arg(0)
		}
		err = config.TriggerConfigGet(appName, key)
	case "config-get-global":
		key := flag.Arg(0)
		err = config.TriggerConfigGetGlobal(key)
	case "config-unset":
		appName := flag.Arg(0)
		key := flag.Arg(1)
		restart := flag.Arg(2)
		if *global {
			appName = "--global"
			key = flag.Arg(0)
			restart = flag.Arg(1)
		}
		err = config.TriggerConfigUnset(appName, key, common.ToBool(restart))
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = config.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = config.TriggerPostAppRenameSetup(oldAppName, newAppName)
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
