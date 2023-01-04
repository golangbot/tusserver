package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"

	"github.com/JitenPalaparthi/cresFileServer/database"
	handler "github.com/JitenPalaparthi/cresFileServer/handlers"
	"github.com/golang/glog"
	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

var (
	PORT    string = "8081"
	DSN     string
	dirName = "fileserver"
)

func createFileDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		log.Println("Error while fetching user home directory", err)
		return "", err
	}
	home := u.HomeDir
	dirPath := path.Join(home, dirName)
	err = os.MkdirAll(dirPath, 0744)
	if err != nil {
		log.Println("Error while creating file server directory", err)
		return "", err
	}
	return dirPath, nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: example -stderrthreshold=[INFO|WARNING|ERROR|FATAL] -log_dir=[string]\n")
	flag.PrintDefaults()
	//os.Exit(2)
}

func init() {
	flag.Usage = usage
	//ctx = context.Background()
}

func main() {

	flag.StringVar(&PORT, "port", "50091", "--port=50091")
	flag.StringVar(&DSN, "db", "host=127.0.0.1 user=postgres password=postgres dbname=vehicle_master_db port=1234 sslmode=disable TimeZone=Asia/Shanghai", "--db=host=localhost user=admin password=admin123 dbname=customersdb port=5432 sslmode=disable TimeZone=Asia/Shanghai")

	flag.Set("stderrthreshold", "INFO") // can set up the glog
	flag.Parse()
	defer glog.Flush()
	if os.Getenv("PORT") != "" {
		PORT = os.Getenv("PORT")
	}
	if os.Getenv("DB_CONN") != "" {
		DSN = os.Getenv("DB_CONN")
	}

	//connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbHost, dbPort, dbUser, dbPwd, dbName, sslMode)
	db, err := database.GetConnection(DSN)
	if err != nil {
		glog.Fatalln(err)
	}
	err = db.Ping()
	if err != nil {
		glog.Fatalln(err)
	}
	glog.Infoln("database Connection established successfully")
	glog.Infoln("TUS Server started on port->", PORT)
	fd := &database.FileDB{
		DB: db,
	}
	fh := handler.FileHandler{
		FileDB: fd,
	}
	dir, err := createFileDir()
	if err != nil {
		glog.Fatalln("Error creating file server directory", err)
	}
	fh.DirPath = dir
	glog.Infoln("Directory created successfully")
	err = fd.CreateTable()
	if err != nil {
		glog.Infoln("Error during table creation", err)
	}
	glog.Infoln("table created successfully")
	r := mux.NewRouter()
	r.HandleFunc("/files", fh.CreateFileHandler).Methods("POST")
	r.HandleFunc("/files/{fileID:[0-9]+}", fh.FileDetailsHandler).Methods("HEAD")
	r.HandleFunc("/files/{fileID:[0-9]+}", fh.FilePatchHandler).Methods("PATCH")
	http.ListenAndServe(":"+PORT, r)
}
