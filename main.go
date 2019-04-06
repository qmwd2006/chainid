package main

import (
	//	"errors"
	//	"io"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	_ "crypto/sha256"
	"github.com/opencontainers/go-digest"
	"github.com/urfave/cli"
)

type ChainID digest.Digest

// String returns a string rendition of a layer ID
func (id ChainID) String() string {
	return string(id)
}

// DiffID is the hash of an individual layer tar.
type DiffID digest.Digest

// String returns a string rendition of a layer DiffID
func (diffID DiffID) String() string {
	return string(diffID)
}

// CreateChainID returns ID for a layerDigest slice
func CreateChainID(dgsts []DiffID) ChainID {
	return createChainIDFromParent("", dgsts...)
}

func createChainIDFromParent(parent ChainID, dgsts ...DiffID) ChainID {
	if len(dgsts) == 0 {
		return parent
	}
	if parent == "" {
		return createChainIDFromParent(ChainID(dgsts[0]), dgsts[1:]...)
	}
	// H = "H(n-1) SHA256(n)"
	dgst := digest.FromBytes([]byte(string(parent) + " " + string(dgsts[0])))
	return createChainIDFromParent(ChainID(dgst), dgsts[1:]...)
}

func FmtPrintlnJson(content []byte) {
	var out bytes.Buffer
	err := json.Indent(&out, content, "", "\t")
	if err != nil {
		fmt.Println(err)
	}
	out.WriteTo(os.Stdout)
}

func main() {
	app := cli.NewApp()

	app.Action = func(c *cli.Context) error {
		if c.NArg() != 1 {
			fmt.Println("docker image name or id needed")
			return nil
		}

		name := c.Args().Get(0)
		cmd := exec.Command("docker", "inspect", "-f", "{{json .RootFS.Layers}}", name)
		content, err := cmd.CombinedOutput()
		if err != nil {
			_, ok := err.(*exec.ExitError)
			if ok {
				_ = strings.Join
				//fmt.Println("exec", "\"" + strings.Join(cmd.Args, " ") + "\"", "failed")
				fmt.Println("exec", "'docker inspect -f \"{{json .RootFS.Layers}}\" " + name + "'", "failed")
				return nil
			}
			fmt.Println(err)
			return nil
		}

		//fmt.Println(string(content))
		fmt.Println("diffIds:")
		FmtPrintlnJson(content)
		fmt.Println()

		diffIds := []DiffID{}
		err = json.Unmarshal(content, &diffIds)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(diffIds)

		var chainIds []ChainID
		for index, diffId := range diffIds {
			chainId := CreateChainID(diffIds[:index+1])
			chainIds = append(chainIds, chainId)
			_ = diffId
			//fmt.Println(index, diffId, chainId)
			//fmt.Println(index)
			//fmt.Println(diffId)
			//fmt.Println(chainId)
			//fmt.Println()
		}
		content, err = json.Marshal(chainIds)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(chainIds)
		fmt.Println("chainIds:")
		FmtPrintlnJson(content)
		fmt.Println()
		return nil
	}

	app.Run(os.Args)
}
