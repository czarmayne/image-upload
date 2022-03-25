package main

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type ImageMetadata struct {
	Id          primitive.ObjectID `bson:"_id,omitempty"`
	Filename    string             `bson:"Filename,omitempty"`
	Size        int64              `bson:"Size,omitempty"`
	ContentType string             `bson:"ContentType,omitempty"`
	HTTPHistory HTTPHistory
}

type HTTPHistory struct {
	Id         primitive.ObjectID `bson:"_id,omitempty"`
	Origin     string             `bson:"Origin,omitempty"`
	Path       string             `bson:"Path,omitempty"`
	Method     string             `bson:"Method,omitempty"`
	UserAgent  string             `bson:"UserAgent,omitempty"`
	RemoteAddr string             `bson:"RemoteAddr,omitempty"`
	DateTime   string             `bson:"DateTime,omitempty"`
}

const MAX_UPLOAD_SIZE = 8388608 // 8MB
var Token = os.Getenv("TOKEN")

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the debug severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	//TODO: add healthcheck / heartbeat for service and database connectivity
	log.Info("Image Upload Service is starting...")
	handleRequests()
	//TODO: add graceful shutdown
}

func close(client *mongo.Client, ctx context.Context,
	cancel context.CancelFunc) {
	defer cancel()
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func handleRequests() {
	log.Info("Preparing routing handlers")
	router := mux.NewRouter()
	router.HandleFunc("/", homePage)
	router.HandleFunc("/token", getToken).Methods("GET")
	router.HandleFunc("/upload", UploadImage).Methods("POST")
	log.Info("Service up!")
	log.Fatal(http.ListenAndServe(":8081", router))
}

// Serves as the landing page once the server is up
func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "index.html")
}

// getToken provides the auth from the environment variable
func getToken(w http.ResponseWriter, r *http.Request) {
	log.Info("Get token from environment variable")
	log.Debugf("TOKEN: %s\n", Token)
	EncodeOkResponse(Token, w)
}

