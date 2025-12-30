package utils

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"unicode"

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

func ExtractNumber(s string) int {
	n := ""
	for _, c := range s {
		if unicode.IsDigit(c) {
			n += string(c)
		} else {
			break
		}
	}

	if len(n) == 0 {
		return 0
	}

	i, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return 0
	}

	return int(i)
}

var validImageExts = []string{
	".png",
	".jpg",
	".jpeg",
}

func IsValidImageExt(ext string) bool {
	for _, e := range validImageExts {
		if ext == e {
			return true
		}
	}

	return false
}

// TODO(patrik): Update this
var validExts []string = []string{
	".wav",
	".flac",
	".opus",
}

func IsValidTrackExt(ext string) bool {
	for _, valid := range validExts {
		if valid == ext {
			return true
		}
	}

	return false
}
