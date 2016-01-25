package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/stretchr/goweb/context"
	"gms/log"
	"gms/models"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ClusterThreshold = 9
	BlurThreshold    = 0.38
)

var mutex sync.Mutex

var defaultPort = "1337"
var defaultBaseUrl = "http://localhost"

var trackIDcounter int64 = 0
var envPort string
var envUrl string

var key []byte

func InitEnv() {
	envPort = os.Getenv("GMS_TP3_PORT")
	envUrl = os.Getenv("GMS_TP3_BASE_URL")
	if envPort == "" {
		envPort = defaultPort
	}
	if envUrl == "" {
		envUrl = defaultBaseUrl
	}

	if envPort != ":80" && envPort != ":443" {
		envUrl = envUrl + ":" + envPort
	}
}

func EnvPort() string {
	return envPort
}

func EnvUrl() string {
	return envUrl
}

func LinkGpsToImages(track m.Track) {
	var possibleImages []m.Image
	m.GetDB("Image").Find(bson.M{"date": bson.M{"$lte": track.MaxDate, "$gte": track.MinDate}, "user": track.User}).All(&possibleImages)

	if len(possibleImages) == 0 {
		return
	}

	var best m.Coordinate

	var before time.Time
	var after time.Time

	for i := 0; i < len(possibleImages); i++ {
		before = possibleImages[i].Date.Add(time.Duration(-10 * time.Second))
		after = possibleImages[i].Date.Add(time.Duration(10 * time.Second))

		for j := 0; j < len(track.Coordinates); j++ {
			if before.Before(track.Coordinates[j].Date) && after.After(track.Coordinates[j].Date) {
				best = track.Coordinates[j]
				break
			}
		}
		if best.Lat != "" && best.Lon != "" {
			possibleImages[i].Lat = best.Lat
			possibleImages[i].Lon = best.Lon
			err := m.GetDB("Image").UpdateId(possibleImages[i].Id, possibleImages[i])
			if err != nil {
				log.Error("LinkGpdToImages, update, " + err.Error())
			}
		}
	}
}

func ImagePostprocessing(image m.Image, imageWaitGroup *sync.WaitGroup) {
	// get the blur score

	var ext = ".sh"
	if runtime.GOOS == "windows" {
		ext = ".bat"
	}

	blurArr, err := exec.Command("./processing/blurScript"+ext, image.Url).Output()

	if err != nil {
		log.Error("ImagePostprocessing, get blur score: " + err.Error())
	}

	image.Blur, err = strconv.ParseFloat(strings.Trim(string(blurArr), "\n "), 64)
	if err != nil {
		log.Error("Find blur " + err.Error())
		image.Blur = 1
	}

	image.Show = image.Blur < BlurThreshold

	// get the phash
	// Old call
	phash, err := exec.Command("./processing/phashScript"+ext, image.Url).Output()

	if err != nil {
		log.Error("ImagePostprocessing, get phash score: " + err.Error())
	}

	image.Phash = string(phash)

	// hash, err := phash.ImageHashDCT("../" + image.Url)
	// if err != nil {
	// 	log.Error("LinkImageToGps, phash: " + err.Error())
	// 	image.Phash = ""
	// } else {
	// 	image.Phash = strconv.FormatUint(hash, 2)
	// 	log.Error(image.Phash)
	// }

	// cluster (by date + phash(humming distance) + user)
	var clusters []m.Cluster

	mutex.Lock()
	err = m.GetDB("Cluster").Find(bson.M{"user": image.User, "date": bson.M{"$gt": time.Now().AddDate(0, 0, -1), "$lt": time.Now().AddDate(0, 0, 1)}}).All(&clusters)
	if err != nil || len(clusters) == 0 {
		// create a new cluster
		CreateNewCluster(image)
	} else {
		found := false
		for index := range clusters {
			if Distance(clusters[index].Phash, image.Phash) < ClusterThreshold {
				// add it to the cluster
				image.Cluster = clusters[index].Id
				if !clusters[index].HasSelected && image.Show {
					clusters[index].HasSelected = true
					err = m.GetDB("Cluster").UpdateId(clusters[index].Id, clusters[index])
					if err != nil {
						log.Error("ImagePostprocessing, update cluster: " + err.Error())
					}
				} else {
					image.Show = false
				}
				found = true
			}
		}
		if !found {
			CreateNewCluster(image)
		}
	}
	mutex.Unlock()

	image.Processed = true

	// save the image
	err = m.GetDB("Image").UpdateId(image.Id, image)
	if err != nil {
		log.Error("ImagePostprocessing, update, " + err.Error())
	}
	if imageWaitGroup != nil {
		imageWaitGroup.Done()
	}
}

