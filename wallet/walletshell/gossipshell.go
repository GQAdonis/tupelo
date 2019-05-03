package walletshell

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/ethereum/go-ethereum/crypto"
	cbornode "github.com/ipfs/go-ipld-cbor"
	"github.com/quorumcontrol/tupelo-go-sdk/consensus"
	"github.com/quorumcontrol/tupelo-go-sdk/gossip3/remote"
	"github.com/quorumcontrol/tupelo-go-sdk/gossip3/types"
	"github.com/quorumcontrol/tupelo/wallet/walletrpc"
)

func confirmPassword(c *ishell.Context) (string, error) {
	for tries := 0; tries < 3; tries++ {
		c.Print("Please enter a new passphrase: ")
		passphrase := c.ReadPassword()
		c.Print("Please confirm your passphrase by entering it again: ")
		confirmation := c.ReadPassword()

		if passphrase == confirmation {
			c.Println("Thank you for confirming your password.")
			return passphrase, nil
		} else {
			c.Println("Sorry, the passphrases you entered do not match.")
		}
	}

	c.Println("Sorry, none of the passwords have matched.")
	return "", errors.New("can't confirm password")
}

func RunGossip(name string, storagePath string, notaryGroup *types.NotaryGroup, pubsub remote.PubSub) {
	// by default, new shell includes 'exit', 'help' and 'clear' commands.
	shell := ishell.New()

	// display welcome info.
	shell.Printf("Loading shell for wallet: %v\n", name)

	// load the session
	session, err := walletrpc.NewSession(storagePath, name, notaryGroup, pubsub)
	if err != nil {
		shell.Printf("error loading shell: %v\n", err)
		return
	}

	shell.AddCmd(&ishell.Cmd{
		Name: "create-wallet",
		Help: "create the shell wallet",
		Func: func(c *ishell.Context) {
			c.Println("Creating wallet: ", name)

			passphrase, err := confirmPassword(c)
			if err != nil {
				c.Printf("Error creating wallet: %v\n", err)
				return
			} else {
				if err = session.CreateWallet(passphrase); err != nil {
					log.Printf("failed to create wallet: %s", err)
					// TODO: Enable
					// panic(fmt.Sprintf("failed to create wallet: %s", err))
				}
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "start-session",
		Help: "start a new session",
		Func: func(c *ishell.Context) {
			if session.IsStopped() {
				c.Print("Passphrase: ")
				passphrase := c.ReadPassword()

				c.Println("Starting session")
				err := session.Start(passphrase)
				if err != nil {
					c.Println("error starting session:", err)
					return
				}
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "stop-session",
		Help: "stops the session",
		Func: func(c *ishell.Context) {
			c.Println("Stopping session")
			session.Stop()
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "create-key",
		Help: "creates a new key and saves it to the wallet",
		Func: func(c *ishell.Context) {
			key, err := session.GenerateKey()
			if err != nil {
				c.Println("error generating key:", err)
				return
			}
			c.Println(crypto.PubkeyToAddress(key.PublicKey).String())
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "list-keys",
		Help: "list the keys in the wallet",
		Func: func(c *ishell.Context) {
			keys, err := session.ListKeys()
			if err != nil {
				c.Printf("error listing key: %v\n", err)
				return
			}
			for i, addr := range keys {
				c.Println(strconv.Itoa(i) + ": " + addr)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "create-chain",
		Help: "create a new chain based on a key",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("not enough arguments passed to create-chain")
				return
			}

			chain, err := session.CreateChain(c.Args[0], nil)
			if err != nil {
				c.Printf("error creating chain tree: %v\n", err)
				return
			}

			chainID, err := chain.Id()
			if err != nil {
				c.Printf("error fetching chain id: %v\n", err)
				return
			}

			c.Printf("chain-id: %s\n", chainID)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "set-owner",
		Help: "transfer ownership of a chain tree",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 3 {
				c.Println("not enough arguments passed to set-owner")
				return
			}

			newOwnerKeys := strings.Split(c.Args[2], ",")
			tip, err := session.SetOwner(c.Args[0], c.Args[1], newOwnerKeys)
			if err != nil {
				c.Printf("error setting owners: %v\n", err)
				return
			}

			c.Printf("new tip: %v\n", tip)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "export-chain",
		Help: "export an existing chain tree",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("not enough arguments passed to export-chain")
				return
			}

			encodedChain, err := session.ExportChain(c.Args[0])
			if err != nil {
				c.Printf("error exporting chain tree: %v\n", err)
				return
			}

			c.Printf("serialized chain tree: %v\n", encodedChain)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "import-chain",
		Help: "import a chain tree",
		Func: func(c *ishell.Context) {
			if len(c.Args) != 1 {
				c.Println("incorrect number of arguments passed to import-chain")
				return
			}

			chain, err := session.ImportChain(c.Args[0], true, nil)
			if err != nil {
				c.Printf("error importing chain tree: %v\n", err)
				return
			}

			c.Printf("chain id: %v\n", chain.MustId())
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "list-chains",
		Help: "list the current chain tree ids",
		Func: func(c *ishell.Context) {
			ids, err := session.GetChainIds()
			if err != nil {
				c.Printf("error listing chain: %v\n", err)
				return
			}
			for i, id := range ids {
				c.Println(strconv.Itoa(i) + ": " + id)
			}
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "print-chain",
		Help: "print an entire chain tree",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("not enough arguments passed to print-chain")
				return
			}

			chain, err := session.GetChain(c.Args[0])
			if err != nil {
				c.Printf("error getting chain: %v\n", err)
				return
			}
			c.Println(chain.ChainTree.Dag.Dump())
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "get-tip",
		Help: "gets the tip (as known by the notary group) for a chain id",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 1 {
				c.Println("not enough arguments passed to get-tip")
				return
			}

			tip, err := session.GetTip(c.Args[0])
			if err != nil {
				c.Printf("error getting tip: %v\n", err)
				return
			}

			c.Printf("tip: %v\n", tip)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "set-data",
		Help: "set-data on a chain-tree usage: set-data chain-id key-id path value",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 4 {
				c.Println("not enough arguments passed to set-data")
				return
			}

			data, err := cbornode.DumpObject(c.Args[3])
			if err != nil {
				c.Printf("error encoding input: %v\n", err)
				return
			}

			tip, err := session.SetData(c.Args[0], c.Args[1], c.Args[2], data)
			if err != nil {
				c.Printf("error setting data: %v\n", err)
				return
			}

			c.Printf("new tip: %v\n", tip)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "resolve",
		Help: "resolve the data at a chain-tree path. usage: resolve chain-id path",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Println("not enough arguments passed to resolve")
				return
			}

			path, err := consensus.DecodePath(c.Args[1])
			if err != nil {
				c.Printf("bad path: %v\n", err)
				return
			}

			data, remaining, err := session.Resolve(c.Args[0], path)
			if err != nil {
				c.Printf("error resolving data: %v\n", err)
				return
			}

			c.Printf("data: %v\nremaining path: %v\n", data, remaining)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "resolve-data",
		Help: "resolve the data in a chain-tree under the user storage path (tree/data). usage: resolve-data chain-id path",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Println("not enough arguments passed to resolve-data")
				return
			}

			c.Args[1] = fmt.Sprintf("tree/data/%s", strings.TrimPrefix(c.Args[1], "/"))

			shellArgs := append([]string{"resolve"}, c.Args...)
			if err = shell.Process(shellArgs...); err != nil {
				log.Printf("failed to run shell args %+v: %s", shellArgs, err)
				// TODO: Enable
				// panic(fmt.Errorf("failed to run shell args %+v: %s", shellArgs, err))
			}
		},
	})

	establishTokenUsage := "usage: establish-token chain-id key-id token-name max-tokens"
	shell.AddCmd(&ishell.Cmd{
		Name: "establish-token",
		Help: "establish new token. 0 for max-tokens means unlimited. " + establishTokenUsage,
		Func: func(c *ishell.Context) {
			if len(c.Args) < 4 {
				c.Println("not enough arguments to establish-token. " + establishTokenUsage)
				return
			}
			maxTokens, err := strconv.ParseUint(c.Args[3], 10, 64)
			if err != nil {
				c.Printf("error parsing max-tokens \"%s\": %v\n", maxTokens, err)
				return
			}
			tip, err := session.EstablishToken(c.Args[0], c.Args[1], c.Args[2], maxTokens)
			if err != nil {
				c.Printf("error establishing token: %v\n", err)
				return
			}

			c.Printf("new tip: %v\n", tip)
		},
	})

	mintTokenUsage := "usage: mint-token chain-id key-id token-name amount"
	shell.AddCmd(&ishell.Cmd{
		Name: "mint-token",
		Help: "mint token(s). must be established first (see establish-token). " + mintTokenUsage,
		Func: func(c *ishell.Context) {
			if len(c.Args) < 4 {
				c.Println("not enough arguments to mint-token. " + mintTokenUsage)
				return
			}
			amount, err := strconv.ParseUint(c.Args[3], 10, 64)
			if err != nil {
				c.Printf("error parsing amount \"%s\": %v\n", amount, err)
				return
			}
			tip, err := session.MintToken(c.Args[0], c.Args[1], c.Args[2], amount)
			if err != nil {
				c.Printf("error minting token: %v\n", err)
				return
			}

			c.Printf("new tip: %v\n", tip)
		},
	})

	sendTokenUsage := "usage: send-token chain-id key-id token-name destination-chain-id amount"
	shell.AddCmd(&ishell.Cmd{
		Name: "send-token",
		Help: "send token(s) to another chaintree. " + sendTokenUsage,
		Func: func(c *ishell.Context) {
			if len(c.Args) < 5 {
				c.Println("not enough arguments to send-token. " + sendTokenUsage)
				return
			}
			amount, err := strconv.ParseUint(c.Args[4], 10, 64)
			if err != nil {
				c.Printf("error parsing amount \"%s\": %v\n", amount, err)
				return
			}
			token, err := session.SendToken(c.Args[0], c.Args[1], c.Args[2], c.Args[3], amount)
			if err != nil {
				c.Printf("error generating send token payload: %v\n", err)
				return
			}

			c.Printf("token: %s\n", token)
		},
	})

	receiveTokenUsage := "usage: receive-token chain-id key-id token-payload"
	shell.AddCmd(&ishell.Cmd{
		Name: "receive-token",
		Help: "receives token(s) sent to a local chaintree. " + receiveTokenUsage,
		Func: func(c *ishell.Context) {
			if len(c.Args) < 3 {
				c.Println("not enough arguments to receive-token. " + receiveTokenUsage)
				return
			}
			tip, err := session.ReceiveToken(c.Args[0], c.Args[1], c.Args[2])
			if err != nil {
				c.Printf("error receiving token: %v\n", err)
				return
			}

			c.Printf("new tip: %v\n", tip)
		},
	})

	getTokenBalanceUsage := "usage: get-token-balance chain-id token-name"
	shell.AddCmd(&ishell.Cmd{
		Name: "get-token-balance",
		Help: "get the balance for a token in a chain tree. " + getTokenBalanceUsage,
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Println("not enough arguments to get-token-balance. " + getTokenBalanceUsage)
				return
			}
			bal, err := session.GetTokenBalance(c.Args[0], c.Args[1])
			if err != nil {
				c.Printf("error getting token balance: %s\n", err)
				return
			}

			c.Printf("balance: %d\n", bal)
		},
	})

	listTokensUsage := "usage: list-tokens chain-id key-id"
	shell.AddCmd(&ishell.Cmd{
		Name: "list-tokens",
		Help: "lists all tokens and their balances. " + listTokensUsage,
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Println("not enough arguments to list-tokens. " + listTokensUsage)
				return
			}
			tokens, err := session.ListTokens(c.Args[0], c.Args[1])
			if err != nil {
				c.Printf("error listing tokens: %v\n", err)
				return
			}

			c.Print(tokens)
		},
	})

	// run shell
	shell.Run()
}
