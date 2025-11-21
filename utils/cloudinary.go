package utils

import (
	"context"
	"mime/multipart"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(file multipart.File, fileHeader *multipart.FileHeader, cloudName, apiKey, apiSecret string, folder string, publicID string) (string, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return "", err
	}
	params := uploader.UploadParams{}
	if folder != "" {
		params.Folder = folder
	}
	if publicID != "" {
		params.PublicID = publicID
	} else {
		params.PublicID = fileHeader.Filename
	}
	resp, err := cld.Upload.Upload(context.Background(), file, params)
	if err != nil {
		return "", err
	}
	return resp.SecureURL, nil
}
