package main

import (
	"fmt"
	"net/http"
	"io"
	"encoding/base64"


	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	mediaType := header.Header.Get("Content-Type") 
	if mediaType == "" {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	fmt.Println("Uploaded media type:", mediaType)

	parts:= strings.Split(mediaType, "/")
	if len(parts) !=  2 {
		respondWithError(w, http.StatusBadRequest, "invalid Content-Type", err)
		return
	}
	ext := parts[1]
	filename := videoID + "." + ext
	path := filepath.Join(cfg.assetsRoot, filename)
	

	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to get info", err)
		return
	}

	//converting the image data to base64 string
	encoded := base64.StdEncoding.EncodeToString(data)

	//build url with encoded image
	encodedURL := fmt.Sprintf("data:%s;base64,%s", mediaType, encoded)

	//Get video from db
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "video not found", err)
		return
	}

	//Check the ownership of the video
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You arent the creater of the video", err)
		return
	}

	video.ThumbnailURL = &encodedURL

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Can't update the video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
