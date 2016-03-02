package cli
import (
	"qiniu/api.v6/rs"
	"fmt"
	"qiniu/api.v6/auth/digest"
	"strconv"
	"strings"
	"encoding/base64"
)

func GenerateToken(cmd string, params ...string) {
	if len(params) >= 2 {
		bucket := params[0]
		key := params[1]
		time := 360000
		if len(params) == 3 {
			time, _ = strconv.Atoi(params[2])
		}

		gErr := accountS.Get()
		if gErr != nil {
			fmt.Println(gErr)
			return
		}

		mac := digest.Mac{accountS.AccessKey, []byte(accountS.SecretKey)}

		policy := rs.PutPolicy{
			Scope: fmt.Sprintf("%s:%s", bucket, key),
			Expires: uint32(time),
		}
		uptoken := policy.Token(&mac)
		fmt.Printf("Token:: %s\n", uptoken)
	} else {
		CmdHelp(cmd)
	}
}

func ParseToken(cmd string, params ...string)  {
	if len(params) >=1 {
		token := params[0]

		splits := strings.Split(token, ":")
		if(len(splits) <3) {
			return
		}
		putPolicy := splits[2]

		decoded, err := base64.StdEncoding.DecodeString(putPolicy)
		if err != nil {
			fmt.Println("decode error:", err)
			return
		}
		fmt.Println(string(decoded))
	} else {
		CmdHelp(cmd)
	}
}