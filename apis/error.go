package apis

import (
	"net/http"

	"github.com/nanoteck137/pyrin"
)

const (
	ErrTypeInvalidAuth      pyrin.ErrorType = "INVALID_AUTH"
	ErrTypeArtistNotFound   pyrin.ErrorType = "ARTIST_NOT_FOUND"
	ErrTypeAlbumNotFound    pyrin.ErrorType = "ALBUM_NOT_FOUND"
	ErrTypeTrackNotFound    pyrin.ErrorType = "TRACK_NOT_FOUND"
	ErrTypeTaglistNotFound  pyrin.ErrorType = "TAGLIST_NOT_FOUND"
	ErrTypeApiTokenNotFound pyrin.ErrorType = "API_TOKEN_NOT_FOUND"
	ErrTypeQueueNotFound    pyrin.ErrorType = "QUEUE_NOT_FOUND"

	ErrTypeInvalidFilter      pyrin.ErrorType = "INVALID_FILTER"
	ErrTypeInvalidSort        pyrin.ErrorType = "INVALID_SORT"
	ErrTypeUserAlreadyExists  pyrin.ErrorType = "USER_ALREADY_EXISTS"
	ErrTypeUserNotFound       pyrin.ErrorType = "USER_NOT_FOUND"
	ErrTypeInvalidCredentials pyrin.ErrorType = "INVALID_CREDENTIALS"

	ErrTypePlaylistNotFound        pyrin.ErrorType = "PLAYLIST_NOT_FOUND"
	ErrTypePlaylistAlreadyHasTrack pyrin.ErrorType = "PLAYLIST_ALREADY_HAS_TRACK"
)

func InvalidAuth(message string) *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypeInvalidAuth,
		Message: "Invalid auth: " + message,
	}
}

func ArtistNotFound() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusNotFound,
		Type:    ErrTypeArtistNotFound,
		Message: "Artist not found",
	}
}

func AlbumNotFound() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusNotFound,
		Type:    ErrTypeAlbumNotFound,
		Message: "Album not found",
	}
}

func TrackNotFound() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusNotFound,
		Type:    ErrTypeTrackNotFound,
		Message: "Track not found",
	}
}

func TaglistNotFound() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusNotFound,
		Type:    ErrTypeTaglistNotFound,
		Message: "Taglist not found",
	}
}

func ApiTokenNotFound() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusNotFound,
		Type:    ErrTypeApiTokenNotFound,
		Message: "Api Token not found",
	}
}

func QueueNotFound() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusNotFound,
		Type:    ErrTypeQueueNotFound,
		Message: "Queue not found",
	}
}

func InvalidFilter(err error) *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypeInvalidFilter,
		Message: err.Error(),
	}
}

func InvalidSort(err error) *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypeInvalidSort,
		Message: err.Error(),
	}
}

func UserAlreadyExists() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypeUserAlreadyExists,
		Message: "User already exists",
	}
}

func UserNotFound() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusUnauthorized,
		Type:    ErrTypeUserNotFound,
		Message: "User not found",
	}
}

func InvalidCredentials() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusUnauthorized,
		Type:    ErrTypeInvalidCredentials,
		Message: "Invalid Credentials",
	}
}

func PlaylistNotFound() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusNotFound,
		Type:    ErrTypePlaylistNotFound,
		Message: "Playlist not found",
	}
}

func PlaylistAlreadyHasTrack() *pyrin.Error {
	return &pyrin.Error{
		Code:    http.StatusBadRequest,
		Type:    ErrTypePlaylistAlreadyHasTrack,
		Message: "Playlist already has track",
	}
}