func LinkImageToGps(image m.Image, imageWaitGroup *sync.WaitGroup) {

	var possibleTracks []m.Track
	m.GetDB("Track").Find(bson.M{"minDate": bson.M{"$lte": image.Date}, "maxDate": bson.M{"$gte": image.Date}, "user": image.User}).All(&possibleTracks)

	var best m.Coordinate

	before := image.Date.Add(time.Duration(-10 * time.Second))
	after := image.Date.Add(time.Duration(10 * time.Second))

	for i := 0; i < len(possibleTracks); i++ {
		for j := 0; j < len(possibleTracks[i].Coordinates); j++ {
			if before.Before(possibleTracks[i].Coordinates[j].Date) && after.After(possibleTracks[i].Coordinates[j].Date) {
				best = possibleTracks[i].Coordinates[j]
				break
			}
		}
		if best.Lat != "" && best.Lon != "" {
			break
		}
	}

	// attach location
	if best.Lat != "" && best.Lon != "" {
		image.Lat = best.Lat
		image.Lon = best.Lon
	}

	// save the image
	err := m.GetDB("Image").UpdateId(image.Id, image)
	if err != nil {
		log.Error("LinkImageToGps, update, " + err.Error())
	}
	if imageWaitGroup != nil {
		imageWaitGroup.Done()
	}
}

func CreateNewCluster(image m.Image) {
	m.GetDB("Cluster").Insert(m.Cluster{image.Cluster, image.Phash, image.User, time.Now(), image.Show})
}

/** Compute hamming distance **/
func Distance(s1 string, s2 string) int {
	if s1 == "" || s2 == "" {
		return 32
	}
	var counter int = 0
	for k := 0; k < len(s1); k++ {
		if string(s1[k]) != string(s2[k]) {
			counter++
		}

	}
	return counter
}

func ParseImageTable(currentUser m.User, file io.Reader) error {

	outputPath := "../uploads/" + currentUser.Username + "/tables/" + base64.StdEncoding.EncodeToString([]byte(currentUser.Username+time.Now().String()))

	if err := os.MkdirAll("../uploads/"+currentUser.Username+"/tables/", 0777); err != nil {
		log.Error("Error in creating output dir " + err.Error())
		return errors.New("Failed to upload the image.")
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Error("Error in creating output file, process image " + err.Error())
		return errors.New("Failed to upload the image.")
	}
	defer outputFile.Close()

	if _, err = io.Copy(outputFile, file); err != nil {
		log.Error("Error in copying file " + err.Error())
		os.Remove(outputPath)
		return errors.New("Failed to upload the image.")
	}

	parser := csv.NewReader(file)
	parser.Read() // dump the header

	timeLayout := "2006-01-02T15:04:05-0700"

	for true {
		items, err := parser.Read()
		if err != nil {
			return errors.New("Failed to parse the image table")
			break
		}

		imageDate, err := time.Parse(timeLayout, items[0])
		if err != nil {
			log.Error("parse date from image table " + err.Error())
			continue
		}

		var image m.Image

		if err = m.GetDB("Image").Find(bson.M{"user": currentUser.Id, "date": imageDate}).One(&image); err != nil {
			if err.Error() != "not found" {
				log.Error("find image from image table " + err.Error())
			}
			continue
		} else {
			image.CamId = items[1]
			image.Size = items[2]
			image.Type = items[3]
			image.PropInfo = items[4]
			image.Accx = items[5]
			image.Accy = items[6]
			image.Accz = items[7]
			image.Magx = items[8]
			image.Magy = items[9]
			image.Magz = items[10]
			image.Red = items[11]
			image.Green = items[12]
			image.Blue = items[13]
			image.Lum = items[14]
			image.Tem = items[15]
			image.GpsStatus = items[16]
			// 17, 18 lat and long
			image.Alt = items[19]
			image.GS = items[20]
			image.Herr = items[21]
			image.Verr = items[22]
			image.Exp = items[23]
			image.Gain = items[24]
			image.RBal = items[25]
			image.GBal = items[26]
			image.BBal = items[27]
			image.Xor = items[28]
			image.Yor = items[29]
			image.Zor = items[30]
			image.Stags = items[31]
			image.Tags = items[32]

			// save the image
			err = m.GetDB("Image").UpdateId(image.Id, image)
			if err != nil {
				log.Error("LinkImageToGps, update, " + err.Error())
			}
		}
	}

	return nil
}

