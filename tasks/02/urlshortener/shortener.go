package urlshortener

import (
	"net/http"
	"net/url"
	"github.com/go-chi/chi"
	"time"
    "math/rand"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

type URLShortener struct {
	aliases map[string] string
	addr string
}

func NewShortener(addr string) *URLShortener {
	return &URLShortener{
		aliases: make(map[string] string),
		addr: addr,
	}
}

func (s *URLShortener) HandleSave(rw http.ResponseWriter, req *http.Request) {
	var init_url string = req.URL.Query().Get("u")
	var key string = randSeq(10)
	_, err := url.Parse(key)
	if err != nil{
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	_, ok := s.aliases[key]
	if ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	s.aliases[key] = init_url
	rw.Write([]byte(s.addr + "/" + key))
}

func (s *URLShortener) HandleExpand(rw http.ResponseWriter, req *http.Request) {
	key := chi.URLParam(req, "key")
	init_addr, ok := s.aliases[key]
	if !ok {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	http.Redirect(rw, req, init_addr, http.StatusMovedPermanently)
}