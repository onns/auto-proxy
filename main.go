package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
)

type GlobalConfig struct {
	Port string              `json:"port"`
	Rule map[string]*UrlRule `json:"rule"`
}

type UrlRule struct {
	Type string `json:"type"`
	Dest string `json:"dest"`
}

var OnnsGlobal GlobalConfig

func loadConfig() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	filename := path.Join(exPath, "config.json")
	if _, err = os.Stat(filename); err != nil {
		panic(err)
	}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(b, &OnnsGlobal)
}

func init() {
	loadConfig()
}

type baseHandle struct{}

func (h *baseHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Println(r.URL.String())
	remoteRole, ok := OnnsGlobal.Rule[r.URL.Path]
	remoteUrl := &url.URL{}
	if !ok {
		remoteUrl = r.URL
	} else {
		remoteUrl, _ = url.Parse(remoteRole.Dest)
	}

	log.Println(remoteUrl.String())
	proxy := httputil.NewSingleHostReverseProxy(remoteUrl)
	proxy.Director = func(req *http.Request) {
		req.URL.RawQuery = remoteUrl.RawQuery
		req.URL.Scheme = remoteUrl.Scheme
		req.URL.Host = remoteUrl.Host
		req.Host = remoteUrl.Host
		req.URL.Path = remoteUrl.Path
		req.URL.RawPath = remoteUrl.RawPath
		if _, ok = req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
	}

	proxy.ModifyResponse = func(response *http.Response) error {
		response.Header.Add("Access-Control-Allow-Origin", "*")
		return nil
	}
	proxy.ServeHTTP(w, r)
	return
}
func main() {

	h := &baseHandle{}
	http.Handle("/", h)

	server := &http.Server{
		Addr:    OnnsGlobal.Port,
		Handler: h,
	}
	log.Fatal(server.ListenAndServe())
}
