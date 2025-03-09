# Process Management

> [!IMPORTANT]
> New as of 0.3.14, Enhanced in 0.7.0

```
ps:inspect <app>                                                  # Displays a sanitized version of docker inspect for an app
ps:rebuild [--parallel count] [--all|<app>]                       # Rebuilds an app from source
ps:report [<app>] [<flag>]                                        # Displays a process report for one or more apps
ps:restart [--parallel count] [--all|<app>]  [<process-name>]     # Restart an app
ps:restore [<app>]                                                # Start previously running apps e.g. after reboot
ps:scale [--skip-deploy] <app> <proc>=<count> [<proc>=<count>...] # Get/Set how many instances of a given process to run
ps:set <app> <key> <value>                                        # Set or clear a ps property for an app
ps:start [--parallel count] [--all|<app>]                         # Start an app
ps:stop [--parallel count] [--all|<app>]                          # Stop an app
```

## Usage

### Inspecting app containers

> [!IMPORTANT]
> New as of 0.13.0

A common administrative task to perform is calling `docker inspect` on the containers that are running for an app. This can be an error-prone task to perform, and may also reveal sensitive environment variables if not done correctly. Dokku provides a wrapper around this command via the `ps:inspect` command:

```shell
dokku ps:inspect node-js-app
```

This command will gather all the running container IDs for your app and call `docker inspect`, sanitizing the output data so it can be copy-pasted elsewhere safely.

### Rebuilding apps

It may be useful to rebuild an app at will, such as for commands that do not rebuild an app or when skipping a rebuild after setting multiple config values. For these use cases, the `ps:rebuild` function can be used.

```shell
dokku ps:rebuild node-js-app
```

All apps may be rebuilt by using the `--all` flag.

```shell
dokku ps:rebuild --all
```

By default, rebuilding all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku ps:rebuild --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku ps:rebuild --all --parallel -1
```

A missing linked container will result in failure to boot apps. Services should all be started for apps being rebuilt.

### Restarting apps

An app may be restarted using the `ps:restart` command.

```shell
dokku ps:restart node-js-app
```

A single process type - such as `web` or `worker` - may also be specified. This _does not_ support specifying a given instance of a process type, and only supports restarting all instances of that process type.

```shell
dokku ps:restart node-js-app web
```

All apps may be restarted by using the `--all` flag. This flag is incompatible with specifying a process type.

```shell
dokku ps:restart --all
```

By default, restarting all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku ps:restart --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku ps:restart --all --parallel -1
```

A missing linked container will result in failure to boot apps. Services should all be started for apps being rebuilt.

### Displaying existing scale properties

Issuing the `ps:scale` command with no arguments will output the current scaling properties for an app.

```shell
dokku ps:scale node-js-app
```

```
-----> Scaling for python
proctype: qty
--------: ---
web:  1
```

### Changing process management settings

The `ps` plugin provides a number of settings that can be used to managed deployments on a per-app basis. The following table outlines ones not covered elsewhere:

| Name                  | Description                                       | Global Default     |
|-----------------------|---------------------------------------------------|--------------------|
| `stop-timeout`        | Configurable grace period given to the `docker stop` command. If a container has not stopped by this time, a `kill -9` signal or equivalent is sent in order to force-terminate the container. Both the `ps:stop` and `apps:destroy` commands *also* respect this value. If not specified, the Docker defaults for the [`docker stop` command](https://docs.docker.com/engine/reference/commandline/stop/) will be used. | `30`             |

All settings can be set via the `scheduler-k3s:set` command. Using `stop-timeout` as an example:

```shell
dokku scheduler-k3s:set node-js-app stop-timeout 60
```

The default value may be set by passing an empty value for the option in question:

```shell
dokku scheduler-k3s:set node-js-app stop-timeout
```

Properties can also be set globally. If not set for an app, the global value will apply.

```shell
dokku scheduler-k3s:set --global stop-timeout 60
```

The global default value may be set by passing an empty value for the option.

```shell
dokku scheduler-k3s:set --global stop-timeout
```


### Defining Processes

#### Procfile

> [!NOTE]
> Dokku supports the Procfile format as defined in [this document](https://github.com/dokku/procfile-util/blob/master/PROCFILE_FORMAT.md) under "Strict Mode" parsing rules.

Apps can define processes to run by using a `Procfile`. A `Procfile` is a simple text file that can be used to specify multiple commands, each of which is subject to process scaling. In the case where the built image sets a default command to run - either through usage of `CMD` for Dockerfile-based builds, a default process for buildpack-based builds, or any other method for the builder in use - the `Procfile` will take precedence.

If the file exists, it should not be empty, as doing so may result in a failed deploy.

The syntax for declaring a `Procfile` is as follows. Note that the format is one process type per line, with no duplicate process types.

```Procfile
<process type>: <command>
```

If, for example, you have multiple queue workers and wish to scale them separately, the following would be a valid way to work around the requirement of not duplicating process types:

```Procfile
worker:           env QUEUE=* bundle exec rake resque:work
importantworker:  env QUEUE=important bundle exec rake resque:work
```

If the app build declares an `ENTRYPOINT`, the command defined in the `Procfile` is passed as an argument to that entrypoint. This is the case for all Dockerfile-based, Docker Image, and Cloud Native Buildpack deployments.

The `web` process type holds some significance in that it is the only process type that is automatically scaled to `1` on the initial application deploy. See the [web process scaling documentation](/docs/processes/process-management.md#the-web-process) for more details around scaling individual processes.

See the [Procfile location documentation](/docs/processes/process-management.md#changing-the-procfile-location) for more information on where to place your `Procfile` file.

#### The `web` process

For initial app deploys, Dokku will default to starting a single `web` process for each app. This process may be defined within the `Procfile` or as the `CMD` (for Dockerfile or Docker image deploys). Scaling of the `web` process - and all other processes - may be managed via `ps:scale` or the `formation` key in the `app.json` file either before or after the initial deploy.

There are also a few other exceptions for the `web` process.

- By default, the built-in nginx proxy implementation only proxies the `web` process (others may be handled via a custom `nginx.conf.sigil`).
    - See the [nginx request proxying documentation](/docs/networking/proxies/nginx.md#request-proxying) for more information on how nginx handles proxied requests.
- Only the `web` process may be bound to an external port.

#### The `release` process

The `Procfile` also supports a special `release` command which acts in a similar way to the [Heroku Release Phase](https://devcenter.heroku.com/articles/release-phase). See the [Release deployment task documentation](/docs/advanced-usage/deployment-tasks.md#procfile-release-command) for more information on how Dokku handles this process type.

#### Changing the `Procfile` location

The `Procfile` is expected to be found in a specific directory, depending on the deploy approach:

- The `WORKDIR` of the Docker image for deploys resulting from `git:from-image` and `git:load-image` commands.
- The root of the source code tree for all other deploys (git push, `git:from-archive`, `git:sync`).

Sometimes it may be desirable to set a different path for a given app, e.g. when deploying from a monorepo. This can be done via the `procfile-path` property:

```shell
dokku ps:set node-js-app procfile-path .dokku/Procfile
```

The value is the path to the desired file *relative* to the base search directory, and will never be treated as absolute paths in any context. If that file does not exist within the repository, Dokku will continue the build process as if the repository has no `Procfile`.

The default value may be set by passing an empty value for the option:

```shell
dokku ps:set node-js-app procfile-path
```

The `procfile-path` property can also be set globally. The global default is `Procfile`, and the global value is used when no app-specific value is set.

```shell
dokku ps:set --global procfile-path global-Procfile
```

The default value may be set by passing an empty value for the option.

```shell
dokku ps:set --global procfile-path
```

### Scaling apps

#### Via CLI

> This functionality is disabled if the formation is managed via the `formation` key of `app.json`.

Dokku can also manage scaling itself via the `ps:scale` command. This command can be used to scale multiple process types at the same time.

```shell
dokku ps:scale node-js-app web=1
```

Multiple process types can be scaled at once:

```shell
dokku ps:scale node-js-app web=1 worker=1
```

If desired, the corresponding deploy will be skipped by using the `--skip-deploy` flag:

```shell
dokku ps:scale --skip-deploy node-js-app web=1
```

#### Manually managing process scaling

> Using a `formation` key in an `app.json` file with _any_ `quantity` specified disables the ability to use `ps:scale` for scaling. All processes not specified in the `app.json` will have their process count set to zero.

Users can also configure scaling within the codebase itself to manage process scaling. The `formation` key should be specified as follows in the `app.json` file:

```json
{
  "formation": {
    "web": {
      "quantity": 1
    },
    "worker": {
      "quantity": 4
    }
  }
}
```

Removing the `formation` key or removing the `app.json` file from your repository will result in Dokku respecting the `ps:scale` command for setting scale values. The values set via the `app.json` file from a previous deploy will be respected.

See the [app.json location documentation](/docs/advanced-usage/deployment-tasks.md#changing-the-appjson-location) for more information on where to place your `app.json` file.

### Stopping apps

Deployed apps can be stopped using the `ps:stop` command. This turns off all running containers for an app, and will result in a **502 Bad Gateway** response for the default nginx proxy implementation.

```shell
dokku ps:stop node-js-app
```

All apps may be stopped by using the `--all` flag.

```shell
dokku ps:stop --all
```

By default, stopping all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku ps:stop --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku ps:stop --all --parallel -1
```

### Starting apps

All stopped containers can be started using the `ps:start` command. This is similar to running `ps:restart`, except no action will be taken if the app containers are running.

```shell
dokku ps:start node-js-app
```

All apps may be started by using the `--all` flag.

```shell
dokku ps:start --all
```

By default, starting all apps happens serially. The parallelism may be controlled by the `--parallel` flag.

```shell
dokku ps:start --all --parallel 2
```

Finally, the number of parallel workers may be automatically set to the number of CPUs available by setting the `--parallel` flag to `-1`

```shell
dokku ps:start --all --parallel -1
```

### Restart policies

> [!IMPORTANT]
> New as of 0.7.0, Command Changed in 0.22.0

By default, Dokku will automatically restart containers that exit with a non-zero status up to 10 times via the [on-failure Docker restart policy](https://docs.docker.com/engine/reference/run/#restart-policies---restart).

#### Setting the restart policy

> A change in the restart policy must be followed by a `ps:rebuild` call.

You can configure this via the `ps:set` command:

```shell
# always restart an exited container
dokku ps:set node-js-app restart-policy always

# never restart an exited container
dokku ps:set node-js-app restart-policy no

# only restart it on Docker restart if it was not manually stopped
dokku ps:set node-js-app restart-policy unless-stopped

# restart only on non-zero exit status
dokku ps:set node-js-app restart-policy on-failure

# restart only on non-zero exit status up to 20 times
dokku ps:set node-js-app restart-policy on-failure:20
```

Restart policies have no bearing on server reboot, and Dokku will always attempt to restart your apps at that point unless they were manually stopped.

Dokku also runs `dokku-event-listener` in the background via the system's init service. This monitors container state, performing the following actions:

- If a web process restarts and it's container IP address changes, the app's proxy configuration will be rebuilt.
- If a process within an app exceeds the restart count, the app will be rebuilt.

### Displaying reports for an app

> [!IMPORTANT]
> New as of 0.12.0

You can get a report about the deployed apps using the `ps:report` command:

```shell
dokku ps:report
```

```
=====> node-js-app ps information
       Deployed:                      false
       Processes:                     0
       Ps can scale:                  true
       Ps computed procfile path:     Procfile2
       Ps global procfile path:       Procfile
       Ps restart policy:             on-failure:10
       Ps procfile path:              Procfile2
       Restore:                       true
       Running:                       false
=====> python-sample ps information
       Deployed:                      false
       Processes:                     0
       Ps can scale:                  true
       Ps computed procfile path:     Procfile
       Ps global procfile path:       Procfile
       Ps restart policy:             on-failure:10
       Ps procfile path:
       Restore:                       true
       Running:                       false
=====> ruby-sample ps information
       Deployed:                      false
       Processes:                     0
       Ps can scale:                  true
       Ps computed procfile path:     Procfile
       Ps global procfile path:       Procfile
       Ps restart policy:             on-failure:10
       Ps procfile path:
       Restore:                       true
       Running:                       false
```

You can run the command for a specific app also.

```shell
dokku ps:report node-js-app
```

```
=====> node-js-app ps information
       Deployed:                      false
       Processes:                     0
       Ps can scale:                  true
       Ps restart policy:             on-failure:10
       Restore:                       true
       Running:                       false
```

You can pass flags which will output only the value of the specific information you want. For example:

```shell
dokku ps:report node-js-app --deployed
```

### Restoring apps after a server reboot

When a server reboots or Docker is restarted/upgraded, Docker may or may not start old app containers automatically, and may in some cases re-assign container IP addresses. To combat this issue, Dokku uses an init process that triggers `dokku ps:restore` after the Docker daemon is detected as starting. When triggered, the `dokku ps:restore` command will serially (one by one) run the following for each:

- Start all linked services.
- Clear generated proxy configuration files.
- Start the app if it has not been manually stopped.
    - If the app containers still exist, they will be started and the generated proxy configuration files will be rebuilt.
    - If any of the app containers are missing, the entire app will be rebuilt.

During this time, requests may route to the incorrect app if the assigned IPs correspond to those for other apps. While dokku makes all efforts to avoid this, there may be a few minutes where urls may route to the wrong app. To avoid this, either use a custom proxy plugin or wait a few minutes until the restoration process is complete.