func ProcessZip(currentUser m.User, outputZipPath string) {

	var imagesFound = 0
	var tablesFound = 0

	var WalkImageCallback = func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		if !strings.Contains(fi.Name(), ".txt") {
			imagesFound++
			log.Info(strconv.Itoa(imagesFound))
			file, err := os.Open(path)
			if err != nil {
				// log.Error("Failed to read image file from zip: " + err.Error())
				return nil
			}
			ProcessImage(currentUser, file, nil, true)
		}
		return nil
	}

	var WalkITableCallback = func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		if strings.Contains(fi.Name(), ".txt") {
			tablesFound++
			// log.Info(strconv.Itoa(tablesFound))
			file, err := os.Open(path)
			if err != nil {
				log.Error("Failed to read txt file from zip: " + err.Error())
				return nil
			}
			defer file.Close()
			ParseImageTable(currentUser, file)
		}
		return nil
	}

	log.Info("Received zip file for processing")

	filepath.Walk(outputZipPath, WalkImageCallback)

	log.Info("Images found in zip: " + strconv.Itoa(imagesFound))

	log.Info("Starting processing of image tables.")

	filepath.Walk(outputZipPath, WalkITableCallback)

	log.Info("Tables found in zip: " + strconv.Itoa(tablesFound))

	os.Remove(outputZipPath)
}