// UploadImage executes in the following manner:
// 1. Validate the input file: size, auth, content-type
// 2. Store the data in temporary file
// 3. Store the image metadata in database
// 4. Store the http details in database
func UploadImage(w http.ResponseWriter, r *http.Request) {
	log.Info("[START] upload image execution")

	log.Infof("[START] validating the input form...")
	if !IsFileSizeValid(w, r) {
		http.Error(w, "File exceeded the size limit of 8 megabytes", http.StatusBadRequest)
		return
	}

	if !IsTokenValid(GetAuthToken(r)) {
		http.Error(w, "Access not allowed", http.StatusForbidden)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Info(err)
		http.Error(w, "Error Retrieving the File", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !IsValidType(getContentType(fileHeader)) {
		http.Error(w, "Invalid File Type", http.StatusUnsupportedMediaType)
		return
	}
	log.Info("[END] validating the input form...")
	
	createTempFile(file)

	log.Info("[START] saving the record...")
	record, err := saveRecord(buildImageMetadata(fileHeader, r))

	if err != nil {
		http.Error(w, "Error occurred while uploading the file", http.StatusInternalServerError)
		return

	}
	log.Infof("[END] saving the record...", record)

	log.Info("[END] upload image execution")
	EncodeOkResponse(record, w)
}

func createTempFile(file multipart.File) {
	log.Info("[START] create temporary file...")
	tempFile, err := ioutil.TempFile("tmp", "upload-*.png")
	if err != nil {
		log.Error(err)
	}
	defer tempFile.Close()
	
	// Copy content to the created temp file
	_, copyError := io.Copy(tempFile, file)
	if copyError != nil {
		log.Println("Error copying data", copyError)
	}
	log.Info("[END] create temporary file...")
}

func buildImageMetadata(fileHeader *multipart.FileHeader, r *http.Request) ImageMetadata {
	log.Debugf("[START] Build Image Metadata Request")
	var img ImageMetadata
	img.Filename = fileHeader.Filename
	img.Size = fileHeader.Size
	img.ContentType = strings.Join(fileHeader.Header["Content-Type"], "")
	img.Filename = fileHeader.Filename
	img.HTTPHistory = buildHttpHist(r)
	log.Debugf("[End] Build Image Metadata Request: ", img)
	return img
}

func buildHttpHist(r *http.Request) HTTPHistory {
	log.Debugf("[START] Build HttpHist Request")
	var httpHist HTTPHistory
	httpHist.Origin = strings.Join(r.Header["Origin"], "")
	httpHist.Path = r.URL.Path
	httpHist.Method = r.Method
	httpHist.UserAgent = strings.Join(r.Header["User-Agent"], "")
	httpHist.RemoteAddr = r.RemoteAddr
	log.Debugf("[END] Build HttpHist Request: ", httpHist)
	return httpHist
}

// FIXME: Not the best practice as it was not initialized during the startup
// Save the image metadata and http history
func saveRecord(img ImageMetadata) (*mongo.InsertOneResult, error) {
	// ctx will be used to set deadline for process, deadline will be 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// connect to database with the give clientOptions
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017")) //TODO: move URI to environment variable

	if err != nil {
		panic(err)
	}
	defer close(client, ctx, cancel)

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	// get collection
	collection := client.Database("brankas").Collection("image")

	log.Infof("Database connected successfully ", collection.Name())

	result, err := collection.InsertOne(ctx, img)
	if err != nil {
		log.Error("Error Occured while trying to persist record in the data store")
		log.Error(err.Error())
		return nil, err
	}
	log.Info("Record has been saved: ", result.InsertedID)
	return result, err
}

// getContentType provides the Content-Type from the Header
func getContentType(fileHeader *multipart.FileHeader) string {
	contentType := strings.Join(fileHeader.Header["Content-Type"], "")
	log.WithFields(log.Fields{
		"contentType": contentType,
		"filename":    fileHeader.Filename,
		"size":        fileHeader.Size,
		"header":      fileHeader.Header,
	}).Debug("File Content Header")
	return contentType
}

//FIXME: methods below can be moved to a util folder

// IsFileSizeValid deals with verification of the file size
// which should not greater than 8 megabytes.
// It parses the form and if it exceeds the max size, error will occur then it returns false
func IsFileSizeValid(w http.ResponseWriter, r *http.Request) bool {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		return false
	}
	return true
}

// IsTokenValid deals with auth validation
// input token value should be the same as the environment Token value
func IsTokenValid(token string) bool {
	log.Debugf("Validate input token: %s == Token: %s", token, Token)
	return token == Token
}

// IsValidType checks the content type of the uploaded file by providing
// t = the contentType from the FileHeader
func IsValidType(t string) bool {
	allowedContentTypes := []string{"image/png", "image/jpeg", "image/gif"} // TODO: move to env variable

	if mimetype.EqualsAny(t, allowedContentTypes...) {
		log.Debugf("%s is allowed\n", t)
		return true
	}
	log.Debugf("%s is not allowed\n", t)
	return false
}

// GetAuthToken from the request
func GetAuthToken(r *http.Request) string {
	log.Debug("Get input token")
	auth := r.FormValue("auth")
	log.Debug("Auth value: ", auth)
	return auth
}

// EnableCors allow access by any origin
// FIXME: to allow specific origin only
func EnableCors(w *http.ResponseWriter) {
	log.Debug("Allow CORS")
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

// EncodeOkResponse marshall response json and write the server response
func EncodeOkResponse(responseStruct interface{}, w http.ResponseWriter) {
	log.Debug("Prepare API Response")
	EnableCors(&w)
	responseJson, err := json.Marshal(responseStruct)
	if err != nil {
		log.Error("Failed to marshall response.")
		http.Error(w, "Failed to encode response.", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}
