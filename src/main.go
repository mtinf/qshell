package main

import (
	"cli"
	"fmt"
	"os"
	"qiniu/log"
	"qiniu/rpc"
	"runtime"
)

var debugMode = false

var supportedCmds = map[string]cli.CliFunc{
	"account":       cli.Account,
	"zone":          cli.Zone,
	"dircache":      cli.DirCache,
	"listbucket":    cli.ListBucket,
//	"alilistbucket": cli.AliListBucket,
	"prefop":        cli.Prefop,
	"stat":          cli.Stat,
	"delete":        cli.Delete,
	"move":          cli.Move,
	"copy":          cli.Copy,
	"chgm":          cli.Chgm,
	"sync":          cli.Sync,
	"fetch":         cli.Fetch,
	"prefetch":      cli.Prefetch,
	"batchstat":     cli.BatchStat,
	"batchdelete":   cli.BatchDelete,
	"batchchgm":     cli.BatchChgm,
	"batchrename":   cli.BatchRename,
	"batchcopy":     cli.BatchCopy,
	"batchmove":     cli.BatchMove,
	"batchrefresh":  cli.BatchRefresh,
	"batchsign":     cli.BatchSign,
	"checkqrsync":   cli.CheckQrsync,
	"fput":          cli.FormPut,
	"qupload":       cli.QiniuUpload,
	"qupload2":      cli.QiniuUpload2,
	"qdownload":     cli.QiniuDownload,
	"rput":          cli.ResumablePut,
	"b64encode":     cli.Base64Encode,
	"b64decode":     cli.Base64Decode,
	"urlencode":     cli.Urlencode,
	"urldecode":     cli.Urldecode,
	"ts2d":          cli.Timestamp2Date,
	"tns2d":         cli.TimestampNano2Date,
	"tms2d":         cli.TimestampMilli2Date,
	"d2ts":          cli.Date2Timestamp,
	"ip":            cli.IpQuery,
	"qetag":         cli.Qetag,
	"help":          cli.Help,
	"unzip":         cli.Unzip,
	"privateurl":    cli.PrivateUrl,
	"saveas":        cli.Saveas,
	"reqid":         cli.ReqId,
	"m3u8delete":    cli.M3u8Delete,
	"buckets":       cli.GetBuckets,
	"domains":       cli.GetDomainsOfBucket,
	"pfop" : 	 cli.Pfop,
	"gtoken" :	 cli.GenerateToken,
	"ptoken" :	 cli.ParseToken,
	"fopcancel" :	 cli.FopCancel,
}

func main() {
	//set cpu count
	runtime.GOMAXPROCS(runtime.NumCPU())
	//set qshell user agent
	rpc.UserAgent = cli.UserAgent()

	//parse command
	args := os.Args
	argc := len(args)
	log.SetOutputLevel(log.Linfo)
	log.SetOutput(os.Stdout)
	if argc > 1 {
		cmd := ""
		params := []string{}
		option := args[1]
		if option == "-d" {
			if argc > 2 {
				cmd = args[2]
				if argc > 3 {
					params = args[3:]
				}
			}
			log.SetOutputLevel(log.Ldebug)
		} else if option == "-v" {
			cli.Version()
			return
		} else if option == "-h" {
			cli.Help("help")
			return
		} else {
			cmd = args[1]
			if argc > 2 {
				params = args[2:]
			}
		}
		if cmd == "" {
			fmt.Println("Error: no subcommand specified")
			return
		}

		if cliFunc, ok := supportedCmds[cmd]; ok {
			cliFunc(cmd, params...)
		} else {
			fmt.Println(fmt.Sprintf("Error: unknown cmd `%s'", cmd))
		}
	} else {
		fmt.Println("Use help or help [cmd1 [cmd2 [cmd3 ...]]] to see supported commands.")
	}
}
