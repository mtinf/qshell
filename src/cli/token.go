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
	if len(params) >= 4 {
		accessKey := params[0]
		secretKey := params[1]
		bucket := params[2]
		key := params[3]
		time := 360000
		if len(params) == 5 {
			time, _ = strconv.Atoi(params[4])
		}

		mac := digest.Mac{accessKey, []byte(secretKey)}

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