package tools

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

type Hashes struct {
	MD5    string `json:"md5"`
	SHA1   string `json:"sha1"`
	SHA224 string `json:"sha224"`
	SHA256 string `json:"sha256"`
	SHA384 string `json:"sha384"`
	SHA512 string `json:"sha512"`
}

func HashAll(data []byte) Hashes {
	md5Sum := md5.Sum(data)
	sha1Sum := sha1.Sum(data)
	sha224Sum := sha256.Sum224(data)
	sha256Sum := sha256.Sum256(data)
	sha384Sum := sha512.Sum384(data)
	sha512Sum := sha512.Sum512(data)
	return Hashes{
		MD5:    hex.EncodeToString(md5Sum[:]),
		SHA1:   hex.EncodeToString(sha1Sum[:]),
		SHA224: hex.EncodeToString(sha224Sum[:]),
		SHA256: hex.EncodeToString(sha256Sum[:]),
		SHA384: hex.EncodeToString(sha384Sum[:]),
		SHA512: hex.EncodeToString(sha512Sum[:]),
	}
}
