# scoreboard tool

Scoreboard "buckets" HTTP responses based on headers or json fields, then
displays a set of graphs in the terminal showing the distibution of different
values.

It can be useful when altering the behavior of a service to illustrate the
effects. Think of a service that is producing 500 errors, and when fixed, the
responses shift to 200s.

![a demo gif](./assets/demo.gif)

### Options

`--url`
: URL to fetch periodically, and use as the source of input

`--interval`
: delay between URL fetches

`--json`
: Use when url return a json object. The specified field is used to categorize
the response

`--header`
: Uses the specified HTTP response header to categorize the response. Absent
headers will appear as an empty string.

`--status`
: Uses the HTTP response status to bucket the response.
