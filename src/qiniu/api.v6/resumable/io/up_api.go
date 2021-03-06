package io

import (
	"encoding/base64"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"strconv"

	. "qiniu/api.v6/conf"
	"qiniu/bytes"
	"qiniu/log"
	"qiniu/rpc"
)

// ----------------------------------------------------------

const (
	InvalidCtx = 701 // UP: 无效的上下文(bput)，可能情况：Ctx非法或者已经被淘汰（太久未使用）
)

// ----------------------------------------------------------

var ErrUnmatchedChecksum = errors.New("unmatched checksum")

// ----------------------------------------------------------

func encode(raw string) string {
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

// ----------------------------------------------------------

type Transport struct {
	token     string
	transport http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	req.Header.Set("Authorization", t.token)
	return t.transport.RoundTrip(req)
}

func NewTransport(token string, transport http.RoundTripper) *Transport {
	return &Transport{"UpToken " + token, transport}
}

func NewClientEx(token string, transport http.RoundTripper, bindRemoteIp string) rpc.Client {
	t := NewTransport(token, transport)
	return rpc.Client{&http.Client{Transport: t}, bindRemoteIp}
}

func NewClient(token string, bindRemoteIp string) rpc.Client {
	transport := http.DefaultTransport
	return NewClientEx(token, transport, bindRemoteIp)
}

// ----------------------------------------------------------

func Mkblock(
c rpc.Client, l rpc.Logger, ret *BlkputRet, blockSize int, body io.Reader, size int) error {
	return c.CallWith(l, ret, UP_HOST + "/mkblk/" + strconv.Itoa(blockSize), "application/octet-stream", body, size)
}

func Blockput(
c rpc.Client, l rpc.Logger, ret *BlkputRet, body io.Reader, size int) error {

	url := UP_HOST + "/bput/" + ret.Ctx + "/" + strconv.FormatUint(uint64(ret.Offset), 10)
	return c.CallWith(l, ret, url, "application/octet-stream", body, size)
}

// ----------------------------------------------------------

func ResumableBlockput(
c rpc.Client, l rpc.Logger, ret *BlkputRet, f io.ReaderAt, blkIdx, blkSize int, extra *PutExtra) (err error) {

	h := crc32.NewIEEE()
	offbase := int64(blkIdx) << blockBits
	chunkSize := extra.ChunkSize

	var bodyLength int

	if ret.Ctx == "" {

		if chunkSize < blkSize {
			bodyLength = chunkSize
		} else {
			bodyLength = blkSize
		}

		body1 := io.NewSectionReader(f, offbase, int64(bodyLength))
		body := io.TeeReader(body1, h)

		err = Mkblock(c, l, ret, blkSize, body, bodyLength)
		if err != nil {
			return
		}
		if ret.Crc32 != h.Sum32() || int(ret.Offset) != bodyLength {
			err = ErrUnmatchedChecksum
			return
		}
		extra.Notify(blkIdx, blkSize, ret)
	}

	for int(ret.Offset) < blkSize {

		if chunkSize < blkSize - int(ret.Offset) {
			bodyLength = chunkSize
		} else {
			bodyLength = blkSize - int(ret.Offset)
		}

		tryTimes := extra.TryTimes

		lzRetry:
		h.Reset()
		body1 := io.NewSectionReader(f, offbase + int64(ret.Offset), int64(bodyLength))
		body := io.TeeReader(body1, h)

		err = Blockput(c, l, ret, body, bodyLength)
		if err == nil {
			if ret.Crc32 == h.Sum32() {
				extra.Notify(blkIdx, blkSize, ret)
				continue
			}
			log.Warn("ResumableBlockput: invalid checksum, retry")
			err = ErrUnmatchedChecksum
		} else {
			if ei, ok := err.(*rpc.ErrorInfo); ok && ei.Code == InvalidCtx {
				ret.Ctx = "" // reset
				log.Warn("ResumableBlockput: invalid ctx, please retry")
				return
			}
			log.Warn("ResumableBlockput: bput failed -", err)
		}
		if tryTimes > 1 {
			tryTimes--
			log.Info("ResumableBlockput retrying ...")
			goto lzRetry
		}
		break
	}
	return
}

// ----------------------------------------------------------

func Mkfile(
c rpc.Client, l rpc.Logger, ret interface{}, key string, hasKey bool, fsize int64, extra *PutExtra) (err error) {

	url := UP_HOST + "/mkfile/" + strconv.FormatInt(fsize, 10)

	if extra.MimeType != "" {
		url += "/mimeType/" + encode(extra.MimeType)
	}
	if hasKey {
		url += "/key/" + encode(key)
	}
	for k, v := range extra.Params {
		url += fmt.Sprintf("/%s/%s", k, encode(v))
	}

	buf := make([]byte, 0, 176 * len(extra.Progresses))
	for _, prog := range extra.Progresses {
		buf = append(buf, prog.Ctx...)
		buf = append(buf, ',')
	}
	if len(buf) > 0 {
		buf = buf[:len(buf) - 1]
	}

	return c.CallWith(l, ret, url, "application/octet-stream", bytes.NewReader(buf), len(buf))
}

// ----------------------------------------------------------
