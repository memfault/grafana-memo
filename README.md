# Memo

inspired by https://github.com/Dieterbe/anthracite/ but now
as part of your chatops, chatdev, chatmarketing, chatWhatever workflow!

# Message format:
```
memo [timespec] <msg> [tags]
```

`[foo]` denotes that `foo` is optional.


## timespec

defaults to `25`, so by default it assumes your message is about 25 seconds after the actual event happened.

It can have the following formats:

* `<duration>` like 0 (seconds), 10 (seconds), 30s, 1min20s, 2h, etc. see https://github.com/raintank/raintank-metric/blob/master/dur/durations.go denotes how long ago the event took place
* `<RFC3339 spec>` like `2013-06-05T14:10:43Z`

## msg
free-form text message

# tags

default tags included:
* `chan:slack channel (if not a PM)`
* `author:slack username`

you can extend these. any words at the end of the command that have `:` will be used as key-value tags.
But you cannot override any of the default tags
