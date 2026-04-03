package storage

import "errors"

var (
    ErrInvalidFileType    = errors.New("invalid file type")
    ErrFileTooLarge       = errors.New("file too large")
    ErrUploadFailed       = errors.New("upload failed")
    ErrFileNotFound       = errors.New("file not found")
    ErrStorageUnavailable = errors.New("storage unavailable")
)