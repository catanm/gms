package endpoints

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/LarryBattle/nonce-golang"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/goweb"
	"github.com/stretchr/goweb/context"
	"gms/log"
	"gms/models"
	"gms/utils"
	"golang.org/x/crypto/bcrypt"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var color = []string{"ffff0000", "ff00ff00", "ff0000ff", "ffffff00", "ffffffff", "ff000000", "ff246b24", "ff990000", "ff669933", "ff660099"}

func Root(ctx context.Context) error {
	file, err := ioutil.ReadFile("../index.html")
	if err != nil {
		log.Error("Error reading index.html: " + err.Error())
		return goweb.Respond.With(ctx, http.StatusOK, []byte(`<html><body><h2>Log in with...</h2>
		<ul><li><a href="/api/auth/facebook/login/">Facebook</a></li></ul></body></html>`))
	}
	return goweb.Respond.With(ctx, http.StatusOK, file)
}

func Connect(ctx context.Context) error {
	provider_type := ctx.PathValue("provider")
	action := ctx.PathValue("action")

	if provider_type == "facebook" {
		provider, err := gomniauth.Provider(provider_type)
		if err != nil {
			log.Error("Error on getting provider: " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"An error has occured."})
		}
		state := gomniauth.NewState("after", "success")
		// if you want to request additional scopes from the provider,
		// pass them as login?scope=scope1,scope2
		//options := objx.MSI("scope", ctx.QueryValue("scope"))
		authUrl, err := provider.GetBeginAuthURL(state, nil)
		if err != nil {
			log.Error("Error on getting url: " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"An error has occured."})
		}
		// redirect
		return goweb.Respond.WithRedirect(ctx, authUrl)
	} else if provider_type == "local" && ctx.MethodString() == "POST" {
		// This is taken care of in separate functions.
		// Local login only with POST
		if action == "login" {
			return nil
		} else if action == "register" {
			return nil
		} else if action == "connect" {
			return nil
		} else {
			return goweb.API.Respond(ctx, 200, nil, []string{"Invalid action."})
		}
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Invalid provider type."})
	}
}

