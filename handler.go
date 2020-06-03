package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/valyala/fasthttp"
	router "github.com/vinhjaxt/fasthttp-staticrouter"
)

var strServerErr = `{"error":"server error"}`

func jwtKeyFunc(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		alg, ok := token.Header["alg"]
		if ok == false {
			alg = "unknown"
		}
		return nil, fmt.Errorf("Unexpected signing method: %v", alg)
	}
	// this is a []byte containing your secret, e.g. []byte("my_secret_key")
	return jwtTokenSecret, nil
}

func buildHTTPHandler(staticDir *string) func(ctx *fasthttp.RequestCtx) {
	var notFoundBytes = []byte("Not found")
	var notFoundHandler = func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(404)
		ctx.SetBody(notFoundBytes)
	}
	if len(*staticDir) != 0 {
		log.Println("Static file serve:", *staticDir)
		// Go to: https://github.com/valyala/fasthttp/blob/master/examples/fileserver/fileserver.go
		// for more infors
		fs := &fasthttp.FS{
			Root:               *staticDir,
			IndexNames:         []string{"index.html"},
			GenerateIndexPages: false,
			Compress:           true,
			AcceptByteRange:    true,
			PathNotFound:       notFoundHandler,
		}
		notFoundHandler = fs.NewRequestHandler()
	}
	r := router.New()
	r.Post("/login", loginHandler)
	r.Post("/register", registerHandler)

	api := r.Group("/api")
	var strInvalidToken = `{"error":"invalid access_token"}`
	api.Use(func(ctx *fasthttp.RequestCtx) (b bool) {
		ctx.SetContentType("application/json;charset=utf-8")

		accessToken := ctx.Request.Header.Peek("X-Token")
		if len(accessToken) == 0 {
			accessToken = ctx.Request.URI().QueryArgs().Peek("access_token")
			if len(accessToken) == 0 {
				b = true
				ctx.SetBodyString(`{"error":"no access_token"}`)
				return
			}
		}

		jwtToken, err := jwt.Parse(string(accessToken), jwtKeyFunc)
		if err != nil {
			b = true
			log.Println(err)
			ctx.SetBodyString(strInvalidToken)
			return
		}

		tokenClaims, ok := jwtToken.Claims.(jwt.MapClaims)
		if ok == false || jwtToken.Valid == false || tokenClaims.VerifyExpiresAt(time.Now().Unix(), true) == false {
			b = true
			ctx.SetBodyString(strInvalidToken)
			return
		}

		uid, ok := tokenClaims["uid"]
		if ok == false {
			b = true
			ctx.SetBodyString(strInvalidToken)
			return
		}
		tid, ok := tokenClaims["id"]
		if ok == false {
			b = true
			ctx.SetBodyString(strInvalidToken)
			return
		}
		row, err := db.StrRow(`select a.id,a.username from users a inner join tokens b on a.id = b.user_id where a.id = ? and b.id = ? limit 1`, uid, tid) //  and b.expire > UNIX_TIMESTAMP()
		if err != nil {
			log.Println(err)
			ctx.SetBodyString(strInvalidToken)
			b = true
			return
		}

		_, err = db.Exec(`update tokens set last_access_at=UNIX_TIMESTAMP(), info_ss = ? where id = ? limit 1`, jsonEncode(&jwtMetaInfo{b2s(ctx.UserAgent()), ctx.RemoteIP()}), tid)
		if err != nil {
			log.Println(err)
			ctx.SetBodyString(strServerErr)
			b = true
			return
		}

		ctx.SetUserValue("user", row)
		ctx.SetUserValue("token", tokenClaims)
		return
	})

	api.Get("/me", func(ctx *fasthttp.RequestCtx) (b bool) {
		ctx.SetBodyString(sqlRowJSON(ctx.UserValue("user").(map[string]sql.NullString)))
		return
	})

	api.Post("/logout", func(ctx *fasthttp.RequestCtx) (b bool) {
		_, err := db.Exec(`delete from tokens where id = ? limit 1`, ctx.UserValue("token").(jwt.MapClaims)["id"])
		if err != nil {
			log.Println(err)
			ctx.SetBodyString(strServerErr)
			return
		}
		ctx.SetBodyString("1")
		return
	})

	api.Get("/tokens", func(ctx *fasthttp.RequestCtx) (b bool) {
		rows, err := db.StrRows(`select * from tokens where user_id = ?`, ctx.UserValue("user").(map[string]sql.NullString)["id"].String)
		if err != nil {
			if err == sql.ErrNoRows {
				ctx.SetBodyString(`[]`)
				return
			}
			log.Println(err)
			ctx.SetBodyString(strServerErr)
			return
		}

		ctx.SetBodyString(sqlRowsJSON(rows))
		return
	})

	api.Delete("/tokens", func(ctx *fasthttp.RequestCtx) (b bool) {
		id := b2s(ctx.Request.Body())
		if len(id) == 0 {
			ctx.SetBodyString(`{"error":"no id"}`)
			return
		}
		query := `delete from tokens where id = ? limit 1`
		var arg interface{} = id
		if id == "all" {
			query = `delete from tokens where id != ?`
			arg = (ctx.UserValue("token").(jwt.MapClaims))["id"]
		}
		row, err := db.Exec(query, arg)
		if err != nil {
			log.Println(err)
			ctx.SetBodyString(strServerErr)
			return
		}
		n, err := row.RowsAffected()
		if err != nil {
			log.Println(err)
			ctx.SetBodyString(strServerErr)
			return
		}
		ctx.SetBodyString(strconv.FormatInt(n, 10))
		return
	})

	r.NotFound(notFoundHandler)
	return r.BuildHandler()
}
