package main

import (
	"bufio"
	"bytes"
	"encoding/base32"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/exec"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"
)

const mySecret = "CE5B8B5FA2F60434613F42EE10359F08"
const myIssuer = "Nuveus"
const myUser = "fatihsoydan"

const UPPER_HALF_BLOCK = "▀"

func clearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// 48;2;r;g;bm - set background colour to rgb
func rgbBackgroundSequence(r, g, b uint8) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

// 38;2;r;g;bm - set text colour to rgb
func rgbTextSequence(r, g, b uint8) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func resetColorSequence() string {
	return "\x1b[0m"
}

func convertColorToRGB(col color.Color) (uint8, uint8, uint8) {
	rgbaColor := color.RGBAModel.Convert(col)
	_r, _g, _b, _ := rgbaColor.RGBA()
	// rgb values are uint8s, I cannot comprehend why the stdlib would return
	// int32s :facepalm:
	r := uint8(_r & 0xFF)
	g := uint8(_g & 0xFF)
	b := uint8(_b & 0xFF)
	return r, g, b
}

func convertImageToANSI(img image.Image, skip int) string {
	// We'll just reuse this to increment the loop counters
	skip += 1
	ansi := resetColorSequence()
	yMax := img.Bounds().Max.Y
	xMax := img.Bounds().Max.X

	sequences := make([]string, yMax)

	for y := img.Bounds().Min.Y; y < yMax; y += 2 * skip {
		sequence := ""
		for x := img.Bounds().Min.X; x < xMax; x += skip {
			upperPix := img.At(x, y)
			lowerPix := img.At(x, y+skip)

			ur, ug, ub := convertColorToRGB(upperPix)
			lr, lg, lb := convertColorToRGB(lowerPix)

			if y+skip >= yMax {
				sequence += resetColorSequence()
			} else {
				sequence += rgbBackgroundSequence(lr, lg, lb)
			}

			sequence += rgbTextSequence(ur, ug, ub)
			sequence += UPPER_HALF_BLOCK

			sequences[y] = sequence
		}
	}

	for y := img.Bounds().Min.Y; y < yMax; y += 2 * skip {
		ansi += sequences[y] + resetColorSequence() + "\n"
	}

	return ansi
}

func Encode(secret []byte) string {
	var encoder = base32.StdEncoding.WithPadding(base32.NoPadding)
	result := encoder.EncodeToString(secret)
	return result
}

func drawQrCode() {
	secret := Encode([]byte(mySecret))
	issuerUrl := gotp.NewDefaultTOTP(secret).ProvisioningUri(myUser, myIssuer)
	var png []byte
	png, err := qrcode.Encode(issuerUrl, qrcode.Medium, 256)
	if err != nil {
		log.Fatal(err)
	}
	img, _, _ := image.Decode(bytes.NewReader(png))
	fmt.Print(convertImageToANSI(img, 1))
}

func getCode() {
	totp := gotp.NewDefaultTOTP(Encode([]byte(mySecret)))
	log.Println(totp.Now())
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Welcome to main menu :) [Please choose an option then hit Enter]")
	fmt.Println("1: Generate QRCode for Registration")
	fmt.Println("2: Check Verification Code")
	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1)
	switch text {
	case "1":
		drawQrCode()
	case "2":
		getCode()
	default:
		fmt.Println("Please enter a number of listed options")
	}
}
