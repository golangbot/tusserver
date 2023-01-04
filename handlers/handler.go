package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/JitenPalaparthi/cresFileServer/database"
	"github.com/JitenPalaparthi/cresFileServer/models"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

type FileHandler struct {
	FileDB  *database.FileDB
	DirPath string
}

func (fh *FileHandler) FileDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fID := vars["fileID"]
	file, err := fh.FileDB.File(fID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	glog.Infoln("going to write upload offset to output")
	w.Header().Set("Upload-Offset", strconv.Itoa(*file.Offset))
	w.WriteHeader(http.StatusOK)
	return
}

func (fh *FileHandler) FilePatchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("going to patch file")
	vars := mux.Vars(r)
	fID := vars["fileID"]
	file, err := fh.FileDB.File(fID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if *file.UploadComplete == true {
		e := "Upload already completed"
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(e))
		return
	}
	off, err := strconv.Atoi(r.Header.Get("Upload-Offset"))
	if err != nil {
		glog.Infoln("Improper upload offset", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	glog.Infof("Upload offset %d\n", off)
	if *file.Offset != off {
		e := fmt.Sprintf("Expected Offset %d got offset %d", *file.Offset, off)
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(e))
		return
	}

	glog.Infoln("Content length is", r.Header.Get("Content-Length"))
	clh := r.Header.Get("Content-Length")
	cl, err := strconv.Atoi(clh)
	if err != nil {
		glog.Infoln("unknown content length")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if cl != (file.UploadLength - *file.Offset) {
		e := fmt.Sprintf("Content length doesn't not match upload length.Expected content length %d got %d", file.UploadLength-*file.Offset, cl)
		glog.Infoln(e)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(e))
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Infof("Received file partially %s\n", err)
		glog.Infoln("Size of received file ", len(body))
	}
	fp := fmt.Sprintf("%s/%s", fh.DirPath, fID)
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		glog.Infof("unable to open file %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	n, err := f.WriteAt(body, int64(off))
	if err != nil {
		glog.Infof("unable to write %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	glog.Infoln("number of bytes written ", n)
	no := *file.Offset + n
	file.Offset = &no

	uo := strconv.Itoa(*file.Offset)
	w.Header().Set("Upload-Offset", uo)
	if *file.Offset == file.UploadLength {
		glog.Infoln("upload completed successfully")
		*file.UploadComplete = true
	}

	err = fh.FileDB.UpdateFile(file)
	if err != nil {
		glog.Infoln("Error while updating file", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

	return

}

func (fh *FileHandler) CreateFileHandler(w http.ResponseWriter, r *http.Request) {
	ul, err := strconv.Atoi(r.Header.Get("Upload-Length"))
	if err != nil {
		e := "Improper upload length"
		glog.Infof("%s %s", e, err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(e))
		return
	}
	glog.Infof("upload length %d\n", ul)
	io := 0
	uc := false
	f := models.File{
		Offset:         &io,
		UploadLength:   ul,
		UploadComplete: &uc,
	}
	fileID, err := fh.FileDB.CreateFile(f)
	if err != nil {
		e := "Error creating file in DB"
		glog.Infof("%s %s\n", e, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	filePath := path.Join(fh.DirPath, fileID)
	file, err := os.Create(filePath)
	if err != nil {
		e := "Error creating file in filesystem"
		glog.Infof("%s %s\n", e, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	w.Header().Set("Location", fmt.Sprintf("localhost:8080/files/%s", fileID))
	w.WriteHeader(http.StatusCreated)
	return
}
