# DirectAdmin Go SDK

Interface with a DirectAdmin installation using Go.

This library supports both the legacy/default DirectAdmin API, and their new modern API still in active development.

**Note: This is in an experimental state. While it's being used in production, the library is very likely to change (
especially in-line with DA's own changes). DA features are being added as needed on our end, but PRs are always welcome!
**

**If you wonder why something has been handled unusually, it's most likely a workaround required by one of DA's many 
quirks.**

## Login as Admin / Reseller / User

To open a session as an admin/reseller/user, follow the following code block:

```go
package main

import (
	"time"
	
	"github.com/levelzerotechnology/directadmin-go"
)

func main() {
	api, err := directadmin.New("https://your.da.address:2222", 5*time.Second)
	if err != nil {
		panic(err)
	}

	userCtx, err := api.LoginAsUser("your_username", "some_password_or_key")
	if err != nil {
		panic(err)
	}

	usage, err := userCtx.GetMyUserUsage()
	if err != nil {
		panic(err)
	}

	userCtx.User.Usage = *usage
}
```

From here, you can call user functions via `userCtx`.

For example, if you wanted to print each of your databases to your terminal:

```go
dbs, err := userCtx.GetDatabases()
if err != nil {
log.Fatalln(err)
}

for _, db := range dbs {
fmt.Println(db.Name)
}
```

## Roadmap

- [ ] Cleanup repo structure (e.g. redis actions being within `admin.go` could go into a dedicated `redis.go` file
  perhaps)
- [ ] Explore DA's new API's update versions of old functions (e.g. user config/usage)
- [ ] Implement testing for all functions
- [ ] Reach stable v1.0

## License

BSD licensed. See the [LICENSE](LICENSE) file for details.