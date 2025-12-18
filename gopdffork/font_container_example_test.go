package gopdffork

import (
	"io"
	"os"
)

func ExampleFontContainer_AddTTFFont() {
	fontContainer := &FontContainer{}
	err := fontContainer.AddTTFFont("LiberationSerif-Regular", "path/to/LiberationSerif-Regular.ttf")
	if err != nil {
		// handle error
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontFromFontContainer("LiberationSerif-Regular", fontContainer)
	if err != nil {
		// handle error
	}
}

func ExampleFontContainer_AddTTFFontWithOption() {
	fontContainer := &FontContainer{}
	err := fontContainer.AddTTFFontWithOption(
		"LiberationSerif-Regular",
		"path/to/LiberationSerif-Regular.ttf",
		TtfOption{})
	if err != nil {
		// handle error
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontFromFontContainer("LiberationSerif-Regular", fontContainer)
	if err != nil {
		// handle error
	}
}

func ExampleFontContainer_AddTTFFontByReader() {
	ttf, err := os.Open("path/to/LiberationSerif-Regular.ttf")
	if err != nil {
		// handle error
	}
	defer ttf.Close()

	fontContainer := &FontContainer{}
	err = fontContainer.AddTTFFontByReader("LiberationSerif-Regular", ttf)
	if err != nil {
		// handle error
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontFromFontContainer("LiberationSerif-Regular", fontContainer)
	if err != nil {
		// handle error
	}
}

func ExampleFontContainer_AddTTFFontByReaderWithOption() {
	ttf, err := os.Open("path/to/LiberationSerif-Regular.ttf")
	if err != nil {
		// handle error
	}
	defer ttf.Close()

	fontContainer := &FontContainer{}
	err = fontContainer.AddTTFFontByReaderWithOption("LiberationSerif-Regular", ttf, TtfOption{})
	if err != nil {
		// handle error
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontFromFontContainer("LiberationSerif-Regular", fontContainer)
	if err != nil {
		// handle error
	}
}

func ExampleFontContainer_AddTTFFontData() {
	ttf, err := os.Open("path/to/LiberationSerif-Regular.ttf")
	if err != nil {
		// handle error
	}
	defer ttf.Close()

	fontData, err := io.ReadAll(ttf)
	if err != nil {
		// handle error
	}

	fontContainer := &FontContainer{}
	err = fontContainer.AddTTFFontData("LiberationSerif-Regular", fontData)
	if err != nil {
		// handle error
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontFromFontContainer("LiberationSerif-Regular", fontContainer)
	if err != nil {
		// handle error
	}
}

func ExampleFontContainer_AddTTFFontDataWithOption() {
	ttf, err := os.Open("path/to/LiberationSerif-Regular.ttf")
	if err != nil {
		// handle error
	}
	defer ttf.Close()

	fontData, err := io.ReadAll(ttf)
	if err != nil {
		// handle error
	}

	fontContainer := &FontContainer{}
	err = fontContainer.AddTTFFontDataWithOption("LiberationSerif-Regular", fontData, TtfOption{})
	if err != nil {
		// handle error
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontFromFontContainer("LiberationSerif-Regular", fontContainer)
	if err != nil {
		// handle error
	}
}

func ExampleGoPdf_AddTTFFontFromFontContainer() {
	fontContainer := &FontContainer{}
	err := fontContainer.AddTTFFontWithOption(
		"LiberationSerif-Regular",
		"path/to/LiberationSerif-Regular.ttf",
		TtfOption{})
	if err != nil {
		// handle error
	}
	pdf := &GoPdf{}
	pdf.Start(Config{PageSize: *PageSizeA4})
	err = pdf.AddTTFFontFromFontContainer("LiberationSerif-Regular", fontContainer)
	if err != nil {
		// handle error
	}
}
