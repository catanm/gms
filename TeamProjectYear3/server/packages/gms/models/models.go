package m

import (
	"gms/log"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

var _db *mgo.Session
var err error
var db_map = make(map[string]*mgo.Collection)

var db_name = "gms"

type User struct {
	Id              bson.ObjectId `json:"id" bson:"_id"`
	Username        string        `json:"username" bson:"username"`
	Email           string        `json:"email" bson:"email"`
	UploadAsPrivate bool          `json:"uploadAsPrivate" bson:"uploadAsPrivate"`
	Token           string        `json:"token" bson:"token"`
	TokenUpdate     time.Time     `json:"tokenupdate" bson:"tokenupdate"`
}

type Passport struct {
	Id       bson.ObjectId `json:"id" bson:"_id"`
	User     bson.ObjectId `json:"user" bson:"user"`
	Password string        `json:"password" bson:"password"`
	// enums in GO are different, cannot (or shouldn't) be used
	// Protocol string `json:"protocol" bson:"protocol"` No need of it IMO
	Provider   string `json:"provider" bson:"provider"`
	Identifier string `json:"identifier" bson:"identifier"`
	// Hardcoding for oauth2
	AccessToken  string `json:"accessToken" bson:"accessToken"`
	RefreshToken string `json:"refreshToken" bson:"refreshToken"`
}

type Video struct {
	User bson.ObjectId `json:"user" bson:"user"`
	Url  string        `json:"url" bson:"url"`
	Date time.Time     `json:"date" bson:"date"`
}

type Image struct {
	Id        bson.ObjectId `json:"id" bson:"_id"`
	User      bson.ObjectId `json:"user" bson:"user"`
	Url       string        `json:"url" bson:"url"`
	Lat       string        `json:"lat" data: "lat"`
	Lon       string        `json:"lon" data: "lon"`
	Date      time.Time     `json:"date" bson:"date"`
	Uploaded  time.Time     `json:"uploaded" bson:"uploaded"`
	Blur      float64       `json:"blur" bson:"blur"`
	Phash     string        `json:"phash" bson:"phash"`
	Show      bool          `json:"show" bson:"show"`
	Cluster   bson.ObjectId `json:"cluster" bson:"cluster"`
	CamId     string        `json:"camid" bson: camid`
	Size      string        `json:"size" bson: size`
	Type      string        `json:"type" bson: type`
	PropInfo  string        `json:"propinfo" bson: propinfo`
	Accx      string        `json:"accx" bson: accx`
	Accy      string        `json:"accy" bson: accy`
	Accz      string        `json:"accz" bson: accz`
	Magx      string        `json:"magx" bson: magx`
	Magy      string        `json:"magy" bson: magy`
	Magz      string        `json:"magz" bson: magz`
	Red       string        `json:"red" bson: red`
	Green     string        `json:"green" bson: green`
	Blue      string        `json:"blue" bson: blue`
	Lum       string        `json:"lum" bson: lum`
	Tem       string        `json:"tem" bson: tem`
	GpsStatus string        `json:"gpsstatus" bson: gpsstatus`
	Alt       string        `json:"alt" bson: alt`
	GS        string        `json:"gs" bson: gs`
	Herr      string        `json:"herr" bson: herr`
	Verr      string        `json:"verr" bson: verr`
	Exp       string        `json:"exp" bson: exp`
	Gain      string        `json:"gain" bson: gain`
	RBal      string        `json:"rbal" bson: rbal`
	GBal      string        `json:"gbal" bson: gbal`
	BBal      string        `json:"bbal" bson: bbal`
	Xor       string        `json:"xor" bson: xor`
	Yor       string        `json:"yor" bson: yor`
	Zor       string        `json:"zor" bson: zor`
	Stags     string        `json:"stags" bson: stags`
	Tags      string        `json:"tags" bson: tags`
	Processed bool          `json:"processed" bson: processed`
}

type Cluster struct {
	Id          bson.ObjectId `json:"id" bson:"_id"`
	Phash       string        `json:"phash" bson:"phash"`
	User        bson.ObjectId `json:"user" bson:"user"`
	Date        time.Time     `json:"date" bson:"date"`
	HasSelected bool          `json:"hasSelected" bson:"hasSelected"`
}

type Track struct {
	Id          bson.ObjectId `json:"id" bson:"_id"`
	User        bson.ObjectId `json:"user" bson:"user"`
	TrackID     string        `json:"trackId" bson:"trackId"`
	Date        time.Time     `json:"date" bson:"date"`
	Coordinates []Coordinate  `json:"coordinates" bson:"coordinates"`
	MinDate     time.Time     `json:"minDate" bson:"minDate"`
	MaxDate     time.Time     `json:"maxDate" bson:"maxDate"`
	Uploaded    time.Time     `json:"uploaded" bson:"uploaded"`
}

type Coordinate struct {
	Lat  string    `json:"lat" bson:"lat"`
	Lon  string    `json:"lon" bson:"lon"`
	Date time.Time `json:"date" bson:"date"`
}

type PopularLocations struct {
	Id              bson.ObjectId    `json:"id" bson:"_id"`
	User            bson.ObjectId    `json:"user" bson:"user"`
	PopularMarkers  []Marker         `json:"marker" bson:"marker"`
	Recommendations []Recommendation `json:"recommend" bson:"recommend"`
}

type Marker struct {
	Lon        string `json:"lon" bson:"lon"`
	Street     string `json:"street" bson:"street"`
	Lat        string `json:"lat" bson:"lat"`
	Popularity string `json:"popularity" bson:"popularity"`
}

type Recommendation struct {
	Lon           string `json:"lon" bson:"lon"`
	PlaceCategory string `json:"placeCategory" bson:"placeCategory"`
	PlaceName     string `json:"placeName" bson:"placeName"`
	Lat           string `json:"lat" bson:"lat"`
	Popularity    string `json:"popularity" bson:"popularity"`
}

type TrendingAllUsersEntry struct {
	Id         bson.ObjectId `json:"id" bson:"_id"`
	Lat        string        `json:"lat" bson:"lat"`
	Lon        string        `json:"lon" bson:"lon"`
	ImageUrl   string        `json:"url" bson:"url"`
	Popularity string        `json:"popularity" bson:"popularity"`
	StreetName string        `json:"streetName" bson:"streetName"`
}

type AllUsersTrendingAndRecom struct {
	Trend []TrendingAllUsersEntry `json:"marker" bson:"marker"`
	Recom []Recommendation        `json:"recommend" bson:"recommend"`
}

func GetDB(collection string) *mgo.Collection {
	return db_map[collection]
}

func Connect() {
	// Database
	// mongolab.com
	var dbServer = "mongolab"
	_db, err = mgo.Dial("mongodb://127.0.0.1:27017/gms")
	if err != nil {
		log.Error("Error on database connection: " + err.Error())
		_db, err = mgo.Dial("mongodb://127.0.0.1:27017/gms")
		dbServer = "local db"
		if err != nil {
			panic(err)
		}
	}

	// This makes the db monotonic
	_db.SetMode(mgo.Monotonic, true)

	// Database name (test) and the collection names ("User") are set
	collections := _db.DB(db_name)
	db_map["User"] = collections.C("User")
	db_map["Passport"] = collections.C("Passport")
	db_map["Comment"] = collections.C("Comment")
	db_map["Track"] = collections.C("Track")
	db_map["Image"] = collections.C("Image")
	db_map["Video"] = collections.C("Video")
	db_map["Cluster"] = collections.C("Cluster")
	db_map["RecommendUser"] = collections.C("RecommendUser")
	db_map["TrendingAllUsers"] = collections.C("TrendingAllUsers")
	db_map["RecommendAllUsers"] = collections.C("RecommendAllUsers")

	log.Info("Database connection established with " + dbServer)
}

func Close() {
	_db.Close()
}
