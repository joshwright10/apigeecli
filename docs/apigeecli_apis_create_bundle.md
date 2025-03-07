## apigeecli apis create bundle

Creates an API proxy from an Zip or folder

### Synopsis

Creates an API proxy from an Zip or folder

```
apigeecli apis create bundle [flags]
```

### Options

```
  -h, --help                  help for bundle
  -n, --name string           API Proxy name
  -f, --proxy-folder string   Path to the Proxy Bundle; ex: ./test/apiproxy
  -p, --proxy-zip string      Path to the Proxy bundle/zip file
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

* [apigeecli apis create](apigeecli_apis_create.md)	 - Creates an API proxy in an Apigee Org

###### Auto generated by spf13/cobra on 3-May-2023
