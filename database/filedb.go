package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/JitenPalaparthi/cresFileServer/models"
	"github.com/golang/glog"
)

type FileDB struct {
	DB *sql.DB
}

func (fh *FileDB) CreateTable() error {
	q := `CREATE TABLE IF NOT EXISTS file(file_id SERIAL PRIMARY KEY, 
 		  file_offset INT NOT NULL, file_upload_length INT NOT NULL, file_upload_complete BOOLEAN NOT NULL, 
          created_at TIMESTAMP default NOW() NOT NULL, modified_at TIMESTAMP default NOW() NOT NULL)`
	_, err := fh.DB.Exec(q)
	if err != nil {
		return err
	}
	return nil
}

func (fh *FileDB) CreateFile(f models.File) (string, error) {
	cfstmt := `INSERT INTO file(file_offset, file_upload_length, file_upload_complete) VALUES($1, $2, $3) RETURNING file_id`
	fileID := 0
	err := fh.DB.QueryRow(cfstmt, f.Offset, f.UploadLength, f.UploadComplete).Scan(&fileID)
	if err != nil {
		return "", err
	}
	fid := strconv.Itoa(fileID)
	return fid, nil
}

func (fh *FileDB) UpdateFile(f models.File) error {
	var query []string
	var param []interface{}
	if f.Offset != nil {
		of := fmt.Sprintf("file_offset = $1")
		ofp := f.Offset
		query = append(query, of)
		param = append(param, ofp)
	}
	if f.UploadComplete != nil {
		uc := fmt.Sprintf("file_upload_complete = $2")
		ucp := f.UploadComplete
		query = append(query, uc)
		param = append(param, ucp)
	}

	if len(query) > 0 {
		mo := "modified_at = $3"
		mop := "NOW()"

		query = append(query, mo)
		param = append(param, mop)

		qj := strings.Join(query, ",")

		sqlq := fmt.Sprintf("UPDATE file SET %s WHERE file_id = $4", qj)

		param = append(param, f.FileID)

		glog.Infoln("generated update query", sqlq)
		_, err := fh.DB.Exec(sqlq, param...)
		if err != nil {
			glog.Infoln("Error during file update", err)
			return err
		}
	}
	return nil
}

func (fh *FileDB) File(fileID string) (models.File, error) {
	fID, err := strconv.Atoi(fileID)
	if err != nil {
		glog.Infoln("Unable to convert fileID to string", err)
		return models.File{}, err
	}
	glog.Infoln("going to query for fileID", fID)
	gfstmt := `select file_id, file_offset, file_upload_length, file_upload_complete from file where file_id = $1`
	row := fh.DB.QueryRow(gfstmt, fID)
	f := models.File{}
	err = row.Scan(&f.FileID, &f.Offset, &f.UploadLength, &f.UploadComplete)
	if err != nil {
		glog.Infoln("error while fetching file", err)
		return models.File{}, err
	}
	return f, nil
}
