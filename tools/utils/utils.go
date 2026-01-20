package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"os/exec"
	"strings"

	"github.com/gosimple/slug"
	"github.com/nrednav/cuid2"
)

var CreateId = createIdGenerator(32)
var CreateSmallId = createIdGenerator(8)

var CreateArtistId = createIdGenerator(10)
var CreateAlbumId = createIdGenerator(16)
var CreateTrackId = createIdGenerator(32)
var CreateTrackMediaId = createIdGenerator(32)

var CreateApiTokenId = createIdGenerator(32)

func createIdGenerator(length int) func() string {
	res, err := cuid2.Init(cuid2.WithLength(length))
	if err != nil {
		log.Fatal("Failed to create id generator", "err", err)
	}

	return res
}

func ParseAuthHeader(authHeader string) string {
	splits := strings.Split(authHeader, " ")
	if len(splits) != 2 {
		return ""
	}

	if splits[0] != "Bearer" {
		return ""
	}

	return splits[1]
}

func CreateResizedImage(src string, dest string, width, height int) error {
	args := []string{
		"convert",
		src,
		"-resize", fmt.Sprintf("%dx%d^", width, height),
		"-gravity", "Center",
		"-extent", fmt.Sprintf("%dx%d", width, height),
		dest,
	}

	cmd := exec.Command("magick", args...)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func ConvertImage(src string, dest string) error {
	args := []string{
		"convert",
		src,
		dest,
	}

	cmd := exec.Command("magick", args...)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func Slug(s string) string {
	return slug.Make(s)
}

func SplitString(s string) []string {
	tags := []string{}
	if s != "" {
		tags = strings.Split(s, ",")
	}

	return tags
}

func TotalPages(perPage, totalItems int) int {
	return int(math.Ceil(float64(totalItems) / float64(perPage)))
}

const (
	letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "0123456789"
)

func randomString(charset string, length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

func GenerateCode() (string, error) {
	part1, err := randomString(letters, 4)
	if err != nil {
		return "", err
	}

	part2, err := randomString(digits, 4)
	if err != nil {
		return "", err
	}

	part3, err := randomString(letters, 4)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s-%s", part1, part2, part3), nil
}

func GenerateAuthChallenge() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
