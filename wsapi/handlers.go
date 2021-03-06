// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/factomapi"
	"github.com/FactomProject/FactomCode/wallet"
	"github.com/FactomProject/gocoding"
	"github.com/hoisie/web"
)

func handleBlockHeight(ctx *web.Context) {
	log := serverLog
	log.Debug("handleBlockHeight")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	height, err := factomapi.GetBlokHeight()
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad Request")
		log.Error(err)
		return
	}

	fmt.Fprint(buf, height)
}

// handleBuyCredit will add entry credites to the specified key. Currently the
// entry credits are given without any factoid transactions occuring.
func handleBuyCredit(ctx *web.Context) {
	log := serverLog
	log.Debug("handleBuyCreditPost")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	var abortMessage, abortReturn string

	defer func() {
		if abortMessage != "" && abortReturn != "" {
			ctx.Header().Add("Location", fmt.Sprint("/failed?message=",
				abortMessage, "&return=", abortReturn))
			ctx.WriteHeader(303)
		}
	}()

	ecPubKey := new(common.Hash)
	if ctx.Params["to"] == "wallet" {
		ecPubKey.Bytes = (*wallet.ClientPublicKey().Key)[:]
	} else {
		ecPubKey.Bytes, _ = hex.DecodeString(ctx.Params["to"])
	}

	log.Info("handleBuyCreditPost using pubkey: ", ecPubKey, " requested",
		ctx.Params["to"])

	factoid, err := strconv.ParseFloat(ctx.Params["value"], 10)
	if err != nil {
		log.Error(err)
	}
	value := uint64(factoid * 1000000000)
	err = factomapi.BuyEntryCredit(1, ecPubKey, nil, value, 0, nil)
	if err != nil {
		abortMessage = fmt.Sprint("An error occured while submitting the buycredit request: ", err.Error())
		log.Error(err)
		return
	} else {
		fmt.Fprintln(ctx, "MsgGetCredit Submitted")
	}
}

// handleChainByHash will take a ChainID and return the associated Chain
func handleChainByHash(ctx *web.Context, hash string) {
	log := serverLog
	log.Debug("handleChainByHash")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	chain, err := factomapi.GetChainByHashStr(hash)
	log.Debugf("%#v", chain)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad Request")
		log.Error(err)
		return
	}

	// Send back JSON response
	err = factomapi.SafeMarshal(buf, chain)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request ")
		log.Error(err)
		return
	}
}

// handleChains will return all chains from the backend database
func handleChains(ctx *web.Context) {
	log := serverLog
	log.Debug("handleChains")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	chains, err := factomapi.GetAllChains()
	log.Debugf("Got %d chains", len(chains))
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad Request")
		log.Error(err)
		return
	}

	// Send back JSON response
	for _, v := range chains {
		err = factomapi.SafeMarshal(buf, v)
		if err != nil {
			log.Error(err)
			break
		}
	}

	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request ")
		log.Error(err)
		return
	}
}

// handleCreditBalance will return the current entry credit balance of the
// spesified pubKey
func handleCreditBalance(ctx *web.Context) {
	log := serverLog
	log.Debug("handleGetCreditBalance")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	ecPubKey := new(common.Hash)
	if ctx.Params["pubkey"] == "wallet" {
		ecPubKey.Bytes = (*wallet.ClientPublicKey().Key)[:]
	} else {
		p, err := hex.DecodeString(ctx.Params["pubkey"])
		if err != nil {
			log.Error(err)
		}
		ecPubKey.Bytes = p
	}

	log.Info("handleGetCreditBalance using pubkey: ", ecPubKey,
		" requested", ctx.Params["pubkey"])

	balance, err := factomapi.GetEntryCreditBalance(ecPubKey)
	if err != nil {
		log.Error(err)
	}

	ecBalance := new(common.ECBalance)
	ecBalance.Credits = balance
	ecBalance.PublicKey = ecPubKey

	log.Info("Balance for pubkey ", ctx.Params["pubkey"], " is: ", balance)

	// Send back JSON response
	err = factomapi.SafeMarshal(buf, ecBalance)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request ")
		log.Error(err)
		return
	}
}

// handleDBlockByHash will take a directory block hash and return the directory
// block information in json format.
func handleDBlockByHash(ctx *web.Context, hashStr string) {
	log := serverLog
	log.Debug("handleDBlockByHash")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	dBlock, err := factomapi.GetDirectoryBlokByHashStr(hashStr)
	log.Debug(dBlock)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad Request")
		log.Error(err)
		return
	}

	// Send back JSON response
	err = factomapi.SafeMarshal(buf, dBlock)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request ")
		log.Error(err)
		return
	}
}

// handleDBInfoByHash will take a Directory Block Hash and return the directory
// block information in json format.
func handleDBInfoByHash(ctx *web.Context, hashStr string) {
	log := serverLog
	log.Debug("handleDBInfoByHash")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	dBInfo, err := factomapi.GetDBInfoByHashStr(hashStr)
	log.Debug(dBInfo)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad Request")
		log.Error(err)
		return
	}

	// Send back JSON response
	err = factomapi.SafeMarshal(buf, dBInfo)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request ")
		log.Error(err)
		return
	}
}

