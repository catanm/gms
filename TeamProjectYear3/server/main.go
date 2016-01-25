package main

import (
	"github.com/stretchr/codecs/services"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/goweb"
	"github.com/stretchr/goweb/context"
	"github.com/stretchr/goweb/responders"
	"github.com/stretchr/signature"
	"gms/endpoints"
	"gms/log"
	"gms/models"
	"gms/utils"
	"labix.org/v2/mgo/bson"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {

	log.Info("Glasgow Memories Server")
	log.Info("=======================")

	utils.InitEnv()
	var Address = ":" + utils.EnvPort()
	var baseURL = utils.EnvUrl()

	m.Connect()
	defer m.Close()

	// prepare the decryption key
	if utils.LoadCypherKey() != nil {
		log.Error("Failed to load the decryption key.")
		return
	}

	// GOMNIAUTH
	gomniauth.SetSecurityKey(signature.RandomKey(64))
	gomniauth.WithProviders(
		facebook.New("1497244403859030", "fbbb08c47e0441bcf23ea82b5f340fe5",
			baseURL+"/api/auth/facebook/callback/"),
	)

	// Attach the DB collection references to the context in order to pass it around
	goweb.MapBefore(func(ctx context.Context) error {
		var user = m.User{}
		cookieC, err := ctx.HttpRequest().Cookie("token")
		var cookie string
		if err != nil {
			cookie = ctx.FormValue("token")
			if cookie == "" {
				return nil
			}
		} else {
			cookie = cookieC.Value
		}
		err = m.GetDB("User").Find(bson.M{"token": cookie}).One(&user)
		if err != nil {
			// log.Info("MapBefore 2 " + err.Error())
			return nil
		}
		ctx.Data()["user"] = user
		return nil
	})

	goweb.MapStatic("/static", "../static")   // This is the directory with all static UI files
	goweb.MapStatic("/uploads", "../uploads") // This is the directory where we should store uploaded files

	// ENDPOINTS
	goweb.Map("GET", "/", endpoints.Root)
	goweb.Map("POST", "api/auth/local/register", endpoints.Register)
	goweb.Map("POST", "api/auth/local/login", endpoints.Login)
	goweb.Map("GET", "api/auth/{provider}/callback", endpoints.Callback)
	goweb.Map([]string{"GET", "POST"}, "api/auth/{provider}/{action}", endpoints.Connect)
	goweb.Map("POST", "api/upload/image", endpoints.UploadImage)
	goweb.Map("GET", "api/images/get", endpoints.GetImages)
	goweb.Map("POST", "api/upload/csv", endpoints.UploadTrail)
	goweb.Map("GET", "api/trails/get", endpoints.GetTrails)
	goweb.Map("POST", "api/upload/video", endpoints.UploadVideo)
	goweb.Map("GET", "api/videos/get", endpoints.GetVideos)
	goweb.Map("GET", "api/user", endpoints.GetUserInfo)
	goweb.Map("GET", "api/stats/get", endpoints.GetStats)
	goweb.Map("GET", "api/popLocations", endpoints.GetPopularLocations)
	goweb.Map("POST", "api/upload/imagetable", endpoints.UploadImageTable)
	goweb.Map("POST", "api/upload/zip", endpoints.UploadZip)
	// TODO: Add new endpoints here

	goweb.Map(endpoints.NotFound)

	// Remove the information from the data just in case the call is intercepted
	goweb.MapAfter(func(ctx context.Context) error {
		ctx.Data()["user"] = ""
		return nil
	})

	// setup the API responder
	codecService := services.NewWebCodecService()
	codecService.RemoveCodec("text/xml")
	apiResponder := responders.NewGowebAPIResponder(codecService, goweb.Respond)
	apiResponder.StandardFieldDataKey = "data"
	apiResponder.StandardFieldStatusKey = "status"
	apiResponder.StandardFieldErrorsKey = "errors"
	goweb.API = apiResponder

	// SERVER
	s := &http.Server{
		Addr:           Address,
		Handler:        goweb.DefaultHttpHandler(),
		ReadTimeout:    5 * time.Minute,
		WriteTimeout:   5 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	listener, listenErr := net.Listen("tcp", Address)
	log.Info("Server port: " + Address)
	log.Info("Server running at: " + baseURL + "\n")
	if listenErr != nil {
		log.Error("Could not listen: " + listenErr.Error())
	}

	go func() {
		for _ = range c {
			// sig is a ^C, handle it
			// stop the HTTP server
			log.Info("Stopping the server...\n")
			listener.Close()
			log.Info("Server stopped.\n")
		}
	}()
	// begin the server
	log.Error("Error in Serve: " + s.Serve(listener).Error())
}
