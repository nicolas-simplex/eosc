// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/dgiagio/getpass"
	"github.com/eoscanada/eos-go/ecc"
	eosvault "github.com/eoscanada/eosc/vault"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var vaultImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a private keys to vault",
	Long: `Import a private keys to vault

A vault contains encrypted private keys, and with 'eosc', can be used to
securely sign transactions.

`,
	Run: func(cmd *cobra.Command, args []string) {
		walletFile := viper.GetString("vault-file")
		if _, err := os.Stat(walletFile); err == nil {
			fmt.Printf("Wallet file %q already exists, rename it before running `eosc vault create`.\n", walletFile)
			os.Exit(1)
		}

		vault := eosvault.NewVault()
		vault.Comment = viper.GetString("comment")

		privateKeys, err := capturePrivateKeys()
		if err != nil {
			fmt.Println("ERROR: importing private key:", err)
			os.Exit(1)
		}

		if err != nil {
			fmt.Println("ERROR: retreiving private key:", err)
			os.Exit(1)
		}

		var newKeys []ecc.PublicKey
		for _, privateKey := range privateKeys {
			vault.AddPrivateKey(privateKey)
			newKeys = append(newKeys, privateKey.PublicKey())
		}

		fmt.Println("Keys imported. Let's secure them before showing the public keys.")

		passphrase, err := getpass.GetPassword("Enter passphrase to encrypt your keys: ")
		if err != nil {
			fmt.Println("ERROR reading password:", err)
			os.Exit(1)
		}

		passphraseConfirm, err := getpass.GetPassword("Confirm passphrase: ")
		if err != nil {
			fmt.Println("ERROR reading confirmation password:", err)
			os.Exit(1)
		}

		if passphrase != passphraseConfirm {
			fmt.Println("ERROR: passphrase mismatch!")
			os.Exit(1)
		}

		err = vault.SealWithPassphrase(passphrase)
		if err != nil {
			fmt.Println("ERROR sealing keys:", err)
			os.Exit(1)
		}

		err = vault.WriteToFile(walletFile)
		if err != nil {
			fmt.Printf("ERROR writing to file %q: %s\n", walletFile, err)
			os.Exit(1)
		}

		fmt.Printf("Wallet file %q created. Here are your public keys:\n", walletFile)
		for _, pub := range newKeys {
			fmt.Printf("- %s\n", pub.String())
		}
	},
}

func init() {
	vaultCmd.AddCommand(vaultImportCmd)
	vaultCreateCmd.Flags().StringP("comment", "c", "", "Label or comment about this key vault")

	for _, flag := range []string{"comment"} {
		if err := viper.BindPFlag(flag, vaultCreateCmd.Flags().Lookup(flag)); err != nil {
			panic(err)
		}
	}

}

func capturePrivateKeys() ([]*ecc.PrivateKey, error) {

	privateKeys, err := capturePrivateKey(true)
	if err != nil {
		return privateKeys, fmt.Errorf("keys capture, %s", err.Error())
	}
	return privateKeys, nil

}
func capturePrivateKey(isFirst bool) (privateKeys []*ecc.PrivateKey, err error) {

	prompt := "Type your first private key : "
	if !isFirst {
		prompt = "Type your next private key or just enter if you are done : "
	}

	enteredKey, err := getpass.GetPassword(prompt)
	if err != nil {
		return privateKeys, fmt.Errorf("key capture, %s", err.Error())
	}

	if enteredKey == "" {
		return privateKeys, nil
	}

	key, err := ecc.NewPrivateKey(enteredKey)
	if err != nil {
		return privateKeys, fmt.Errorf("new private key, %s", err.Error())
	}

	privateKeys = append(privateKeys, key)
	nextPrivateKeys, err := capturePrivateKey(false)
	if err != nil {
		return privateKeys, fmt.Errorf("next capture, %s", err.Error())
	}

	privateKeys = append(privateKeys, nextPrivateKeys...)

	return
}