func ProcessImage(currentUser m.User, file io.ReadCloser, imageWaitGroup *sync.WaitGroup, closeReader bool) error {
	if closeReader {
		defer file.Close()
	}

	outputPath := "../uploads/" + currentUser.Username + "/" + base64.StdEncoding.EncodeToString([]byte(currentUser.Username+time.Now().String()))

	if err := os.MkdirAll("../uploads/"+currentUser.Username+"/", 0777); err != nil {
		log.Error("Error in creating output dir " + err.Error())
		return errors.New("Failed to upload the image.")
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Error("Error in creating output file, process image " + err.Error())
		return errors.New("Failed to upload the image.")
	}
	defer outputFile.Close()

	if _, err = io.Copy(outputFile, file); err != nil {
		log.Error("Error in copying file " + err.Error())
		os.Remove(outputPath)
		return errors.New("Failed to upload the image.")
	}

	var lat, long string
	var date time.Time

	openedFile, err := os.Open(outputPath)
	if err != nil {
		log.Error("Error opening file in image upload " + err.Error())
	}
	defer openedFile.Close()

	if exifParser, _ := exif.Decode(openedFile); exifParser == nil {
		lat = ""
		long = ""
		date = time.Now()
	} else {
		lat = ""
		long = ""
		date, err = exifParser.DateTime()
		if err != nil {
			log.Error("Error in exif parser date: " + err.Error())
			date = time.Now()
		}
	}

	image := m.Image{bson.NewObjectId(), currentUser.Id, outputPath, lat, long, date, time.Now(), 1, "", true, bson.NewObjectId(),
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", false}
	if err = m.GetDB("Image").Insert(&image); err != nil {
		log.Error("Error in saving image file to db " + err.Error())
		return errors.New("Failed to upload the image.")
	}
	LinkImageToGps(image, imageWaitGroup)
	return nil
}

func GetMinMaxDateFromCoordinates(coordinates []m.Coordinate) (time.Time, time.Time) {
	if len(coordinates) > 0 {
		first := 0
		last := len(coordinates) - 1
		return coordinates[first].Date, coordinates[last].Date
	} else {
		return time.Now(), time.Now()
	}
}

func TimeInSec(time string) int {
	splitTime := strings.Split(time, ":")
	var splitTimeInt [3]int
	for i := 0; i < len(splitTime); i++ {
		j, err := strconv.Atoi(splitTime[i])

		if err == nil {
			splitTimeInt[i] = j
		} else {
			splitTimeInt[i] = 0
		}
	}

	return splitTimeInt[0]*3600 + splitTimeInt[1]*60 + splitTimeInt[2]
}

func ParseDateFromContext(prefix string, ctx context.Context) (time.Time, bool) {
	now := time.Now()

	if strings.Contains(prefix, "from") {
		now = now.AddDate(0, 0, -1)
	}
	if strings.Contains(prefix, "to") {
		now = now.AddDate(0, 0, 1)
	}

	anythingSelected := false

	year, err := strconv.Atoi(ctx.FormValue(prefix + "year"))
	if err != nil {
		year = now.Year()
	} else {
		anythingSelected = true
	}
	var month time.Month
	month_int, err := strconv.Atoi(ctx.FormValue(prefix + "month"))
	if err != nil {
		month = now.Month()
	} else {
		month = time.Month(month_int)
		anythingSelected = true
	}
	day, err := strconv.Atoi(ctx.FormValue(prefix + "day"))
	if err != nil {
		day = now.Day()
	} else {
		anythingSelected = true
	}
	hour, err := strconv.Atoi(ctx.FormValue(prefix + "hour"))
	if err != nil {
		hour = now.Hour()
	} else {
		anythingSelected = true
	}
	min, err := strconv.Atoi(ctx.FormValue(prefix + "min"))
	if err != nil {
		min = now.Minute()
	} else {
		anythingSelected = true
	}
	return time.Date(year, month, day, hour, min, 0, 0, time.UTC), anythingSelected
}

func ParseDateFromStrings(dateString string, timeString string) time.Time {
	now := time.Now()

	dateValues := strings.Split(dateString, "/")
	timeValues := strings.Split(timeString, ":")

	year, err := strconv.Atoi(dateValues[0])
	if err != nil {
		year = now.Year()
	}
	var month time.Month
	month_int, err := strconv.Atoi(dateValues[1])
	if err != nil {
		month = now.Month()
	} else {
		month = time.Month(month_int)
	}
	day, err := strconv.Atoi(dateValues[2])
	if err != nil {
		day = now.Day()
	}
	hour, err := strconv.Atoi(timeValues[0])
	if err != nil {
		hour = now.Hour()
	}
	min, err := strconv.Atoi(timeValues[1])
	if err != nil {
		min = now.Minute()
	}
	var sec int
	if len(timeValues) > 2 {
		sec, err = strconv.Atoi(timeValues[2])
		if err != nil {
			min = now.Second()
		}
	} else {
		sec = 0
	}

	return time.Date(year, month, day, hour, min, sec, 0, time.UTC)
}

func CheckValidPassword(password string) error {
	if len(password) < 8 {
		return errors.New("The password has to be at least 8 characters long.")
	}

	if password == strings.ToLower(password) {
		return errors.New("The password has to have at least one uppercase letter.")
	}

	if password == strings.ToUpper(password) {
		return errors.New("The password has to have at least one lowercase letter.")
	}

	var hasSymbol = regexp.MustCompile(`[^0-9a-zA-Z]`)
	if !hasSymbol.MatchString(password) {
		return errors.New("The password has to have at least one symbol.")
	}

	return nil
}

func LoadCypherKey() error {
	// Load the encryption key
	var err error
	key, err = ioutil.ReadFile("../key.pub")
	if err != nil {
		log.Error("Error reading key: " + err.Error())
		return err
	}
	return nil
}

func Decrypt(file []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(file) < aes.BlockSize {
		return nil, errors.New("Failed to decrypt the file")
	}
	iv := file[:aes.BlockSize]
	file = file[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(file, file)
	return file, nil
}

func GetTrackID(dateArray1 []string, dateArray2 []string) int64 {
	if dateArray1 == nil || dateArray2 == nil {
		trackIDcounter++
		return trackIDcounter
	}

	if !strings.EqualFold(dateArray1[0], dateArray2[0]) {
		trackIDcounter++
		return trackIDcounter
	}

	seconds1 := TimeInSec(dateArray1[1])
	seconds2 := TimeInSec(dateArray2[1])
	if seconds2-seconds1 > 120 {
		trackIDcounter++
		return trackIDcounter
	}
	return trackIDcounter
}

func Append(slice [][]string, elements ...[]string) [][]string {
	n := len(slice)
	total := len(slice) + len(elements)
	if total > cap(slice) {
		// Reallocate. Grow to 1.5 times the new size, so we can still grow.
		newSize := total*3/2 + 1
		newSlice := make([][]string, total, newSize)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[:total]
	copy(slice[n:], elements)
	return slice
}

func GetStyle(d []string, l int) int {
	result := 0
	for k := 0; k < len(d); k++ {
		num, _ := strconv.ParseUint(d[k], 0, 64)
		result += int(num)
	}
	return int(result % l)
}
