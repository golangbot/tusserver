package models

type File struct {
	FileID         int
	Offset         *int
	UploadLength   int
	UploadComplete *bool
}