// handleDBlockByRange will get a block height range and return the information
// for all of the directory blocks within the range in json format.
func handleDBlocksByRange(ctx *web.Context, fromHeightStr string,
	toHeightStr string) {
	log := serverLog
	log.Debug("handleDBlocksByRange")

	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	fromBlockHeight, err := strconv.Atoi(fromHeightStr)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad fromBlockHeight")
		log.Error(err)
		return
	}
	toBlockHeight, err := strconv.Atoi(toHeightStr)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad toBlockHeight")
		log.Error(err)
		return
	}

	dBlocks, err := factomapi.GetDirectoryBloks(uint32(fromBlockHeight),
		uint32(toBlockHeight))
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request")
		log.Error(err)
		return
	}

	// Send back JSON response
	for _, block := range dBlocks {
		err = factomapi.SafeMarshal(buf, &block)
		if err != nil {
			httpcode = 400
			buf.WriteString("Bad request")
			log.Error(err)
			return
		}
	}
}

func handleEBlockByHash(ctx *web.Context, hashStr string) {
	log := serverLog
	log.Debug("handleEBlockByHash")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	eBlock, err := factomapi.GetEntryBlokByHashStr(hashStr)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad Request")
		log.Error(err)
		return
	}

	// Send back JSON response
	err = factomapi.SafeMarshal(buf, eBlock)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request")
		log.Error(err)
		return
	}
}

func handleEBlockByMR(ctx *web.Context, mrStr string) {
	log := serverLog
	log.Debug("handleEBlockByMR")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()
	log.Info("mrstr:", mrStr)
	newstr, _ := url.QueryUnescape(mrStr)
	log.Info("newstr:", newstr)
	eBlock, err := factomapi.GetEntryBlokByMRStr(newstr)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad Request")
		log.Error(err)
		return
	}

	// Send back JSON response
	err = factomapi.SafeMarshal(buf, eBlock)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request")
		log.Error(err)
		return
	}
}

func handleEntryByHash(ctx *web.Context, hashStr string) {
	log := serverLog
	log.Debug("handleEBlockByMR")
	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	entry, err := factomapi.GetEntryByHashStr(hashStr)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad Request")
		log.Error(err)
		return
	}

	// Send back JSON response
	err = factomapi.SafeMarshal(buf, entry)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request")
		log.Error(err)
		return
	}
}

// handleEntriesByExtID will get a series of entries and return them as a
// stream of json objects.
func handleEntriesByExtID(ctx *web.Context, eid string) {
	log := serverLog
	log.Debug("handleEntriesByExtID")

	var httpcode int = 200
	buf := new(bytes.Buffer)

	defer func() {
		ctx.WriteHeader(httpcode)
		ctx.Write(buf.Bytes())
	}()

	entries, err := factomapi.GetEntriesByExtID(eid)
	if err != nil {
		httpcode = 400
		buf.WriteString("Bad request")
		log.Error(err)
		return
	}

	// Send back JSON response
	for _, entry := range entries {
		err = factomapi.SafeMarshal(buf, &entry)
		if err != nil {
			httpcode = 400
			buf.WriteString("Bad request")
			log.Error(err)
			return
		}
	}
}

// handleSubmitChain converts a json post to a factomapi.Chain then submits the
// entry to factomapi.
func handleSubmitChain(ctx *web.Context) {
	log := serverLog
	log.Debug("handleSubmitChain")
	switch ctx.Params["format"] {
	case "json":
		reader := gocoding.ReadBytes([]byte(ctx.Params["chain"]))
		c := new(common.EChain)
		factomapi.SafeUnmarshal(reader, c)

		if c.FirstEntry == nil {
			fmt.Fprintln(ctx,
				"The first entry is required for submitting the chain:")
			log.Warning("The first entry is required for submitting the chain")
			return
		} else {
            c.FirstEntry.GenerateIDFromName()
            c.ChainID = &c.FirstEntry.ChainID
		}

		log.Debug("c.ChainID:", c.ChainID.String())

		if err := factomapi.CommitChain(c); err != nil {
			fmt.Fprintln(ctx,
				"there was a problem with submitting the chain:", err)
			log.Error(err)
		}

		time.Sleep(1 * time.Second)

		if err := factomapi.RevealChain(c); err != nil {
			fmt.Println(err.Error())
			fmt.Fprintln(ctx,
				"there was a problem with submitting the chain:", err)
			log.Error(err)
		}

		fmt.Fprintln(ctx, "Chain Submitted")
	default:
		ctx.WriteHeader(403)
	}
}

// handleSubmitEntry converts a json post to a factom.Entry then submits the
// entry to factom.
func handleSubmitEntry(ctx *web.Context) {
	log := serverLog
	log.Debug("handleSubmitEntry")

	switch ctx.Params["format"] {
	case "json":
		entry := new(common.Entry)
		reader := gocoding.ReadBytes([]byte(ctx.Params["entry"]))
		err := factomapi.SafeUnmarshal(reader, entry)
		if err != nil {
			fmt.Fprintln(ctx,
				"there was a problem with submitting the entry:", err.Error())
			log.Error(err)
		}

		if err := factomapi.CommitEntry(entry); err != nil {
			fmt.Fprintln(ctx,
				"there was a problem with submitting the entry:", err.Error())
			log.Error(err)
		}

		time.Sleep(1 * time.Second)
		if err := factomapi.RevealEntry(entry); err != nil {
			fmt.Fprintln(ctx,
				"there was a problem with submitting the entry:", err.Error())
			log.Error(err)
		}
		fmt.Fprintln(ctx, "Entry Submitted")
	default:
		ctx.WriteHeader(403)
	}
}
