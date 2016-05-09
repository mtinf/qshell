package cli

import (
	"fmt"
	"qiniu/rpc"
	"qshell"
	"qiniu/api.v6/auth/digest"
	"net/http"
	"net/url"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io/ioutil"
	"encoding/json"
)

func Prefop(cmd string, params ...string) {
	if len(params) == 2 {
		persistentId := params[0]
		host := params[1]
		fopRet := qshell.FopRet{}
		err := rsfopS.Prefop(persistentId, host, &fopRet)
		if err != nil {
			if v, ok := err.(*rpc.ErrorInfo); ok {
				fmt.Println("Prefop error,", v.Code, v.Err)
			} else {
				fmt.Println("Prefop error,", err)
			}
		} else {
			fmt.Println(fopRet.String())
		}
	} else {
		CmdHelp(cmd)
	}
}

func Pfop(cmd string, params ...string) {
	if len(params) == 3 || len(params) == 4 || len(params) == 5 {
		bucket := params[0]
		key := params[1]
		fops := params[2]
		host := params[3]
		gErr := accountS.Get()
		if gErr != nil {
			fmt.Println(gErr)
			return
		}

		mac := digest.Mac{accountS.AccessKey, []byte(accountS.SecretKey)}
		bodyStr := fmt.Sprintf("bucket=%s&key=%s&fops=%s", url.QueryEscape(bucket), url.QueryEscape(key), url.QueryEscape(fops))
		h := hmac.New(sha1.New, []byte(mac.SecretKey))
		signingStr := "/pfop\n" + bodyStr
		h.Write([]byte(signingStr))
		sign := h.Sum(nil)
		encodedSign := base64.URLEncoding.EncodeToString(sign)
		authorization := "QBox " + mac.AccessKey + ":" + string(encodedSign)

		client := &http.Client{}
		r, err := http.NewRequest("POST", fmt.Sprintf("%s/pfop", host), bytes.NewBufferString(bodyStr))
		if (err != nil) {
			fmt.Println(err)
			return
		}
		r.Header.Add("Authorization", authorization)
		resp, _ := client.Do(r)
		result, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(result))
	} else {
		CmdHelp(cmd)
	}
}

func FopCancel(cmd string, params ...string) {
	if len(params) == 2 {
		persistentId := params[0]
		host := params[1]

		client := rpc.DefaultClient
		resp, err := client.Get(nil, fmt.Sprintf("%s/status/cancel/prefop?id=%s", host, persistentId))
		if (err != nil && resp.StatusCode != 200) {
			fmt.Printf("cancel pfop error, persistentId: %s,statusCode: %d, %+v\n", persistentId, resp.StatusCode, err)
			return
		}
		defer resp.Body.Close()

		result, _ := ioutil.ReadAll(resp.Body)

		fopRet := qshell.FopRet{}
		json.Unmarshal(result, &fopRet);

		fmt.Println(fopRet.String())
	} else {
		CmdHelp(cmd)
	}
}
