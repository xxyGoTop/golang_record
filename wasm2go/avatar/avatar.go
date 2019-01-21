package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"strings"
	"syscall/js"

	"github.com/o1egl/govatar"
	"github.com/satori/go.uuid"

	"github.com/Chyroc/web"
)

func generateAvator(gender govatar.Gender) (string, error) {
	name := uuid.Must(uuid.NewV4()).String()

	img, err := govatar.GenerateFromUsername(gender, name)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return "", err
	}

	fmt.Printf("img buf %s", buf)

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func main() {
	gender := govatar.MALE
	imgDom := web.Document.GetElementById("img")
	genderDom := web.Document.GetElementById("gender").(web.HTMLSelectElement)

	genderDom.AddEventListener("change", func(args []js.Value) {
		if strings.ToLower(genderDom.Value()) == "male" {
			gender = govatar.MALE
		} else {
			gender = govatar.FEMALE
		}

		img, err := generateAvator(gender)
		if err != nil {
			web.Console.Log(err.Error())
		} else {
			imgDom.SetAttribute("src", fmt.Sprintf("data:image/png;base64,%s", img))
		}
	})

	select {}
}
