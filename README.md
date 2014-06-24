# QMD

..prounounced "command", is utility http service for executing arbitary scripts.
We use it to compile javascript assets and upload them to s3. 


# Requirements

In addition to the QMD daemon's dependencies it also requires the following items to be running:

* Redis
* nsqd
* nsqlookupd *optional* *recommended*
* nsqadmin *optional*

For the NSQ binaries please checkout the docs for installation instructions: http://nsq.io/deployment/installing.html

# API

## Scripts resource


### QMD Heartbeat

Request:
```
GET /
```

Response (Text):

* `200 OK` status and the text `.`
* To be used as a form of heartbeat check.


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
* `duration`: the amount of time taken to run the script

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
    "duration": "17.023044394s"
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
        "duration": "17.023044394s"
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
    "duration": "17.023044394s"
}
```

# Security Note

QMD exposes the shell over HTTP, so be careful! It's intended to be used as an
internal service. Even then authentication is required and the daemon should be run as
in normal user mode (duh).


# Authors / License

Created by David Kua and Peter Kieltyka at Pressly Inc.

Copyright (c) 2014 Pressly Inc. www.pressly.com

MIT License

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
