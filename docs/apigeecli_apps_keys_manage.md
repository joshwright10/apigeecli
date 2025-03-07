## apigeecli apps keys manage

Approve or revoke a developer app key

### Synopsis

Approve or revoke a developer app key

```
apigeecli apps keys manage [flags]
```

### Options

```
  -x, --action string   Action to perform - revoke or approve (default "revoke")
  -h, --help            help for manage
  -k, --key string      Developer app consumer key
```

### Options inherited from parent commands

```
  -a, --account string   Path Service Account private key in JSON
  -d, --dev string       Developer email
      --disable-check    Disable check for newer versions
  -n, --name string      Developer App name
      --no-output        Disable printing all statements to stdout
  -o, --org string       Apigee organization name
      --print-output     Control printing of info log statements (default true)
  -t, --token string     Google OAuth Token
```

### SEE ALSO

* [apigeecli apps keys](apigeecli_apps_keys.md)	 - Manage developer app keys

###### Auto generated by spf13/cobra on 3-May-2023
