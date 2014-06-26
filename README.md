# QMD

..prounounced "command", is utility http service for executing arbitary scripts.
We use it to compile javascript assets and upload them to s3. 


# Requirements

In addition to the QMD daemon's dependencies it also requires the following items to be running:

* Redis
* nsqd
* nsqlookupd *optional but recommended*
* nsqadmin *optional*

For the NSQ binaries please checkout the docs for installation instructions: http://nsq.io/deployment/installing.html


# Getting Started

After pulling the repo and building the QMD binary, you will need to create three things

* `config` file  *see [config.sample.toml](config.sample.toml) for example settings*
* `scripts` directory  *where QMD looks for scripts to run, see [examples](examples)*
* `whitelist` file  *list of runnable scripts should be located in the scripts directory, see [examples/.whitelist](examples/.whitelist)*

Then you will need to start up Redis and NSQ. QMD can then be started with the command
```
qmd -config-file <path to config file>
```


# API

### QMD Admin Panel (nsqadmin)

Request:
```
GET /
```

Response (Text):

* The root path is proxied to the nsqadmin panel.


### Fetch a list of the available scripts

Request:

```
GET /scripts
```

Response (JSON):

* array of executable scripts from `ScriptDir`

**Example:**

```
GET /scripts
==
["hi.sh", "testing.rb"]
```


### Reload the whitelist of scripts

Request:

```
PUT /scripts
```

Response (Text):

* `200 OK` status and the text `Reload request sent`


### Execute a script

```
POST /scripts/:script_id
```

Request params (JSON):

* `callback_url`:  (optional) execute the script in the background and send the output to the callback_url when the script finishes
* `args`: array of command line arguments to pass to the script upon execution
* `files`: JSON object containing filename : filedata pairs which are saved in $QMD_TMP for the script to use


Response (JSON):

* `id`: an auto-incrementing number
* `script`: the filename in the scripts directory
* `args`: the user given arguments if any
* `files`: the user given files if any
* `callback_url`: an endpoint to send the output
* `output`: the $QMD_OUT output
* `exec_log`: the piped STDOUT and STDERR script execution log
* `status`: the exit status of the script; either OK or ERR
* `start_time`: the time (in local system time) the script began to execute
* `end_time`: the time (in local system time) the script finished executing
* `duration`: the amount of time taken to run the script in seconds as a string

**Example: Enqueue a script to execute in the background and send output to a callback URL**

```
POST /scripts/bench.sh
{
    "callback_url": "http://...", 
    "args": ["1", "20"],
}
```
```
{
    "id": 1,
    "script": "bench.sh",
    "args": [
        "1",
        "20"
    ],
    "callback_url": "http://..."
}
```

...the job then runs in the background.. and when finishes it will send the following
response to `callback_url`:

```
{
    "id": 1,
    "script": "bench.sh",
    "args": [
        "1",
        "20"
    ],
    "callback_url": "http://...",
    "output": "",
    "exec_log": "Running under PID #16231\nMaking file in tmp dir at /home/vagrant/test/tmp/883\n/home/vagrant/test/scripts/bench.sh: line 10: Test #$PID: command not found\nMaking file in store dir at /home/vagrant/test/store\n/home/vagrant/test/scripts/bench.sh: line 13: Test #$PID: command not found\nRandomly selected 2\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 12\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 3\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 20\nAbandon ship! Abandon ship! Everyone for themselves!\n\nsignal: hangup",
    "status": "ERR",
    "start_time": "2014-06-24T17:26:39.643458173Z",
    "end_time": "2014-06-24T17:26:56.666502567Z",
    "duration": "17.023044394"
}
```


## Fetch all exec logs for a script

Request:

```
GET /scripts/:script_id/logs
```

Response (JSON):

* array of `exec log` objects as would be received from a completed job. In descending exec_id order.
* Maximum of 50 items as we trim the logs in Redis.

**Example:**

```
GET /scripts/hi.sh/logs
```
```
[
    {
        "id": 50,
        "script": "bench.sh",
        "args": [
            "1",
            "20"
        ],
        "callback_url": "http://...",
        "output": "",
        "exec_log": "Running under PID #16231\nMaking file in tmp dir at /home/vagrant/test/tmp/883\n/home/vagrant/test/scripts/bench.sh: line 10: Test #$PID: command not found\nMaking file in store dir at /home/vagrant/test/store\n/home/vagrant/test/scripts/bench.sh: line 13: Test #$PID: command not found\nRandomly selected 2\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 12\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 3\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 20\nAbandon ship! Abandon ship! Everyone for themselves!\n\nsignal: hangup",
        "status": "ERR",
        "start_time": "2014-06-24T17:26:39.643458173Z",
        "end_time": "2014-06-24T17:26:56.666502567Z",
        "duration": "17.023044394"
    },
    {
        "id": 49,
        "script": "bench.sh",
        "args": [
            "1",
            "20"
        ],
        "callback_url": "http://...",
        "output": "",
        "exec_log": "Running under PID #16231\nMaking file in tmp dir at /home/vagrant/test/tmp/883\n/home/vagrant/test/scripts/bench.sh: line 10: Test #$PID: command not found\nMaking file in store dir at /home/vagrant/test/store\n/home/vagrant/test/scripts/bench.sh: line 13: Test #$PID: command not found\nRandomly selected 2\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 12\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 3\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 20\nAbandon ship! Abandon ship! Everyone for themselves!\n\nsignal: hangup",
        "status": "ERR",
        "start_time": "2014-06-24T17:26:39.643458173Z",
        "end_time": "2014-06-24T17:26:56.666502567Z",
        "duration": "17.023044394s"
    },
    ...
]
```


## Fetch a specific exec log for a script

Request:

```
GET /scripts/:script_id/logs/:id
```

Response (JSON):

* a JSON object containing all the exec log information

**Example:**

```
GET /scripts/hi.sh/logs/49
```
```
{
    "id": 49,
    "script": "bench.sh",
    "args": [
        "1",
        "20"
    ],
    "callback_url": "http://...",
    "output": "",
    "exec_log": "Running under PID #16231\nMaking file in tmp dir at /home/vagrant/test/tmp/883\n/home/vagrant/test/scripts/bench.sh: line 10: Test #$PID: command not found\nMaking file in store dir at /home/vagrant/test/store\n/home/vagrant/test/scripts/bench.sh: line 13: Test #$PID: command not found\nRandomly selected 2\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 12\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 3\nzzzzzzzzzzzzzzzzz\nzzzzzz..I'M AWAKE\nRandomly selected 20\nAbandon ship! Abandon ship! Everyone for themselves!\n\nsignal: hangup",
    "status": "ERR",
    "start_time": "2014-06-24T17:26:39.643458173Z",
    "end_time": "2014-06-24T17:26:56.666502567Z",
    "duration": "17.023044394"
}
```

# Notes

* QMD exposes the shell over HTTP, this is very very dangerous so be careful! It's intended to be used as an internal service. Even then authentication should be used and the daemon should be run as in normal user mode.

* Scripts will have access to the following environment variables
  * `QMD_TMP`: the directory where your script is being run. All files here will be deleted unless `keepTemp` is set to `true` in your config. Located at `workingDir/tmp/:id`.
  * `QMD_STORE`: the directory set in your config as `storeDir`. All files written here will be left alone.
  * `QMD_OUT`: the output file for your script. Everything written here will be persisted to the response/log under the `output` key. Located at `QMD_TMP/qmd.out`.


# Authors

Created by David Kua (dkua) and Peter Kieltyka (pkieltyka) at [Pressly Inc.](http://www.pressly.com/)
