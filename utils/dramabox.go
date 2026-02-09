package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"github.com/google/uuid"
)

type DramaboxUtils struct {
	privateKey *rsa.PrivateKey
}

func NewDramaboxUtils() *DramaboxUtils {
	d := &DramaboxUtils{}
	d.initPrivateKey()
	return d
}

func (d *DramaboxUtils) decodeString(str string) string {
	result := make([]byte, len(str))
	for i := 0; i < len(str); i++ {
		c := int(str[i])
		if c >= 33 && c <= 126 {
			c -= 20
			if c < 33 {
				c += 126 - 33
			}
		}
		result[i] = byte(c)
	}
	return string(result)
}

func (d *DramaboxUtils) initPrivateKey() {
	part1 := "MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC9Q4Y5QX5j08HrnbY3irfKdkEllAU2OORnAjlXDyCzcm2Z6ZRrGvtTZUAMelfU5PWS6XGEm3d4kJEKbXi4Crl8o2E/E3YJPk1lQD1d0JTdrvZleETN1ViHZFSQwS3L94Woh0E3TPebaEYq88eExvKu1tDdjSoFjBbgMezySnas5Nc2xF28"
	part2 := d.decodeString(`l|d,WL$EI,?xyw+*)^#?U` + "`" + `[whXlG` + "`" + `-GZif,.jCxbKkaY"{w*y]_jax^/1iVDdyg(Wbz+z/$xVjCiH0lZf/d|%gZglW)"~J,^~}w"}m(E'eEunz)eyEy` + "`" + `XGaVF|_(Kw)|awUG"'{{e#%$0E.ffHVU++$giHzdvC0ZLXG|U{aVUUYW{{YVU^x),J'If` + "`" + `nG|C[` + "`" + `ZF),xLv(-H'}ZIEyCfke0dZ%aU[V)"V0}mhKvZ]Gw%-^a|m'` + "`" + `\f}{(~kzi&zjG+|fXX0$IH#j` + "`" + `+hfnME"|fa/{.j.xf,"LZ.K^bZy%c.W^/v{x#(J},Ua,ew#.##K(ki)$LX{a-1\MG/zL&JlEKEw'Hg|D&{EfuKYM[nGKx1V#lFu^V_LjVzw+n%+,Xd` + "`")
	part3 := "x52e71nafqfbjXxZuEtpu92oJd6A9mWbd0BZTk72ZHUmDcKcqjfcEH19SWOphMJFYkxU5FRoIEr3/zisyTO4Mt33ZmwELOrY9PdlyAAyed7ZoH+hlTr7c025QROvb2LmqgRiUT56tMECgYEA+jH5m6iMRK6XjiBhSUnlr3DzRybwlQrtIj5sZprWe2my5uYHG3jbViYIO7GtQvMTnDrBCxNhuM6dPrL0cRnbsp/iBMXe3pyjT/aWveBkn4R+UpBsnbtDn28r1MZpCDtr5UNc0TPj4KFJvjnV/e8oGoyYEroECqcw1LqNOGDiLhkCgYEAwaemNePYrXW+MVX/hatfLQ96tpxwf7yuHdENZ2q5AFw73GJWYvC8VY+TcoKPAmeoCUMltI3TrS6K5Q/GoLd5K2BsoJrSxQNQFd3ehWAtdOuPDvQ5rn/2fsvgvc3rOvJh7uNnwEZCI/45WQg+UFWref4PPc+ArNtp9Xj2y7LndwkCgYARojIQeXmhYZjG6JtSugWZLuHGkwUDzChYcIPd"
	part4 := "W25ndluokG/RzNvQn4+W/XfTryQjr7RpXm1VxCIrCBvYWNU2KrSYV4XUtL+B5ERNj6In6AOrOAifuVITy5cQQQeoD+AT4YKKMBkQfO2gnZzqb8+ox130e+3K/mufoqJPZeyrCQKBgC2fobjwhQvYwYY+DIUharri+rYrBRYTDbJYnh/PNOaw1CmHwXJt5PEDcml3+NlIMn58I1X2U/hpDrAIl3MlxpZBkVYFI8LmlOeR7ereTddN59ZOE4jY/OnCfqA480Jf+FKfoMHby5lPO5OOLaAfjtae1FhrmpUe3EfIx9wVuhKBAoGBAPFzHKQZbGhkqmyPW2ctTEIWLdUHyO37fm8dj1WjN4wjRAI4ohNiKQJRh3QE11E1PzBTl9lZVWT8QtEsSjnrA/tpGr378fcUT7WGBgTmBRaAnv1P1n/Tp0TSvh5XpIhhMuxcitIgrhYMIG3GbP9JNAarxO/qPW6Gi0xWaF7il7Or"

	fullPem := part1 + part2 + part3 + part4
	formattedKey := fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s\n-----END PRIVATE KEY-----", fullPem)

	block, _ := pem.Decode([]byte(formattedKey))
	if block == nil {
		return
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return
	}

	d.privateKey = key.(*rsa.PrivateKey)
}

func (d *DramaboxUtils) Sign(str string) string {
	if d.privateKey == nil {
		d.initPrivateKey()
	}

	if d.privateKey == nil {
		return ""
	}

	hash := sha256.Sum256([]byte(str))
	signature, err := rsa.SignPKCS1v15(rand.Reader, d.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(signature)
}

func (d *DramaboxUtils) RandomAndroidId() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("ffffffff%x000000000", b)
}

func (d *DramaboxUtils) GenerateRandomIP() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%d.%d.%d.%d", b[0], b[1], b[2], b[3])
}

func (d *DramaboxUtils) GenerateUUID() string {
	return uuid.New().String()
}
