// Copyright © 2018 EOS Canada <info@eoscanada.com>

package cmd

import (
	"errors"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/system"
	"github.com/spf13/cobra"
)

var systemUpdateauthCmd = &cobra.Command{
	Use:   `updateauth [account] [permission_name] [parent permission or ""] [authority]`,
	Short: "Set or update a permission on an account. See --help for more details.",
	Long: `Set or update a permission on an account.

The [authority] field can be either a *public key* or a path to a YAML
file.

If you specify a public key, a simple 'authority' structure is built,
with a threshold of 1, and a single key.

Otherwise, it should be a path to a YAML file.  Here is a sample YAML
authority file:

---
threshold: 3
keys:
- key: EOS6MRyAjQq8ud7hVNYcfn................tHuGYqET5GDW5CV
  weight: 1
accounts:
- permission:
    actor: accountname
    permission: namedperm
  weight: 1
waits:
- wait_sec: 300
  weight: 1
---

`,
	Args: cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		account := toAccount(args[0], "account")
		permissionName := toName(args[1], "permission_name")

		var parent eos.Name
		if args[2] != "" {
			parent = toName(args[2], "parent permission")
		}
		authParam := args[3]

		var auth eos.Authority
		authKey, err := ecc.NewPublicKey(authParam)
		if err == nil {
			auth = eos.Authority{
				Threshold: 1,
				Keys: []eos.KeyWeight{
					{PublicKey: authKey, Weight: 1},
				},
			}
		} else {
			err := loadYAMLOrJSONFile(authParam, &auth)
			errorCheck("authority file invalid", err)
		}

		err = ValidateAuth(auth)
		errorCheck("authority file invalid", err)


		api := getAPI()

		var updateAuthActionPermission = "active"
		if parent == "" {
			updateAuthActionPermission = "owner"
		}
		pushEOSCActions(api, system.NewUpdateAuth(account, eos.PermissionName(permissionName), eos.PermissionName(parent), auth, eos.PermissionName(updateAuthActionPermission)))
	},
}

func init() {
	systemCmd.AddCommand(systemUpdateauthCmd)
}

func ValidateAuth( auth eos.Authority) error {
	for _,  account := range auth.Accounts {
		if len(account.Permission.Permission) == 0 {
			return errors.New("account is missing permission")
		}
		if len(account.Permission.Actor) == 0 {
			return errors.New("account is missing actor")
		}

		if account.Weight == 0 {
			return errors.New("account is missing weight")
		}
	}

	for _, key := range auth.Keys {
		if len(key.PublicKey.Content) == 0 {
			return errors.New("key is missing its public Key")
		}

		if key.Weight == 0 {
			 return errors.New("key is missing weight")
		}
	}

	for _, wait := range auth.Waits {
		if wait.WaitSec == 0 {
			return errors.New("wait cannot be 0")
		}

		if wait.Weight == 0 {
			return errors.New("wait is missing weight")
		}
	}
	return nil
}