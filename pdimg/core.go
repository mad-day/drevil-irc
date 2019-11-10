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



package pdimg

import (
	jsoniter "github.com/json-iterator/go"
	shell "github.com/ipfs/go-ipfs-api"
	"io"
	"io/ioutil"
	"bytes"
)

var json = jsoniter.ConfigFastest

func drainClose(r io.ReadCloser) {
	defer r.Close()
	io.Copy(ioutil.Discard,r)
}

type PDImageInfo struct {
	/* Technical Data */
	Type string `json:"type"`
	Fext string `json:"fext"`
	Width   int `json:"width,omitempty"`
	Height  int `json:"height,omitempty"`
	
	/* Meta Data */
	Title       string `json:"title,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Album       string `json:"album,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Description string `json:"description,omitempty"`
	
}

type Poster interface {
	CreatePDI(meta, img io.Reader) (key string,err error)
	LocallyPin(key string) (err error)
}
type Obtainer interface {
	CatMeta(key string) (io.ReadCloser,error)
	CatImage(key string) (io.ReadCloser,error)
}
type PinningDB interface {
	PinKey(key string) error
}
var Noop_PinningDB PinningDB = fakePinning{}
type fakePinning struct{}
func (fakePinning) PinKey(key string) error { return nil }
func (fakePinning) String() string { return "{Noop_PinningDB}" }

type Facade struct {
	PinningDB
	Poster
	Obtainer
}
func (f *Facade) InsertImage(meta *PDImageInfo, img io.Reader) (key string,err error) {
	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(meta)
	if err!=nil { return }
	key,err = f.CreatePDI(buf,img)
	if err!=nil { return }
	err = f.PinKey(key)
	if err!=nil { return }
	err = f.LocallyPin(key)
	return
}
func (f *Facade) GetMetadata(key string) (*PDImageInfo,error) {
	rc,err := f.CatMeta(key)
	if err!=nil { return nil,err }
	defer drainClose(rc)
	meta := new(PDImageInfo)
	err = json.NewDecoder(rc).Decode(meta)
	return meta,err
}


type ShellPoster struct {
	Sh *shell.Shell
}
func (sp *ShellPoster) CreatePDI(meta, img io.Reader) (key string,err error) {
	var dk,mk string
	dk,err = sp.Sh.NewObject("unixfs-dir")
	if err!=nil { return }
	mk,err = sp.Sh.Add(meta,shell.Pin(false))
	if err!=nil { return }
	dk,err = sp.Sh.PatchLink(dk, "meta", mk, true)
	if err!=nil { return }
	mk,err = sp.Sh.Add(img,shell.Pin(false))
	if err!=nil { return }
	dk,err = sp.Sh.PatchLink(dk, "image", mk, true)
	if err!=nil { return }
	key = dk
	return
}
func (sp *ShellPoster) LocallyPin(key string) (err error) {
	return sp.Sh.Pin("/ipfs/"+key)
}
func (sp *ShellPoster) CatMeta(key string) (io.ReadCloser,error) {
	return sp.Sh.Cat("/ipfs/"+key+"/meta")
}
func (sp *ShellPoster) CatImage(key string) (io.ReadCloser,error) {
	return sp.Sh.Cat("/ipfs/"+key+"/image")
}

