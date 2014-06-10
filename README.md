# QMD

..prounounced "command", is utility http service for executing arbitary scripts.
We use it to compile javascript assets and upload them to s3. 

# API

## Scripts resource

### Fetch a list of the available scripts

Request:

```
GET /scripts
```

Response (json):

* array of executable scripts from `ScriptDir`

**Example:**

```
GET /scripts
==
["hi.sh", "testing.rb"]
```


### Fetch a script

Request:

```
GET /scripts/:script_id
```

Response (text):

* the raw script data

**Example:**

```
GET /scripts/hi.sh
```
```
#!/bin/sh
echo "some script output $1 $2" > $QMD_OUT
echo "hiii this is part of the exec log"
```


### Execute a script

```
POST /scripts/:script_id
```

Request params (json):

* `callback_url`:  (optional) execute the script in the background and send the output to the callback_url when the script finishes
* `args`: array of command line arguments to pass to the script upon execution


Response (json):

* `exec_id`: an auto-incrementing number
* `script_id`: the filename in the scripts directory
* `callback_url`: an endpoint to send the output
* `exec_log`: the piped STDOUT and STDERR script execution log
* `output`: the $QMD_OUT output
* `start_time`: the time (in local system time) the script began to execute
* `elapsed_usec`: the amount of time in microseconds to run the script
* `status`: the exit status of the script; either OK or ERR

**Example: Enqueue a script to execute in the background and send output to a callback URL**

```
POST /scripts/hi.sh
{
 "callback_url": "http://...", args: ["a", "b"]
}
```
```
{
 "exec_id": 1, "script_id": "hi.sh", "args": ["a", "b"], "callback_url": "http://...", 
 "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": null, "status": null
}
```

...the job then runs in the background.. and when finishes it will send the following
response to `callback_url`:

```
{
 "exec_id": 1, "script_id": "hi.sh", "args": ["a", "b"], "callback_url": "http://...",
 "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": 3000, "status": "OK",
 "output": "some script output a b", "exec_log": "hiii this is part of the exec log"
}
```

**Example: Execute a script synchronously and return the output**

NOTE: this currently isn't implemented.. all script executions happen asynchronously
and require a callback

```
POST /scripts/hi.sh
{
 "args": ["a"]
}
```
```
{
 "exec_id": 2, "script_id": "hi.sh", "args": ["a"],
 "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": 10000, "status": "OK",
 "output": "some script output a", "exec_log": "hiii this is part of the exec log"
}
```


## Fetch the exec log for a script

Request:

```
GET /scripts/:script_id/log/:exec_id
```

Params:

* none for now.. we can add a cursor some day

Response (json):

* array of `exec` objects as would be received from a completed job. Descending exec_id order.

**Example:**

```
GET /scripts/hi.sh/log
```
```
[{
 "exec_id": 2, "script_id": "hi.sh", "args": ["a"],
 "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": 10000, "status": "OK",
 "output": "some script output a", "exec_log": "hiii this is part of the exec log"
},
{
 "exec_id": 1, "script_id": "hi.sh", "args": ["a", "b"], "callback_url": "http://...",
 "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": 3000, "status": "OK",
 "output": "some script output a b", "exec_log": "hiii this is part of the exec log"
}]
```

# Security Note

QMD exposes the command line over http, so be careful! It's intended to be used as an
internal service. Even then authentication is required and the daemon should be run as
in normal user mode (duh).


# Design thoughts

1. How should we handle retrying of jobs, and when?

2. How should we handle multiple script requests with the same parameters? ie. an enqueued script 
   is requested a second time. Consider the desired behaviour and if we need unique-jobs

3. We need a way to persist the script output. Just some simple database.. my suggestion
   is to use redis. It's fast, we can set key expiry and its lightweight. 

4. Should we add script exec priority? 0, 1, 2, 3 .. etc. any number really.. like a z-order. 

5. TODO: require authentication for web service use via http basic auth


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
