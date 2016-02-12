# QMD

QMD (prounounced "command") is a service for executing arbitrary shell scripts over HTTP.
We use it to compile javascript assets and upload them to s3.

*Note: Exposing shell scripts over HTTP might be very dangerous, so be careful! QMD is intended to be used as an internal service, running in a non-privileged Docker container.*

# Usage

```bash
sudo docker run qmd -config /etc/qmd.conf
```

* `qmd.conf` *see [example file](./etc/qmd.conf.sample)*
* `scripts` directory *where QMD looks for shell scripts to run, see [examples](examples)*

# REST API

### Create QMD job - Execute a script

```
POST /scripts/:filename
```

Request params (JSON):

* `callback_url`:  (optional) execute the script in the background and send the output to the callback_url when the script finishes
* `args`: array of command line arguments to pass to the script upon execution
* `files`: JSON object containing filename : filedata pairs which are saved in $QMD_TMP for the script to use

Response (JSON):

* `id`: unique job ID
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

### List QMD jobs and their state

```
GET /jobs/
```

# Notes

* Scripts will have access to the following environment variables
  * `QMD_TMP`: the directory where your script is being run. All files here will be deleted unless `keepTemp` is set to `true` in your config. Located at `workingDir/tmp/:id`.
  * `QMD_STORE`: the directory set in your config as `storeDir`. All files written here will be left alone.
  * `QMD_OUT`: the output file for your script. Everything written here will be persisted to the response/log under the `output` key. Located at `QMD_TMP/qmd.out`.

# Requirements

* [Redis](https://github.com/antirez/redis)
* [Disque](https://github.com/antirez/disque)

```
sudo docker run -d --name redis -v /data/redis:/data --restart=always -p 6379:6379 redis:latest redis-server --appendonly yes
sudo docker run -d --name disque -v /data/disque:/data --restart=always -p 7711:7711 richnorth/disque:latest disque-server --appendonly yes
```

# LICENSE

Licensed under the [MIT License](./LICENSE).
