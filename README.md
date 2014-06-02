# QMD

Qmd prounounced "command" is utility http service for executing arbitary scripts.
We use it to compile javascript assets and upload them to s3. 

# API

## Scripts resource

### Fetch a list of the available scripts

Request:

```
GET /scripts
```

Response (json):

* array of executable script files from the script working directory

**Example:**

```
GET /scripts
==
["hi.sh", "testing.rb"]
```


### Fetch a script

Request:

```
GET /scripts/<script_id>
```

Response (text):

* the raw script data

**Example:**

```
GET /scripts/hi.sh
==
#!/bin/sh
echo "hiii"
```


### Execute a script

Request:

```
GET /scripts/<script_id>/exec[?params]
```

Params:

* `callback_url`:  (optional) execute the script in the background and send the output to the callback_url when the script finishes
* ..all other query params will be passed as environment variables to the script

Response (json):

* `exec_id`: an auto-incrementing number
* `script_id`: the filename in the scripts directory
* `vars`: the key/value pairs that will be capitalized and passed
  to the script as environment variables
* `callback_url`: an endpoint to send the output
* `output`: the piped STDOUT and STDERR output of the script
* `start_time`: the time (in local system time) the script began to execute
* `elapsed_usec`: the amount of time in microseconds to run the script
* `status`: the exit status of the script; either OK or ERR

** Example: Enqueue a script to execute in the background and send output to a callback URL **

```
GET /scripts/hi.sh?callback_url=X&param1=a&param2=b
==
{
 "exec_id": 1, "script_id": "hi.sh", "vars": {"PARAM1": "a", "PARAM2": "b"},
 "callback_url": "X", "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": null,
 "status": null
}
```

...the job then runs in the background.. and when finishes it will send the following
response to `callback_url`:

```
{
 "exec_id": 1, "script_id": "hi.sh", "vars": {"PARAM1": "a", "PARAM2": "b"},
 "callback_url": "X", "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": 3000,
 "status": "OK", "output": "the piped STDOUT and STDERR output of script"
}
```

** Example: Execute a script synchronously and return the output **

```
GET /scripts/hi.sh/exec?param1=A
==
{
 "exec_id": 2, "script_id": "hi.sh", "vars": {"PARAM1": "a"},
 "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": 10000,
 "status": "OK", "output": "the piped STDOUT and STDERR output of script"
}
```


## Fetch the exec log for a script

Request:

```
GET /scripts/<script_id>/log
```

Params:

* none for now.. we can add a cursor some day

Response (json):

* array of `exec` objects as would be received from a completed job. Descending exec_id order.

**Example:**

```
GET /scripts/hi.sh/log
==
[{
 "exec_id": 2, "script_id": "hi.sh", "vars": {"PARAM1": "a"},
 "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": 10000,
 "status": "OK", "output": "the piped STDOUT and STDERR output of script"
},
{
 "exec_id": 1, "script_id": "hi.sh", "vars": {"PARAM1": "a", "PARAM2": "b"},
 "callback_url": "X", "start_time": "2014-05-25T16:45:49Z", "elapsed_usec": 3000,
 "statu
}]
```


## (TODO) Registering a new script

Request:

```
POST /scripts
```

Params:

* `id`: the script filename
* `script`: the actual script
* `vars`: a list of accepted environment variables (is this necessary??)

Response (json):

* http status: 200
* `id`: echo back the script filename to acknowledge a successful registration

in case of an error adding a script, instead return:

* http status: 422
* `err`: error message explaining the problem

** Example: **

```
POST /scripts
id=test.sh
script=
#!/bin/sh
echo "hey yo"

==

{"id": "test.sh"}

```

### NOTES

* We have to first consider security implications before adding support for
uploading scripts. The user could really upload anything. Some ideas are to sandbox
the scripts, or make this endpoint optional / at-own-risk.


# Design thoughts

1. How should we handle retrying of jobs, and when?

2. How should we handle multiple script requests with the same parameters? ie. an enqueued script 
   is requested a second time. Consider the desired behaviour and if we need unique-jobs

3. We need a way to persist the script output. Just some simple database.. my suggestion
   is to use redis. It's fast, we can set key expiry and its lightweight. 

4. Should we add script exec priority? 0, 1, 2, 3 .. etc. any number really.. like a z-order. How does
   nsq implement this?
   

# Authors / License

Created by David Kua and Peter Kieltyka at Pressly Inc.

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
