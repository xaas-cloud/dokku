package scheduler

import (
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerSchedulerDetect outputs a manually selected scheduler for the app
func TriggerSchedulerDetect(appName string) error {
	if appName != "--global" {
		if scheduler := common.PropertyGet("scheduler", appName, "selected"); scheduler != "" {
			fmt.Println(scheduler)
			return nil
		}
	}

	if scheduler := common.PropertyGet("scheduler", "--global", "selected"); scheduler != "" {
		fmt.Println(scheduler)
		return nil
	}

	fmt.Println("docker-local")
	return nil
}

// TriggerInstall runs the install step for the scheduler plugin
func TriggerInstall() error {
	if err := common.PropertySetup("scheduler"); err != nil {
		return fmt.Errorf("Unable to install the scheduler plugin: %s", err.Error())
	}

	apps, err := common.UnfilteredDokkuApps()
	if err != nil {
		return nil
	}

	results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
		Trigger: "config-get-global",
		Args:    []string{"DOKKU_SCHEDULER"},
	})
	globalScheduler := results.StdoutContents()
	if globalScheduler != "" {
		common.LogVerboseQuiet(fmt.Sprintf("Setting scheduler property 'selected' to %v", globalScheduler))
		if err := common.PropertyWrite("scheduler", "--global", "selected", globalScheduler); err != nil {
			return err
		}

		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "config-unset",
			Args:    []string{"--global", "DOKKU_SCHEDULER"},
		})
		if err != nil {
			common.LogWarn(err.Error())
		}
	}

	for _, appName := range apps {
		results, _ := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "config-get",
			Args:    []string{appName, "DOKKU_SCHEDULER"},
		})
		scheduler := results.StdoutContents()
		if scheduler == "" {
			continue
		}

		common.LogVerboseQuiet(fmt.Sprintf("Setting %s scheduler property 'selected' to %v", appName, scheduler))
		if err := common.PropertyWrite("scheduler", appName, "selected", scheduler); err != nil {
			return err
		}

		_, err := common.CallPlugnTrigger(common.PlugnTriggerInput{
			Trigger: "config-unset",
			Args:    []string{appName, "DOKKU_SCHEDULER"},
		})
		if err != nil {
			common.LogWarn(err.Error())
		}
	}

	return nil
}

// TriggerPostAppCloneSetup creates new scheduler files
func TriggerPostAppCloneSetup(oldAppName string, newAppName string) error {
	err := common.PropertyClone("scheduler", oldAppName, newAppName)
	if err != nil {
		return err
	}

	return nil
}

// TriggerPostAppRenameSetup renames scheduler files
func TriggerPostAppRenameSetup(oldAppName string, newAppName string) error {
	if err := common.PropertyClone("scheduler", oldAppName, newAppName); err != nil {
		return err
	}

	if err := common.PropertyDestroy("scheduler", oldAppName); err != nil {
		return err
	}

	return nil
}

// TriggerPostDelete destroys the scheduler property for a given app container
func TriggerPostDelete(appName string) error {
	return common.PropertyDestroy("scheduler", appName)
}
