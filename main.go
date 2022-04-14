package main

import (
	"encoding/hex"
	"fmt"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
	"os"
	"syscall"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/polynetwork/bridge-common/chains/ont"
	"github.com/polynetwork/bridge-common/util"
	"github.com/polynetwork/bridge-common/log"
	"github.com/polynetwork/bridge-common/wallet"
)

func main() {
	app := &cli.App{
		Name:  "ont-tool",
		Usage: "ont tool",
		Before: Init,
		Commands: []*cli.Command{
			&cli.Command{
				Name:   "bindasset",
				Usage:  "bind asset",
				Action: BindAsset,
				Flags: []cli.Flag{
					&cli.Int64Flag{
						Name:     "tochain",
						Usage:    "target side chain",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "url",
						Usage:    "rpc url",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "wallet",
						Usage:    "wallet file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "proxy",
						Usage:    "proxy contract",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "fromasset",
						Usage:    "fromasset",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "toasset",
						Usage:    "to asset",
						Required: true,
					},
				},
			},
			&cli.Command{
				Name:   "bindproxy",
				Usage:  "bind proxy",
				Action: BindProxy,
				Flags: []cli.Flag{
					&cli.Int64Flag{
						Name:     "tochain",
						Usage:    "target side chain",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "url",
						Usage:    "rpc url",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "wallet",
						Usage:    "wallet file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "proxy",
						Usage:    "proxy contract",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "toproxy",
						Usage:    "to proxy",
						Required: true,
					},
				},
			},
			&cli.Command{
				Name:   "bindassetcheck",
				Usage:  "bind asset check",
				Action: BindAssetCheck,
				Flags: []cli.Flag{
					&cli.Int64Flag{
						Name:     "tochain",
						Usage:    "target side chain",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "url",
						Usage:    "rpc url",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "wallet",
						Usage:    "wallet file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "proxy",
						Usage:    "proxy contract",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "fromasset",
						Usage:    "fromasset",
						Required: true,
					},
				},
			},
			&cli.Command{
				Name:   "bindproxycheck",
				Usage:  "bind proxy check",
				Action: BindProxyCheck,
				Flags: []cli.Flag{
					&cli.Int64Flag{
						Name:     "tochain",
						Usage:    "target side chain",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "url",
						Usage:    "rpc url",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "wallet",
						Usage:    "wallet file",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "proxy",
						Usage:    "proxy contract",
						Required: true,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("Start error", "err", err)
	}
}

func setup(c *cli.Context) (sdk *ont.SDK, account *wallet.OntSigner, err error) {
	sdk, err = ont.WithOptions(3, []string{c.String("url")}, time.Minute, 10)
	if err != nil {
		return
	}
	fmt.Println("Enter Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return
	}
	account, err = wallet.NewOntSigner(&wallet.Config{Path: c.String("wallet"), Password: string(password)})
	if err != nil {
		return
	}
	return
}

func BindProxyCheck(c *cli.Context) (err error) {
	proxy, err := common.AddressFromHexString(util.LowerHex(c.String("proxy")))
	if err != nil {
		return fmt.Errorf("invalid proxy address %v", c.String("proxy"))
	}
	sdk, err := ont.WithOptions(3, []string{c.String("url")}, time.Minute, 10)
	if err != nil {
		return
	}
	toChain := c.Int("tochain")
	if toChain == 0 {
		return fmt.Errorf("invalid to chain id %v", toChain)
	}

	res, err := sdk.Node().NeoVM.PreExecInvokeNeoVMContract(proxy, []interface{}{"getProxyHash", []interface{}{toChain}})
	if err != nil { return }
	s, err := res.Result.ToByteArray()
	if err != nil { return }
	log.Info("BindProxyCheck", "res", util.Json(res))
	fmt.Printf("state: %v\n", res.State)
	fmt.Printf("%x\n", s)
	return
}

func BindProxy(c *cli.Context) (err error) {
	proxy, err := common.AddressFromHexString(util.LowerHex(c.String("proxy")))
	if err != nil {
		return fmt.Errorf("invalid proxy address %v", c.String("proxy"))
	}
	sdk, account, err := setup(c)
	if err != nil {
		return
	}

	toProxy, err := hex.DecodeString(util.LowerHex(c.String("toproxy")))
	if err != nil {
		return fmt.Errorf("invalid to proxy %s", c.String("toproxy"))
	}
	toChain := c.Int("tochain")
	if toChain == 0 {
		return fmt.Errorf("invalid to chain id %v", toChain)
	}

	res, err := sdk.Node().NeoVM.InvokeNeoVMContract(2500, 200000, account.Account, account.Account, proxy, []interface{}{"bindProxyHash", []interface{}{toChain, toProxy}})
	log.Info("BindProxy", "hash", res.ToHexString(), "err", err)

	if err == nil {
		log.Info("Will check after 10 seconds")
		time.Sleep(10 * time.Second)
	}
	BindProxyCheck(c)
	return
}


func BindAssetCheck(c *cli.Context) (err error) {
	proxy, err := common.AddressFromHexString(util.LowerHex(c.String("proxy")))
	if err != nil { return fmt.Errorf("invalid proxy address %v", c.String("proxy"))}
	sdk, err := ont.WithOptions(3, []string{c.String("url")}, time.Minute, 10)
	if err != nil {
		return
	}

	fromAsset, err := hex.DecodeString(util.LowerHex(c.String("fromasset")))
	if err != nil { return fmt.Errorf("invalid from asset %s", c.String("fromasset")) }
	toChain := c.Int("tochain")
	if toChain == 0 {
		return fmt.Errorf("invalid to chain id %v", toChain)
	}

	res, err := sdk.Node().NeoVM.PreExecInvokeNeoVMContract(proxy, []interface{}{"getAssetHash", []interface{}{fromAsset, toChain}})
	if err != nil { return }
	s, err := res.Result.ToByteArray()
	if err != nil { return }
	log.Info("BindAssetCheck", "res", util.Json(res))
	fmt.Printf("state: %v\n", res.State)
	fmt.Printf("%x\n", s)
	return
}


func BindAsset(c *cli.Context) (err error) {
	proxy, err := common.AddressFromHexString(util.LowerHex(c.String("proxy")))
	if err != nil { return fmt.Errorf("invalid proxy address %v", c.String("proxy"))}
	sdk, account, err := setup(c)
	if err != nil { return }
	fromAsset, err := hex.DecodeString(util.LowerHex(c.String("fromasset")))
	if err != nil { return fmt.Errorf("invalid from asset %s", c.String("fromasset")) }
	toAsset, err := hex.DecodeString(util.LowerHex(c.String("toasset")))
	if err != nil { return fmt.Errorf("invalid to asset %s", c.String("toasset")) }
	toChain := c.Int("tochain")
	if toChain == 0 {
		return fmt.Errorf("invalid to chain id %v", toChain)
	}

	res, err := sdk.Node().NeoVM.InvokeNeoVMContract(2500, 200000, account.Account, account.Account, proxy, []interface{}{"bindAssetHash", []interface{}{fromAsset, toChain, toAsset}})
	log.Info("BindAsset", "hash", res.ToHexString(), "err", err)

	if err == nil {
		log.Info("Will check after 10 seconds")
		time.Sleep(10 * time.Second)
	}
	BindAssetCheck(c)
	return
}

func Init(ctx *cli.Context) (err error) {
	log.Init()
	return
}
