package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	"github.com/dokku/dokku/plugins/cron"
)

// main entrypoint to all triggers
func main() {
	parts := strings.Split(os.Args[0], "/")
	trigger := parts[len(parts)-1]
	global := flag.Bool("global", false, "--global: Whether global or app-specific")
	flag.Parse()

	var err error
	switch trigger {
	case "cron-get-property":
		appName := flag.Arg(0)
		property := flag.Arg(1)
		if *global {
			appName = "--global"
			property = flag.Arg(0)
		}
		err = cron.TriggerCronGetProperty(appName, property)
	case "install":
		err = cron.TriggerInstall()
	case "post-app-clone-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = cron.TriggerPostAppCloneSetup(oldAppName, newAppName)
	case "post-app-rename-setup":
		oldAppName := flag.Arg(0)
		newAppName := flag.Arg(1)
		err = cron.TriggerPostAppRenameSetup(oldAppName, newAppName)
	case "post-delete":
		appName := flag.Arg(0)
		err = cron.TriggerPostDelete(appName)
	case "report":
		appName := flag.Arg(0)
		err = cron.ReportSingleApp(appName, "", "")
	case "scheduler-stop":
		scheduler := flag.Arg(0)
		appName := flag.Arg(1)
		removeContainers := flag.Arg(2)
		err = cron.TriggerSchedulerStop(scheduler, appName, removeContainers)
	default:
		err = fmt.Errorf("Invalid plugin trigger call: %s", trigger)
	}

	if err != nil {
		common.LogFailWithError(err)
	}
}