func Callback(ctx context.Context) error {
	provider_type := ctx.PathValue("provider")
	provider, err := gomniauth.Provider(provider_type)
	if err != nil {
		log.Error("Error on getting provider: " + err.Error())
		return goweb.API.Respond(ctx, 200, nil, []string{"An error has occured."})
	}
	creds, err := provider.CompleteAuth(ctx.QueryParams())
	if err != nil {
		log.Error("Error on completing auth: " + err.Error())
		return goweb.API.Respond(ctx, 200, nil, []string{"An error has occured."})
	}
	// load the user
	// https://github.com/stretchr/gomniauth/blob/master/common/user.go
	user, userErr := provider.GetUser(creds)
	if userErr != nil {
		log.Error("Error on getting user: " + userErr.Error())
		return goweb.API.Respond(ctx, 200, nil, []string{"An error has occured."})
	}

	passport := m.Passport{}
	err = m.GetDB("Passport").
		Find(bson.M{"provider": provider_type,
		"identifier": user.IDForProvider(provider_type)}).
		One(&passport)
	if err != nil {
		if err.Error() == "not found" {
			var currentUser, ok = ctx.Data()["user"].(m.User)
			if ok {
				err = m.GetDB("Passport").
					Insert(&m.Passport{bson.NewObjectId(), currentUser.Id, "", provider_type,
					user.IDForProvider(provider_type), fmt.Sprintf("%v", creds.Map.Get("access_token").Data()),
					fmt.Sprintf("%v", creds.Map.Get("refresh_token").Data())})
				if err != nil {
					log.Error("Error on registration with provider " + provider_type + ", new passport: " + err.Error())
					return goweb.API.Respond(ctx, 200, nil, []string{"Could not create your new authorization."})
				}
				log.Info("Connecting user")
				url, _ := url.Parse(utils.EnvUrl() + "/#/fblogin/?token=" + currentUser.Token)
				return goweb.Respond.WithRedirect(ctx, url)
			} else {
				// No user, create user, create passport
				var token = nonce.NewToken()
				nonce.MarkToken(token)
				newUser := m.User{bson.NewObjectId(), user.Nickname(), user.Email(), true, token, time.Now()}
				err = m.GetDB("User").Insert(&newUser)
				if err != nil {
					log.Error("Error on registration with provider " + provider_type + ", new user: " + err.Error())
					return goweb.API.Respond(ctx, 200, nil, []string{"Failed to register."})
				}
				err = m.GetDB("Passport").
					Insert(&m.Passport{bson.NewObjectId(), newUser.Id, "", provider_type,
					user.IDForProvider(provider_type), fmt.Sprintf("%v", creds.Map.Get("access_token").Data()),
					fmt.Sprintf("%v", creds.Map.Get("refresh_token").Data())})
				if err != nil {
					log.Error("Error on registration with provider " + provider_type + ", new user passport: " + err.Error())
					return goweb.API.Respond(ctx, 200, nil, []string{"Failed to create your new passport."})
				}
				log.Info("New user registered")
				url, _ := url.Parse(utils.EnvUrl() + "/#/fblogin/?token=" + newUser.Token)
				return goweb.Respond.WithRedirect(ctx, url)
			}
		} else {
			log.Error("Error on registration with provider " + provider_type + ", new passport: " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Could not find your authorization."})
		}
	} else {
		// login the user
		var user = m.User{}
		fmt.Println(passport)
		err = m.GetDB("User").Find(bson.M{"_id": passport.User}).One(&user)
		if err != nil {
			log.Error("Error on login with provider " + provider_type + ", user query: " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Could not find you on the database."})
		}
		log.Info("Found user returning id")
		url, _ := url.Parse(utils.EnvUrl() + "/#/fblogin/?token=" + user.Token)
		return goweb.Respond.WithRedirect(ctx, url)
	}
}

func Register(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		return goweb.API.RespondWithData(ctx, currentUser.Token)
	} else {
		username := ctx.FormValue("username")
		email := ctx.FormValue("email")
		password := ctx.FormValue("password")

		// Check or basic auth header
		authHeader := ctx.HttpRequest().Header.Get("Authorization")
		if len(authHeader) > 0 {
			info, err := base64.StdEncoding.DecodeString(strings.Split(authHeader, " ")[1])
			if err == nil {
				decryptedInfo := strings.Split(string(info), ":")
				username = decryptedInfo[0]
				password = decryptedInfo[1]
				email = decryptedInfo[2]
			}
		}

		if len(username) < 3 {
			return goweb.API.Respond(ctx, 200, nil, []string{"The username has to be at least 4 characters long."})
		}
		if len(email) == 0 {
			return goweb.API.Respond(ctx, 200, nil, []string{"Please supply an email address."})
		}

		err := utils.CheckValidPassword(password)
		if err != nil {
			return goweb.API.Respond(ctx, 200, nil, []string{err.Error()})
		}

		var oldUser m.User
		err = m.GetDB("User").Find(bson.M{"username": username}).One(&oldUser)
		if err == nil {
			log.Debug("Username already taken.")
			return goweb.API.Respond(ctx, 200, nil, []string{"Username already taken."})
		}

		var token = nonce.NewToken()
		nonce.MarkToken(token)
		newUser := m.User{bson.NewObjectId(), username, email, true, token, time.Now()}
		err = m.GetDB("User").Insert(&newUser)
		if err != nil {
			log.Error("Error on local registration, new user: " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to register."})
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)

		err = m.GetDB("Passport").Insert(&m.Passport{bson.NewObjectId(), newUser.Id, string(hashedPassword), "local", "", "", ""})
		if err != nil {
			log.Error("Error on local registration, new user passport: " + err.Error())
			m.GetDB("User").RemoveId(newUser.Id)
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to create your new passport."})
		}
		log.Info("New user registered local")
		return goweb.API.RespondWithData(ctx, newUser.Token)
	}
}

func Login(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		return goweb.API.RespondWithData(ctx, currentUser.Token)
	} else {
		username := ctx.FormValue("username")
		password := ctx.FormValue("password")

		// Check or basic auth header
		authHeader := ctx.HttpRequest().Header.Get("Authorization")
		if len(authHeader) > 0 {
			info, err := base64.StdEncoding.DecodeString(strings.Split(authHeader, " ")[1])
			if err == nil {
				decryptedInfo := strings.Split(string(info), ":")
				username = decryptedInfo[0]
				password = decryptedInfo[1]
			}
		}

		if len(username) < 3 || len(password) < 6 {
			return goweb.API.Respond(ctx, 200, nil, []string{"Invalid data supplied, please fill all fields."})
		}

		user := m.User{}
		query := bson.M{}
		if strings.Contains(username, "@") {
			query["username"] = username
		} else {
			query["email"] = username
		}
		err := m.GetDB("User").Find(query).One(&user)
		if err != nil {
			log.Error("Error in find user on login " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"No such username, please register."})
		}
		passport := m.Passport{}
		err = m.GetDB("Passport").Find(bson.M{"provider": "local", "user": user.Id}).One(&passport)
		if err != nil {
			log.Error("Error in find user passport on login " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"You have registered through Facebook."})
		}

		err = bcrypt.CompareHashAndPassword([]byte(passport.Password), []byte(password))
		if err != nil {
			return goweb.API.Respond(ctx, 200, nil, []string{"Incorrect password."})
		}
		log.Info("User found on local login")
		if user.Token == "" || user.TokenUpdate.Before(time.Now().AddDate(0, 0, -1)) {
			var token = nonce.NewToken()
			nonce.MarkToken(token)
			user.Token = token
			user.TokenUpdate = time.Now()

			err := m.GetDB("User").UpdateId(user.Id, user)
			if err != nil {
				log.Error("Failed to update token: " + err.Error())
			}
		}
		ctx.HttpResponseWriter().Header().Add("cookie", "token="+user.Token)
		return goweb.API.RespondWithData(ctx, user.Token)
	}
}

