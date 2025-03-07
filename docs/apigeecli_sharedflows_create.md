## apigeecli sharedflows create

Creates a sharedflow in an Apigee Org

### Synopsis

Creates a sharedflow in an Apigee Org

```
apigeecli sharedflows create [flags]
```

### Options

```
  -h, --help               help for create
  -n, --name string        Sharedflow name
  -f, --sf-folder string   Path to the Sharedflow Bundle; ex: ./test/sharedflowbundle
  -p, --sf-zip string      Path to the Sharedflow bundle/zip file
```

### Options inherited from parent commands

```
  -a, --account string   Path Service Account private key in JSON
      --disable-check    Disable check for newer versions
      --no-output        Disable printing all statements to stdout
  -o, --org string       Apigee organization name
      --print-output     Control printing of info log statements (default true)
  -t, --token string     Google OAuth Token
```

### SEE ALSO

* [apigeecli sharedflows](apigeecli_sharedflows.md)	 - Manage Apigee shared flows in an org

###### Auto generated by spf13/cobra on 3-May-2023
