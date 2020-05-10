/*
 * Copyright 2019 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	//"crypto/tls"
	"fmt"
	"github.com/AletheiaWareLLC/netgo"
	"github.com/AletheiaWareLLC/switchgo"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "start":
			if err := start(); err != nil {
				log.Println(err)
				return
			}
		default:
			log.Println("Cannot handle", os.Args[1])
		}
	} else {
		PrintUsage(os.Stdout)
	}
}

func start() error {
	store, ok := os.LookupEnv("LOG_DIRECTORY")
	if !ok {
		store = "logs"
	}
	if err := os.MkdirAll(store, os.ModePerm); err != nil {
		return err
	}
	logFile, err := os.OpenFile(path.Join(store, time.Now().Format(time.RFC3339)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Log File:", logFile.Name())

	certificates, ok := os.LookupEnv("CERTIFICATE_DIRECTORY")
	if !ok {
		certificates = "certificates"
	}
	log.Println("Certificate Directory:", certificates)

	html, ok := os.LookupEnv("HTML_DIRECTORY")
	if !ok {
		html = "html"
	}
	log.Println("HTML Directory:", html)

	routeMap := make(map[string]bool)

	routes, ok := os.LookupEnv("ROUTES")
	if ok {
		for _, route := range strings.Split(routes, ",") {
			routeMap[route] = true
		}
	}

	sw := &switchgo.Switch{
		Name:  "Light",
		State: "off",
		Next:  "on",
	}
	// Redirect HTTP Requests to HTTPS
	// go http.ListenAndServe(":80", http.HandlerFunc(netgo.HTTPSRedirect(routeMap)))

	// Serve Web Requests
	mux := http.NewServeMux()
	mux.HandleFunc("/", netgo.StaticHandler(path.Join(html, "static")))
	mux.HandleFunc("/state", StateHandler(sw))
	switchTemplate, err := template.ParseFiles(path.Join(path.Join(html, "template"), "switch.html"))
	if err != nil {
		return err
	}
	mux.HandleFunc("/switch", SwitchHandler(sw, switchTemplate))
	return http.ListenAndServe(":80", mux)
	/*
	   // Serve HTTPS Requests
	   config := &tls.Config{MinVersion: tls.VersionTLS10}
	   server := &http.Server{Addr: ":443", Handler: mux, TLSConfig: config}
	   return server.ListenAndServeTLS(path.Join(certificates, "fullchain.pem"), path.Join(certificates, "privkey.pem"))
	*/
}

func PrintUsage(output io.Writer) {
	fmt.Fprintln(output, "Switch Server Usage:")
	fmt.Fprintln(output, "\tserver - display usage")
	fmt.Fprintln(output)
	fmt.Fprintln(output, "\tserver start - starts the server")
}

func StateHandler(sw *switchgo.Switch) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path)
		switch r.Method {
		case "GET":
			count, err := w.Write([]byte(sw.State))
			if err != nil {
				log.Println(err)
				return
			} else {
				log.Println("Wrote", count, "bytes")
			}
		default:
			log.Println("Unsupported method", r.Method)
		}
	}
}

func SwitchHandler(sw *switchgo.Switch, template *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RemoteAddr, r.Proto, r.Method, r.Host, r.URL.Path)
		switch r.Method {
		case "POST":
			r.ParseForm()
			log.Println("Request", r)
			state := r.Form["state"]
			log.Println("State", state)
			if state == nil || len(state) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				sw.Switch(state[0])
			}
			fallthrough
		case "GET":
			data := struct {
				Name      string
				Timestamp string
				State     string
				Next      string
			}{
				Name:      sw.Name,
				Timestamp: time.Unix(0, int64(sw.Timestamp)).Format(time.RFC3339),
				State:     sw.State,
				Next:      sw.Next,
			}
			if err := template.Execute(w, data); err != nil {
				log.Println(err)
				return
			}
		default:
			log.Println("Unsupported method", r.Method)
		}
	}
}