func GetUserInfo(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		return goweb.API.RespondWithData(ctx, currentUser)
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Please log in."})
	}
}

func UploadVideo(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		file, _, err := ctx.HttpRequest().FormFile("file")
		if err != nil {
			log.Error("Error on image upload FormFile " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the image."})
		}
		defer file.Close()

		outputPath := "../uploads/" + currentUser.Username + "/" + base64.StdEncoding.EncodeToString([]byte(currentUser.Username+time.Now().String()))

		if err := os.MkdirAll("../uploads/"+currentUser.Username+"/", 0777); err != nil {
			log.Error("Error in creating output file " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the image."})
		}

		outputFile, err := os.Create(outputPath)
		if err != nil {
			log.Error("Error in creating output file " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the image."})
		}
		defer outputFile.Close()

		if _, err = io.Copy(outputFile, file); err != nil {
			log.Error("Error in copying file " + err.Error())
			os.Remove(outputPath)
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the image."})
		}
		if err = m.GetDB("Video").Insert(&m.Video{currentUser.Id, outputPath, time.Now()}); err != nil {
			log.Error("Error in saving video file to db " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the image."})
		}
		return goweb.API.Respond(ctx, 200, nil, nil)
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Please log in to upload your videos."})
	}
}

func GetVideos(ctx context.Context) error {
	var result []m.Video

	// Get the info from the context
	fromDate, fromDateExists := utils.ParseDateFromContext("from_", ctx)
	toDate, toDateExists := utils.ParseDateFromContext("to_", ctx)

	limit, err := strconv.Atoi(ctx.FormValue("limit"))
	if err != nil {
		limit = 5
	}

	skip, err := strconv.Atoi(ctx.FormValue("skip"))
	if err != nil {
		skip = 0
	}

	var currentUser, ok = ctx.Data()["user"].(m.User)
	isPersonal, err := strconv.ParseBool(ctx.FormValue("personal"))
	if err != nil {
		isPersonal = false
	}

	query := bson.M{}
	if ok && isPersonal {
		if fromDateExists || toDateExists {
			query["date"] = bson.M{"$gte": fromDate, "$lte": toDate}
			query["user"] = currentUser.Id
		} else {
			query["user"] = currentUser.Id
		}
	} else {
		isPersonal = false
		if fromDateExists || toDateExists {
			query["date"] = bson.M{"$gte": fromDate, "$lte": toDate}
		}
	}

	if !fromDateExists && !toDateExists {
		var lastVideo m.Image
		var oneQuery = bson.M{}
		if isPersonal {
			oneQuery["user"] = currentUser.Id
		}
		if err = m.GetDB("Video").Find(oneQuery).Sort("-date").One(&lastVideo); err == nil {
			query["date"] = bson.M{"$gte": lastVideo.Date.AddDate(0, 0, -1), "$lte": lastVideo.Date.AddDate(0, 0, 1)}
		}
	}

	if err := m.GetDB("Video").Find(query).Skip(skip).Limit(limit).All(&result); err != nil {
		log.Error("Failed to search for tracks: " + err.Error())
		return goweb.API.Respond(ctx, 200, nil, []string{"Failed to find the tracks."})
	}

	return goweb.API.RespondWithData(ctx, result)
}

func UploadImgData(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		file, _, err := ctx.HttpRequest().FormFile("file")
		if err != nil {
			log.Error("Error on image upload FormFile " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the image."})
		}
		defer file.Close()
		utils.ParseImageTable(currentUser, file)
		return goweb.API.Respond(ctx, 200, nil, nil)
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Please log in to upload your images."})
	}
}

func UploadZip(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {

		file, _, err := ctx.HttpRequest().FormFile("file")
		if err != nil {
			log.Error("Error on zip upload FormFile " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the zip."})
		}
		defer file.Close()

		outputZipPath := "../uploads/temp/" + base64.StdEncoding.EncodeToString([]byte(currentUser.Username+time.Now().String()))

		if err := os.MkdirAll("../uploads/temp/", 0777); err != nil {
			log.Error("Error in creating output dir " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the zip."})
		}

		outputZipFile, err := os.Create(outputZipPath)
		if err != nil {
			log.Error("Error in creating output file in zip " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the zip."})
		}
		defer outputZipFile.Close()

		if _, err = io.Copy(outputZipFile, file); err != nil {
			log.Error("Error in copying file " + err.Error())
			os.Remove(outputZipPath)
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the zip."})
		}
		var extention = ""
		if runtime.GOOS == "windows" {
			extention = ".exe"
		}

		extractPath := "../uploads/temp/" + base64.StdEncoding.EncodeToString([]byte(currentUser.Username+time.Now().String()))
		commandString := fmt.Sprintf(`7za%s e %s -o%s -p%s -aoa`, extention, outputZipPath, extractPath, "SuperSecurePassword")
		commandSlice := strings.Fields(commandString)
		commandCall := exec.Command(commandSlice[0], commandSlice[1:]...)
		// err = commandCall.Run()
		value, err := commandCall.Output()
		if err != nil {
			log.Error("in unarchiving zip file: " + err.Error())
			log.Error("Full info about the error: " + string(value))
		}

		go utils.ProcessZip(currentUser, extractPath)

		return goweb.API.Respond(ctx, 200, nil, nil)
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Please log in to upload a zip."})
	}
}

func DeleteImage(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		imageId := ctx.FormValue("imageId")

		if imageId == "" {
			return goweb.API.Respond(ctx, 200, nil, []string{"You have to specify an image to be deleted."})
		}

		var image m.Image

		err := m.GetDB("Image").FindId(bson.ObjectIdHex(imageId)).One(&image)
		if err != nil {
			log.Error("DeleteImage, remove query " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"You have to specify an image to be deleted."})
		}
		if image.User.Hex() != currentUser.Id.Hex() {
			return goweb.API.Respond(ctx, 200, nil, []string{"You cannot delete this image."})
		} else {
			m.GetDB("Image").RemoveId(image.Id)
			err = os.Remove(image.Url)
		}

		return goweb.API.RespondWithData(ctx, "Image deleted.")
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Please log in to delete an images."})
	}
}

func UploadImage(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		file, _, err := ctx.HttpRequest().FormFile("file")
		if err != nil {
			log.Error("Error on image upload FormFile " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the image."})
		}
		defer file.Close()

		err = utils.ProcessImage(currentUser, file, nil, true)

		if err != nil {
			return goweb.API.Respond(ctx, 200, nil, []string{err.Error()})
		}
		return goweb.API.Respond(ctx, 200, nil, nil)
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Please log in to upload your images."})
	}
}

func UploadImageTable(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		file, _, err := ctx.HttpRequest().FormFile("file")
		if err != nil {
			log.Error("Error on image table upload FormFile " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the image table."})
		}
		defer file.Close()

		err = utils.ParseImageTable(currentUser, file)

		if err != nil {
			return goweb.API.Respond(ctx, 200, nil, []string{err.Error()})
		}
		return goweb.API.Respond(ctx, 200, nil, nil)
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Please log in to upload your image table."})
	}
}

func GetImages(ctx context.Context) error {
	var result []m.Image

	// Get the info from the context
	fromDate, fromDateExists := utils.ParseDateFromContext("from_", ctx)
	toDate, toDateExists := utils.ParseDateFromContext("to_", ctx)

	searchType := ctx.FormValue("upload")
	if searchType != "" {
		searchType = "uploaded"
	} else {
		searchType = "date"
	}

	limit, err := strconv.Atoi(ctx.FormValue("limit"))
	if err != nil {
		limit = 20
	}

	skip, err := strconv.Atoi(ctx.FormValue("skip"))
	if err != nil {
		skip = 0
	}

	var currentUser, ok = ctx.Data()["user"].(m.User)
	isPersonal, err := strconv.ParseBool(ctx.FormValue("personal"))
	if err != nil {
		isPersonal = false
	}
	query := bson.M{}
	if ok && isPersonal {
		if fromDateExists || toDateExists {
			query[searchType] = bson.M{"$gte": fromDate, "$lte": toDate}
			query["user"] = currentUser.Id
		} else {
			query["user"] = currentUser.Id
		}
	} else {
		isPersonal = false
		if fromDateExists || toDateExists {
			query[searchType] = bson.M{"$gte": fromDate, "$lte": toDate}
		}
	}

	addLocation := true
	var lat, long, radius float64
	rad, err := strconv.Atoi(ctx.FormValue("radius"))
	if err != nil {
		addLocation = false
	}
	if addLocation {
		radius = float64(rad)
		lat, err = strconv.ParseFloat(ctx.FormValue("lat"), 64)
		if err != nil {
			addLocation = false
		}
		long, err = strconv.ParseFloat(ctx.FormValue("log"), 64)
		if err != nil {
			addLocation = false
		}
	}

	if addLocation {
		query["lat"] = bson.M{"$gt": lat - radius, "$lt": lat + radius}
		query["log"] = bson.M{"$gt": long - radius, "$lt": long + radius}
	}

	showAll := ctx.FormValue("allImages")
	if showAll == "" {
		query["show"] = true
	}
	keyMoments := ctx.FormValue("keyMoments")
	if keyMoments != "" {
		query["processed"] = true
	}

	if err = m.GetDB("Image").Find(query).Sort("-" + searchType).Skip(skip).Limit(limit).All(&result); err != nil {
		log.Error("Failed to search for images: " + err.Error())
		return goweb.API.Respond(ctx, 200, nil, []string{"Failed to find the images."})
	}
	return goweb.API.RespondWithData(ctx, result)
}

func UploadTrail(ctx context.Context) error {
	var currentUser, ok = ctx.Data()["user"].(m.User)
	if ok {
		// read the file from the request
		file, _, err := ctx.HttpRequest().FormFile("file")
		if err != nil {
			log.Error("Error on gps upload FormFile " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the file."})
		}
		defer file.Close()

		outputPath := "../uploads/" + currentUser.Username + "/gps/" + base64.StdEncoding.EncodeToString([]byte(currentUser.Username+time.Now().String()))

		if err := os.MkdirAll("../uploads/"+currentUser.Username+"/gps/", 0777); err != nil {
			log.Error("Error in creating output dir, gps " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the gps."})
		}

		outputFile, err := os.Create(outputPath)
		if err != nil {
			log.Error("Error in creating output file, gps " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the gps."})
		}
		defer outputFile.Close()

		if _, err = io.Copy(outputFile, file); err != nil {
			log.Error("Error in copying file, gps " + err.Error())
			os.Remove(outputPath)
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the gps."})
		}

		var latIndex int = 4
		var lonIndex int = 6
		var TimeDate bool = true
		//get rid of heading
		scanner := bufio.NewScanner(file)
		if scanner.Scan() {

			line := scanner.Text()
			lineRecords := strings.Split(line, ",")
			if strings.EqualFold(lineRecords[2], "DATE") {
				latIndex++
				lonIndex++
				TimeDate = false
			}
		}

		//initialise vars
		var lastLon string = ""
		var lastLat string = ""

		var track m.Track
		var lastTrackID string = ""
		var LastDateArray []string
		var currDateArray []string
		var date string
		var date2 string

		//first record
		if scanner.Scan() {
			record := scanner.Text()
			recordSlice := strings.Split(record, ",")
			if TimeDate {
				LastDateArray = strings.Split(recordSlice[2], " ")
			} else {
				date = recordSlice[2] + " " + recordSlice[3]
				LastDateArray = strings.Split(date, " ")
			}
			lastTrackID = strconv.FormatInt(utils.GetTrackID(LastDateArray, currDateArray), 10)

			track = m.Track{bson.NewObjectId(), currentUser.Id, lastTrackID, utils.ParseDateFromStrings(LastDateArray[0], LastDateArray[1]), []m.Coordinate{}, time.Now(), time.Now(), time.Now()}
			if strings.EqualFold(recordSlice[latIndex+1], "S") {
				recordSlice[latIndex] = "-" + recordSlice[latIndex]
			}
			if strings.EqualFold(recordSlice[lonIndex+1], "W") {
				recordSlice[lonIndex] = "-" + recordSlice[lonIndex]
			}
			coor := m.Coordinate{recordSlice[latIndex], recordSlice[lonIndex], utils.ParseDateFromStrings(LastDateArray[0], LastDateArray[1])}
			track.Coordinates = append(track.Coordinates, coor)

			lastLat = recordSlice[latIndex]
			lastLon = recordSlice[lonIndex]

		}

		//read records
		for scanner.Scan() { //if there is something left

			line := scanner.Text()        //read one line
			s := strings.Split(line, ",") //split to slice

			if TimeDate {
				currDateArray = strings.Split(s[2], " ")
			} else {
				date2 = s[2] + " " + s[3]
				currDateArray = strings.Split(date2, " ")
			}
			if (strings.EqualFold(lastLat, s[latIndex])) || (strings.EqualFold(lastLon, s[lonIndex])) {

				continue //if same lat or lon as previous row: skip it

			} else {

				lastLat = s[latIndex] //update last record
				lastLon = s[lonIndex]
				trackID := strconv.FormatInt(utils.GetTrackID(LastDateArray, currDateArray), 10)

				if strings.EqualFold(s[latIndex+1], "S") {
					s[latIndex] = "-" + s[latIndex]
				}
				if strings.EqualFold(s[lonIndex+1], "W") {
					s[lonIndex] = "-" + s[lonIndex]
				}

				if strings.EqualFold(lastTrackID, trackID) { //add new coordinate to current track
					track.Coordinates = append(track.Coordinates, m.Coordinate{s[latIndex], s[lonIndex], utils.ParseDateFromStrings(currDateArray[0], currDateArray[1])})
				} else {
					//store track and start a new one!
					min, max := utils.GetMinMaxDateFromCoordinates(track.Coordinates)
					doc := m.Track{Id: bson.NewObjectId(), User: currentUser.Id, TrackID: track.TrackID, Date: track.Date, Coordinates: track.Coordinates, MinDate: min, MaxDate: max, Uploaded: time.Now()}
					err = m.GetDB("Track").Insert(doc)
					if err != nil {
						log.Error("Can't insert track: " + err.Error())
					} else {
						utils.LinkGpsToImages(doc)
					}

					track = m.Track{bson.NewObjectId(), currentUser.Id, trackID, utils.ParseDateFromStrings(currDateArray[0], currDateArray[1]), []m.Coordinate{}, time.Now(), time.Now(), time.Now()}
					track.Coordinates = append(track.Coordinates, m.Coordinate{s[latIndex], s[lonIndex], utils.ParseDateFromStrings(currDateArray[0], currDateArray[1])})
					lastTrackID = trackID

				}
				LastDateArray = currDateArray
			}
		}

		//store last track
		min, max := utils.GetMinMaxDateFromCoordinates(track.Coordinates)
		doc := m.Track{Id: bson.NewObjectId(), User: currentUser.Id, TrackID: track.TrackID, Date: track.Date, Coordinates: track.Coordinates, MinDate: min, MaxDate: max}
		err = m.GetDB("Track").Insert(doc)
		if err != nil {
			log.Error("Can't insert track: " + err.Error())
			return goweb.API.Respond(ctx, 200, nil, []string{"Failed to upload the track."})
		}
		utils.LinkGpsToImages(doc)
		return goweb.API.Respond(ctx, 200, nil, nil)
	} else {
		return goweb.API.Respond(ctx, 200, nil, []string{"Please log in to upload your GPS data."})
	}
}

func GetTrails(ctx context.Context) error {
	var result []m.Track

	// Get the info from the context
	isHeatmap, err := strconv.ParseBool(ctx.FormValue("heatmap"))
	if err != nil {
		isHeatmap = false
	}

	isPersonal, err := strconv.ParseBool(ctx.FormValue("personal"))
	if err != nil {
		isPersonal = false
	}

	fromDate, fromDateExists := utils.ParseDateFromContext("from_", ctx)
	toDate, toDateExists := utils.ParseDateFromContext("to_", ctx)

	if !isPersonal {
		err = m.GetDB("Track").Find(bson.M{}).All(&result)

	} else if isPersonal {
		var query bson.M
		var currentUser, ok = ctx.Data()["user"].(m.User)
		if ok && !fromDateExists && !toDateExists {
			query = bson.M{}
			query["user"] = currentUser.Id
			err = m.GetDB("Track").Find(query).All(&result)

		} else if ok && fromDateExists && toDateExists {
			query = bson.M{"user": currentUser.Id, "date": bson.M{"$gte": fromDate, "$lte": toDate}}
			err = m.GetDB("Track").Find(query).All(&result)

		}
	}
	if !isHeatmap {
		mapName := "GpsTrails"

		var buffer bytes.Buffer

		buffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<kml xmlns=\"http://earth.google.com/kml/2.0\">\n<Document>\n<name>")
		buffer.WriteString(mapName)
		buffer.WriteString("</name>\n<description>Glasgow Memories Server</description>\n")
		for i := 0; i < len(color); i++ {

			buffer.WriteString("<Style id=\"line")
			buffer.WriteString(strconv.Itoa(i + 1))
			buffer.WriteString("\">\n<LineStyle>\n<color>")
			buffer.WriteString(color[i%len(color)])
			buffer.WriteString("</color>\n<width>4</width>\n</LineStyle>\n</Style>\n\n")
		}

		for i := 0; i < len(result); i++ {
			buffer.WriteString("<Placemark>\n<name>Track ")
			buffer.WriteString(result[i].Date.String())
			buffer.WriteString("</name>\n<styleUrl>#line")
			buffer.WriteString(strconv.Itoa(utils.GetStyle(strings.Split(result[i].Date.String()[0:10], "-"), len(color)) + 1))
			buffer.WriteString("</styleUrl>\n<LineString>\n<altitudeMode>relative</altitudeMode>\n<coordinates>\n")
			for j := 0; j < len(result[i].Coordinates); j++ {
				buffer.WriteString(result[i].Coordinates[j].Lon + "," + result[i].Coordinates[j].Lat + ",0\n")
			}
			buffer.WriteString("</coordinates>\n</LineString>\n</Placemark>\n")
		}
		buffer.WriteString("</Document>\n")
		buffer.WriteString("</kml>\n")
		ctx.HttpResponseWriter().Header().Set("Content-Type", "application/vnd.google-earth.kml+xml")
		ctx.HttpResponseWriter().Write([]byte(buffer.String()))
		return nil
		// return goweb.Respond.With(ctx, http.StatusOK, []byte(buffer.String()))
	} else {
		var formattedData [][]string
		for i := 0; i < len(result); i++ {
			for j := 0; j < len(result[i].Coordinates); j++ {
				formattedData = utils.Append(formattedData, []string{result[i].Coordinates[j].Lat, result[i].Coordinates[j].Lon})
			}
		}
		return goweb.API.RespondWithData(ctx, formattedData)
	}
}

func GetPopularLocations(ctx context.Context) error {
	var query bson.M
	isPersonal, err := strconv.ParseBool(ctx.FormValue("personal"))
	if err != nil {
		isPersonal = false
	}
	if isPersonal {
		var result m.PopularLocations
		var currentUser, ok = ctx.Data()["user"].(m.User)
		if ok {
			query = bson.M{"user": currentUser.Id}
			if err := m.GetDB("RecommendUser").Find(query).One(&result); err != nil {
				//respond with empty arrays it no recommendations/popular location for current user
				result = m.PopularLocations{PopularMarkers: make([]m.Marker, 0), Recommendations: make([]m.Recommendation, 0)}
				//fmt.Printf("%v  %v", err, result)
			}
			return goweb.API.RespondWithData(ctx, result)
		}
	}
	var trendResult []m.TrendingAllUsersEntry
	var recomResult []m.Recommendation
	m.GetDB("TrendingAllUsers").Find(query).All(&trendResult)
	m.GetDB("RecommendAllUsers").Find(query).All(&recomResult)
	result := m.AllUsersTrendingAndRecom{Trend: trendResult, Recom: recomResult}
	return goweb.API.RespondWithData(ctx, result)
}

func NotFound(ctx context.Context) error {
	return goweb.Respond.WithStatus(ctx, http.StatusNotFound)
}

func GetStats(ctx context.Context) error {
	var err error
	var pageType = ctx.FormValue("page")
	query := bson.M{}
	//barchartQuery := bson.M{}

	var totalFileSize int64 = 0

	var imageCount int = 0

	if pageType == "pstats" {

		var currentUser, ok = ctx.Data()["user"].(m.User)
		if ok {
			fmt.Println("USER OK")
			query["user"] = currentUser.Id
			//barchartQuery["user"] = currentUser.Id

			var identifier = currentUser.Username
			if identifier == "" {
				identifier = currentUser.Email
			}

			var dirs []os.FileInfo
			dirs, err = ioutil.ReadDir("../uploads/" + identifier)
			if err != nil {
				log.Error("Failed to read dir:" + err.Error())
			}
			fmt.Println("ID:" + identifier)
			for i := 0; i < len(dirs); i++ {
				totalFileSize += dirs[i].Size()
			}
			imageCount = len(dirs)
		} else {
			return goweb.API.Respond(ctx, 200, nil, []string{"Please log in to view your stats."})
		}
	} else {
		var subdirInfo []os.FileInfo
		var dirInfo []os.FileInfo

		dirInfo, err = ioutil.ReadDir("../uploads/")
		if err != nil {
			log.Error("Failed to read dir:" + err.Error())
		}

		for i := 0; i < len(dirInfo); i++ {
			if dirInfo[i].IsDir() {
				subdirInfo, err = ioutil.ReadDir("../uploads/" + dirInfo[i].Name())
				if err != nil {
					log.Error("Failed to read dir:" + err.Error())
				}
				for j := 0; j < len(subdirInfo); j++ {
					totalFileSize += subdirInfo[j].Size()
				}
				imageCount += len(subdirInfo)
			}
		}
	}

	// convert to MB
	var fileSizeInMB = strconv.FormatFloat((float64(totalFileSize) / 1000000), 'f', 2, 32)
	fmt.Println(fileSizeInMB)

	// Get the info from the context
	fromDate, fromDateExists := utils.ParseDateFromContext("from_", ctx)
	toDate, toDateExists := utils.ParseDateFromContext("to_", ctx)

	//Total number of users
	userCount, err := m.GetDB("User").Count()
	if err != nil {
		log.Error("Failed to find user count: " + err.Error())
	}

	//Total number of tracks
	trailCount, err := m.GetDB("Track").Find(query).Count()
	if err != nil {
		log.Error("Failed to find track count: " + err.Error())
	}

	// Query add date/time
	if fromDateExists || toDateExists {
		query["date"] = bson.M{"$gte": fromDate, "$lte": toDate}
	}

	//Filtered images and tracks by date/time and user
	filteredImageCount, err := m.GetDB("Image").Find(query).Count()
	if err != nil {
		log.Error("Failed to find image count: " + err.Error())
	}
	filteredTrackCount, err := m.GetDB("Track").Find(query).Count()
	if err != nil {
		log.Error("Failed to find track count: " + err.Error())
	}

	fmt.Print("FILTERED IMAGE AND TRACK COUNT: ")
	fmt.Print(filteredImageCount)
	fmt.Print(" ")
	fmt.Print(filteredTrackCount)

	fmt.Println("")
	fmt.Println("")

	// Barchart number of images
	var dDay, dMonth, dYear = getDifference(fromDate, toDate)

	var imageCurrentCount int
	var imageCountBarchartValues []int
	var gpsCurrentCount int
	var gpsCountBarchartValues []int
	var currentLabel string
	var barchartLabels []string

	var nextDate, endNextDate time.Time

	var February = leapYearDays(fromDate.Year())
	var months = [13]int{0, 31, February, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

	if dYear < 1 {
		if dMonth < 1 { // days view

			for i := 0; i <= dDay; i++ {

				if i == 0 {
					nextDate, endNextDate = getDates(fromDate, 0, 0)
				}
				query["date"] = bson.M{"$gte": nextDate, "$lte": endNextDate}

				imageCurrentCount, err = m.GetDB("Image").Find(query).Count()
				if err != nil {
					log.Error("Failed to find image count: " + err.Error())
				}
				imageCountBarchartValues = append(imageCountBarchartValues, imageCurrentCount)

				gpsCurrentCount, err = m.GetDB("Track").Find(query).Count()
				if err != nil {
					log.Error("Failed to find gps count: " + err.Error())
				}
				gpsCountBarchartValues = append(gpsCountBarchartValues, gpsCurrentCount)

				currentLabel = strconv.Itoa(nextDate.Day()) + " " + nextDate.Month().String() + " " + strconv.Itoa(nextDate.Year())
				barchartLabels = append(barchartLabels, currentLabel)

				nextDate, endNextDate = getDates(nextDate.AddDate(0, 0, 1), 0, 0)
			}

		} else { // months view

			for i := 0; i <= dMonth; i++ {

				if i == 0 { //first third
					nextDate, endNextDate = getDates(fromDate, 0, months[int(fromDate.Month())])

				} else if i == dMonth { //last third
					nextDate = nextDate.AddDate(0, 1, 1-nextDate.Day())
					nextDate, endNextDate = getDates(nextDate, 0, toDate.Day())

				} else { // everything else from the beginning to the end of month
					nextDate = nextDate.AddDate(0, 1, 1-nextDate.Day())
					mEnd := months[int(nextDate.Month())]
					nextDate, endNextDate = getDates(nextDate, 0, mEnd)
				}

				query["date"] = bson.M{"$gte": nextDate, "$lte": endNextDate}

				imageCurrentCount, err = m.GetDB("Image").Find(query).Count()
				if err != nil {
					log.Error("Failed to find image count: " + err.Error())
				}
				imageCountBarchartValues = append(imageCountBarchartValues, imageCurrentCount)

				gpsCurrentCount, err = m.GetDB("Track").Find(query).Count()
				if err != nil {
					log.Error("Failed to find gps count: " + err.Error())
				}
				gpsCountBarchartValues = append(gpsCountBarchartValues, gpsCurrentCount)

				currentLabel = nextDate.Month().String() + " " + strconv.Itoa(nextDate.Year())
				barchartLabels = append(barchartLabels, currentLabel)

			}
		}

	} else { // years view

		for i := 0; i <= dYear; i++ {

			if i == 0 { //first third
				nextDate, endNextDate = getDates(fromDate, 12, months[int(fromDate.Month())]) //31?

			} else if i == dMonth { //last third
				nextDate = nextDate.AddDate(1, 1-int(nextDate.Month()), 1-nextDate.Day())
				nextDate, endNextDate = getDates(nextDate, int(toDate.Month()), toDate.Day())

			} else { // everything else from the beginning to the end of year
				nextDate = nextDate.AddDate(1, 1-int(nextDate.Month()), 1-nextDate.Day())
				nextDate, endNextDate = getDates(nextDate, 12, months[int(nextDate.Month())]) //31?
			}

			query["date"] = bson.M{"$gte": nextDate, "$lte": endNextDate}

			imageCurrentCount, err = m.GetDB("Image").Find(query).Count()
			if err != nil {
				log.Error("Failed to find image count: " + err.Error())
			}
			imageCountBarchartValues = append(imageCountBarchartValues, imageCurrentCount)

			gpsCurrentCount, err = m.GetDB("Track").Find(query).Count()
			if err != nil {
				log.Error("Failed to find gps count: " + err.Error())
			}
			gpsCountBarchartValues = append(gpsCountBarchartValues, gpsCurrentCount)

			currentLabel = nextDate.Month().String() + " " + strconv.Itoa(nextDate.Year())
			barchartLabels = append(barchartLabels, currentLabel)
		}

	}

	var stats = []int{userCount, imageCount, trailCount, filteredImageCount, filteredTrackCount}
	var result []interface{} = []interface{}{imageCountBarchartValues, barchartLabels, gpsCountBarchartValues, stats, fileSizeInMB}

	return goweb.API.RespondWithData(ctx, result)
}

func leapYearDays(year int) int {
	if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
		return 29
	} else {
		return 28
	}
}

func getDifference(fromDate time.Time, toDate time.Time) (int, int, int) {
	return absoluteValue(toDate.Day() - fromDate.Day()),
		absoluteValue((int(toDate.Month()) - int(fromDate.Month()))),
		absoluteValue(toDate.Year() - fromDate.Year())
}

func absoluteValue(value int) int {
	if value < 0 {
		return value * (-1)
	} else {
		return value
	}
}

func getDates(date time.Time, endYear int, endMonth int) (time.Time, time.Time) {
	var m int
	if endMonth == 0 {
		m = date.Day()
	} else {
		m = endMonth
	}

	var y time.Month
	if endYear == 0 {
		y = date.Month()
	} else {
		y = time.Month(endYear)
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC),
		time.Date(date.Year(), y, m, 23, 59, 0, 0, time.UTC)
}
