package main

import (
	"database/sql"
	"log"
	"math/rand"
	"net"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
)

type jwtMetaInfo struct {
	UA string `json:"ua"`
	IP net.IP `json:"ip"`
}

var usernameRegexStr = `^[a-zA-Z0-9_\-\[\]{}^@\*()=$.<>]{5,30}$`
var usernameRegex = regexp.MustCompile(usernameRegexStr)
var errLogin = `{"error":"Please check your username and password"}`
var jwtTokenSecret = []byte(`VeriS3c3tStrHiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiixx:P. mlem mdem. Em deep le'm!!`)
var accessTokenDuration int64 = 5 * 24 * 3600

func loginHandler(ctx *fasthttp.RequestCtx) (b bool) {
	username := ctx.FormValue("username")
	if usernameRegex.Match(username) == false {
		ctx.SetBodyString(`{"error":` + jsonEncode(`Username does not match `+usernameRegexStr) + `}`)
		return
	}
	password := ctx.FormValue("password")
	if len(password) < 6 || len(password) > 100 {
		ctx.SetBodyString(`{"error":"5 < password length < 100"}`)
		return
	}
	time.Sleep(time.Duration(1000+rand.Int63n(2000)) * time.Millisecond)
	row, err := db.Row(`select id,password from users where username = ? limit 1`, username)
	if err != nil {
		log.Println(err)
		ctx.SetBodyString(errLogin)
		return
	}
	if VerifyPasswdHash(row["password"], password) == false {
		// Đăng nhập thất bại
		ctx.SetBodyString(errLogin)
		return
	}
	// Đăng nhập thành công
	_, err = db.Exec(`delete from tokens where expire < UNIX_TIMESTAMP()`)
	if err != nil {
		log.Println(err)
		ctx.SetBodyString(strServerErr)
		return
	}

	timeNow := time.Now().Unix()
	expire := timeNow + accessTokenDuration
	id, err := db.Insert("tokens", map[string]interface{}{
		"user_id": row["id"],
		"info":    jsonEncode(&jwtMetaInfo{b2s(ctx.UserAgent()), ctx.RemoteIP()}), // Lưu UA, IP, thông tin về trình duyệt gì đó
		"expire":  expire,
	})
	if err != nil {
		log.Println(err)
		ctx.SetBodyString(strServerErr)
		return
	}
	// Tạo jwt
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"uid": string(row["id"]),
		"nbf": timeNow,
		"exp": expire,
		"iat": timeNow,
	}).SignedString(jwtTokenSecret)
	if err != nil {
		log.Println(err)
		ctx.SetBodyString(strServerErr)
		return
	}
	ctx.SetBodyString(jsonEncode(map[string]interface{}{
		"access_token": accessToken,
		"expire":       expire,
	}))
	return
}

var errSrvRegister = `{"error":"Server error"}`

func registerHandler(ctx *fasthttp.RequestCtx) (b bool) {
	var err error
	username := ctx.FormValue("username")
	if usernameRegex.Match(username) == false {
		ctx.SetBodyString(`{"error":` + jsonEncode(`Username does not match `+usernameRegexStr) + `}`)
		return
	}
	password := ctx.FormValue("password")
	if len(password) < 6 || len(password) > 100 {
		ctx.SetBodyString(`{"error":"5 < password length < 100"}`)
		return
	}
	_, err = db.Row(`select 1 from users where username = ? limit 1`, username)
	if err == nil {
		ctx.SetBodyString(`{"error":"Username exists"}`)
		return
	} else if err != sql.ErrNoRows {
		log.Println(err)
		ctx.SetBodyString(errSrvRegister)
		return
	}
	passwdHash, err := CreatePasswdHash(password)
	if err != nil {
		log.Println(err)
		ctx.SetBodyString(errSrvRegister)
		return
	}
	_, err = db.Insert("users", map[string]interface{}{
		"username":  username,
		"password":  passwdHash,
		"create_at": time.Now().Unix(),
	})
	if err != nil {
		log.Println(err)
		ctx.SetBodyString(errSrvRegister)
		return
	}
	ctx.SetBodyString(`{}`)
	return
}
