/*
This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or
distribute this software, either in source code form or as a compiled
binary, for any purpose, commercial or non-commercial, and by any
means.

In jurisdictions that recognize copyright laws, the author or authors
of this software dedicate any and all copyright interest in the
software to the public domain. We make this dedication for the benefit
of the public at large and to the detriment of our heirs and
successors. We intend this dedication to be an overt act of
relinquishment in perpetuity of all present and future rights to this
software under copyright law.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

For more information, please refer to <https://unlicense.org>
*/


package search

import (
	"database/sql"
	"fmt"
	"github.com/json-iterator/go"
)

var ignoreme = fmt.Errorf("Ook ook!")

var json = jsoniter.ConfigFastest

type KeyBuild struct {
	*sql.DB
	Prefix string
}
func (k *KeyBuild) Initialize(connStr string) (err error) {
	k.DB, err = sql.Open("postgres", connStr)
	if err!=nil { return }
	k.DB.Exec(`CREATE TABLE `+k.Prefix+`searchtab (
		prid text primary key,
		tsbody tsvector,
		txbody text,
		jsmeta jsonb,
		isactive boolean default true
	)`)
	k.DB.Exec(`CREATE INDEX `+k.Prefix+`searchtab_tsgi ON `+k.Prefix+`searchtab USING gist (tsbody)`)
	return
}
func (k *KeyBuild) Insert(prid,body string,meta interface{}) {
	jsmeta,_ := json.MarshalToString(meta)
	k.DB.Exec(`
	INSERT INTO `+k.Prefix+`searchtab
		(prid,tsbody,txbody,jsmeta) VALUES
		($1,to_tsvector('simple',$2),$2,$3)
	`,prid,body,jsmeta)
}
func (k *KeyBuild) InsertPassive(prid,body string,meta interface{},isactive bool) {
	jsmeta,_ := json.MarshalToString(meta)
	k.DB.Exec(`
	INSERT INTO `+k.Prefix+`searchtab
		(prid,tsbody,txbody,jsmeta) VALUES
		($1,to_tsvector('simple',$2),$2,$3,$4)
	`,prid,body,jsmeta,isactive)
}
func (k *KeyBuild) Query(qstring string,f func(ke,va string)) {
	rows,err := k.DB.Query(`
	SELECT prid,jsmeta FROM `+k.Prefix+`searchtab
		WHERE tsbody @@ to_tsquery('simple',$1)
		AND isactive
	`,qstring)
	if err!=nil { fmt.Println(err) ; return }
	defer rows.Close()
	var ke,va string
	for rows.Next() {
		rows.Scan(&ke,&va)
		f(ke,va)
	}
}
func (k *KeyBuild) Read(prid string) (rec string,err error) {
	var txbody,jsmeta string
	var isactive bool
	err = k.DB.QueryRow(`SELECT txbody,jsmeta,isactive FROM `+k.Prefix+`searchtab WHERE prid=$1`,prid).Scan(&txbody,&jsmeta,&isactive)
	a,_ := json.MarshalToString(prid)
	b,_ := json.MarshalToString(txbody)
	if err==nil { rec = fmt.Sprintf(`{"prid": %s, "body": %s, "meta": %s, "isactive": %v}`,a,b,jsmeta,isactive) }
	return
}

//func (k *KeyBuild) 


