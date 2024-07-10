package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/atotto/clipboard"
	log "github.com/inconshreveable/log15"
	"github.com/lemonade-command/lemonade/lemon"
	"github.com/pocke/go-iprange"
	"github.com/skratchdot/open-golang/open"
)

var (
	logger     log.Logger
	lineEnding string
	ra         *iprange.Range
	port       int
	path       = "./files"
)

func handleCopy(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Copy only support post", http.StatusMethodNotAllowed)
		return
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("io.ReadAll error", "err", err.Error())

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	text := lemon.ConvertLineEnding(string(buf), lineEnding)

	logger.Debug("Copy:", "text", text)

	err = clipboard.WriteAll(text)
	if err != nil {
		logger.Error("clipboard.WriteAll error", "err", err.Error())
	}
}

func handlePaste(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Paste only support get", http.StatusMethodNotAllowed)
		return
	}

	text, err := clipboard.ReadAll()
	if err != nil {
		logger.Error("clipboard.ReadAll error", "err", err.Error())
		return
	}

	_, err = io.WriteString(w, text)
	if err != nil {
		logger.Error("io.WriteString error", "err", err.Error())
		return
	}

	logger.Debug("Paste: ", "text", text)
}

func translateLoopbackIP(uri string, remoteIP string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		return uri
	}

	host, port, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		return uri
	}

	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return uri
	}

	if len(port) == 0 {
		parsed.Host = remoteIP
	} else {
		parsed.Host = fmt.Sprintf("%s:%s", remoteIP, port)
	}

	return parsed.String()
}

func handleOpen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Open only support get", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	uri := q.Get("uri")
	isBase64 := q.Get("base64")
	if isBase64 == "true" {
		decodeURI, err := base64.URLEncoding.DecodeString(uri)
		if err != nil {
			logger.Error("base64 decode error", "uri", uri)
			return
		}
		uri = string(decodeURI)
	}

	transLoopback := q.Get("transLoopback")
	if transLoopback == "true" {
		remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)
		uri = translateLoopbackIP(uri, remoteIP)
	}

	logger.Info("Open: ", "uri", uri)
	open.Run(uri)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Upload only support post", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("uploadFile")
	if err != nil {
		http.Error(w, "Error Retrieving the File", 500)
		logger.Error("Error Retrieving the File", "err", err)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error Read the File", 500)
		logger.Error("Error Read the File", "err", err)
		return
	}

	os.WriteFile(path+"/"+handler.Filename, fileBytes, os.ModePerm)

	q := r.URL.Query()
	isOpen := q.Get("open")
	if isOpen == "true" {
		uri := fmt.Sprintf("http://127.0.0.1:%d/files/%s", port, handler.Filename)
		logger.Info("Open: ", "uri", uri)
		open.Run(uri)
	}
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Not support method.", http.StatusMethodNotAllowed)
			return
		}

		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "RemoteAddr error.", 500)
			return
		}
		if !ra.IncludeStr(remoteIP) {
			http.Error(w, "Not allow ip.", http.StatusServiceUnavailable)
			logger.Info("not in allow ip. from: ", remoteIP)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func Serve(c *lemon.CLI, _logger log.Logger) error {
	logger = _logger
	lineEnding = c.LineEnding
	port = c.Port

	var err error
	ra, err = iprange.New(c.Allow)
	if err != nil {
		logger.Error("allowIp error")
		return err
	}

	os.MkdirAll(path, os.ModePerm)

	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(path))))
	http.Handle("/copy", middleware(http.HandlerFunc(handleCopy)))
	http.Handle("/paste", middleware(http.HandlerFunc(handlePaste)))
	http.Handle("/open", middleware(http.HandlerFunc(handleOpen)))
	http.Handle("/upload", middleware(http.HandlerFunc(handleUpload)))

	err = http.ListenAndServe(fmt.Sprintf(":%d", c.Port), nil)
	if err != nil {
		return err
	}

	return nil
}
