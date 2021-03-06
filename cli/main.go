package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	panthalassa "github.com/Bit-Nation/panthalassa"
	km "github.com/Bit-Nation/panthalassa/keyManager"
	ks "github.com/Bit-Nation/panthalassa/keyStore"
	mnemonic "github.com/Bit-Nation/panthalassa/mnemonic"
	profile "github.com/Bit-Nation/panthalassa/profile"
	log "github.com/ipfs/go-log"
	jsonDB "github.com/nanobox-io/golang-scribble"
	uuid "github.com/satori/go.uuid"
	iShell "gopkg.in/abiosoft/ishell.v2"
)

const DevRendezvousKey = "akhgp58izorhalsdipfo3uh5orpawoudshfalskduf43topa"
const LogFile = "log.out"
const DBName = ".database"

var logger = log.Logger("cli")

type Account struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AccountStore string `json:"account_store"`
	Profile      string `json:"profile"`
}

func (a Account) String() string {
	return fmt.Sprintf("%s, (%s)", a.Name, a.ID)
}

func main() {

	//Database
	db, err := jsonDB.New(DBName, nil)
	if err != nil {
		panic(err)
	}

	shell := iShell.New()

	var userDB *jsonDB.Driver

	// register command to start panthalassa
	shell.AddCmd(&iShell.Cmd{
		Name: "start",
		Help: "start panthalassa",
		Func: func(c *iShell.Context) {

			// fetch account's
			rawAccounts, err := db.ReadAll("account")
			if err != nil {
				c.Err(err)
				return
			}

			// exit if not enough account's
			if len(rawAccounts) == 0 {
				c.Err(errors.New("please create an account first"))
				return
			}

			accounts := []string{}
			myAccounts := map[int]Account{}

			for k, v := range rawAccounts {

				var acc Account

				if err := json.Unmarshal([]byte(v), &acc); err != nil {
					c.Err(err)
					continue
				}

				accounts = append(accounts, acc.String())
				myAccounts[k] = acc
			}

			choice := c.MultiChoice(accounts, "Chose your account:")

			selectedAccount, exist := myAccounts[choice]
			if !exist {
				c.Err(errors.New("account does not exist"))
				return
			}

			//Ask for password to decrypt account
			c.Print("Please enter your password for this account: ")
			password := c.ReadLine()

			// create account database
			userDB, err = jsonDB.New(filepath.FromSlash(fmt.Sprintf("%s/%s", DBName, selectedAccount.ID)), nil)
			if err != nil {
				c.Err(err)
				return
			}

			config := panthalassa.StartConfig{
				EncryptedKeyManager: selectedAccount.AccountStore,
				SignedProfile:       selectedAccount.Profile,
			}

			rawConfig, err := json.Marshal(config)
			if err != nil {
				c.Err(err)
				return
			}

			err = panthalassa.Start(string(rawConfig), password, &Store{
				Account: selectedAccount,
				DB:      userDB,
			})
			if err != nil {
				c.Err(err)
				return
			}

			c.Println("Started panthalassa")

			//fetch id key
			idPubKey, err := panthalassa.IdentityPublicKey()
			if err != nil {
				c.Err(err)
				return
			}

			c.Println("Your identity is: ", idPubKey)

		},
	})

	// stop panthalassa
	shell.AddCmd(&iShell.Cmd{
		Name: "stop",
		Help: "stop's the current panthalassa instance",
		Func: func(c *iShell.Context) {
			userDB = nil
			err := panthalassa.Stop()
			if err != nil {
				c.Err(err)
				return
			}
			c.Println("stopped panthalassa")
		},
	})

	// display private key
	shell.AddCmd(&iShell.Cmd{
		Name: "eth:private",
		Help: "show's the ethereum private key",
		Func: func(c *iShell.Context) {
			pk, err := panthalassa.EthPrivateKey()
			if err != nil {
				c.Err(err)
				return
			}
			c.Println("your private key is: ", pk)
		},
	})

	// display address
	shell.AddCmd(&iShell.Cmd{
		Name: "eth:address",
		Help: "display ethereum address",
		Func: func(c *iShell.Context) {
			addr, err := panthalassa.EthAddress()
			if err != nil {
				c.Err(err)
				return
			}

			c.Println("your ethereum address is:", addr)
		},
	})

	shell.AddCmd(&iShell.Cmd{
		Name: "log:debug",
		Help: "Enable debug logging",
		Func: func(c *iShell.Context) {
			f, err := os.Create(LogFile)
			if err != nil {
				c.Err(err)
				return
			}
			log.Configure(log.Output(f), log.LevelDebug)
			c.Println("Enabled logging (for debug). Output file: ", f.Name())
		},
	})

	shell.AddCmd(&iShell.Cmd{
		Name: "log:warn",
		Help: "Enable debug logging",
		Func: func(c *iShell.Context) {
			f, err := os.Create(LogFile)
			if err != nil {
				c.Err(err)
				return
			}
			log.Configure(log.Output(f))
			//2 = WARNING
			log.SetAllLoggers(2)
			c.Println("Enabled logging (for warning's). Output file: ", f.Name())
		},
	})

	shell.AddCmd(&iShell.Cmd{
		Name: "log:info",
		Help: "Enable info logging",
		Func: func(c *iShell.Context) {
			f, err := os.Create(LogFile)
			if err != nil {
				c.Err(err)
				return
			}
			log.Configure(log.Output(f))
			//2 = WARNING
			log.SetAllLoggers(4)
			c.Println("Enabled logging (for info's). Output file: ", f.Name())
		},
	})

	shell.AddCmd(&iShell.Cmd{
		Name: "account:new",
		Help: "Create a new Account",
		Func: func(c *iShell.Context) {

			c.ShowPrompt(false)

			// get username
			c.Println("Account name: ")
			accountName := c.ReadLine()

			// get password
			c.Print("Password for account: ")
			password := c.ReadLine()

			// get location
			c.Print("Location: ")
			location := c.ReadLine()

			// get image
			c.Print("Image")
			image := c.ReadLine()

			// create mnemonic
			mne, err := mnemonic.New()
			if err != nil {
				c.Err(err)
				return
			}

			// create key store form mnemonic
			store, err := ks.NewFromMnemonic(mne)
			if err != nil {
				c.Err(err)
				return
			}

			// create key manager from key store
			keyManager := km.CreateFromKeyStore(store)
			exportedAccount, err := keyManager.Export(password, password)
			if err != nil {
				c.Err(err)
				return
			}

			// create profile
			p, err := profile.SignWithKeyManagerStore(accountName, location, image, exportedAccount, password)
			if err != nil {
				c.Err(err)
				return
			}
			rawProfile, err := json.Marshal(p)
			if err != nil {
				c.Err(err)
				return
			}

			// uuid
			id, err := uuid.NewV4()
			if err != nil {
				c.Err(err)
				return
			}

			rawStore, err := exportedAccount.Marshal()
			if err != nil {
				panic(err)
			}

			err = db.Write("account", id.String(), &Account{
				ID:           id.String(),
				Name:         accountName,
				AccountStore: string(rawStore),
				Profile:      string(rawProfile),
			})

			c.ShowPrompt(true)

			if err != nil {
				c.Err(err)
				return
			}

			c.Println("safed account store")

		},
	})

	shell.AddCmd(&iShell.Cmd{
		Name: "contact:add",
		Help: "adds a contact to your database",
		Func: func(c *iShell.Context) {

			if userDB == nil {
				c.Err(errors.New("you need to start panthalassa first"))
				return
			}

			c.Print("Enter the public key of your contract: ")
			pubKey := c.ReadLine()

			err := userDB.Write("contact", pubKey, pubKey)
			if err != nil {
				c.Err(err)
				return
			}
			c.Println("Safed friend with public key: ", pubKey)
		},
	})

	shell.AddCmd(&iShell.Cmd{
		Name: "contact:list",
		Help: "list your contacts",
		Func: func(c *iShell.Context) {

			if userDB == nil {
				c.Err(errors.New("you need to start panthalassa first"))
				return
			}

			contacts, err := userDB.ReadAll("contact")
			if err != nil {
				c.Err(err)
			}

			c.Println("your contract's: ")

			for _, contact := range contacts {
				c.Println(contact)
			}

		},
	})

	// run shell
	shell.Run()
}
