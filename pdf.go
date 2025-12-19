package gopdf

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"io"
	"log/slog"
	"sync"

	"os"

	"github.com/skirrund/go-pdf/font"
	"github.com/skirrund/go-pdf/gopdffork"
	"github.com/skirrund/go-pdf/pdfimporter"
)

var fontBytes []byte

var bufferPool = &sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

func getByteBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func releaseByteBuffer(buffer *bytes.Buffer) {
	buffer.Reset()
	bufferPool.Put(buffer)
}

func init() {
	var err error
	fontBytes, err = font.Asset("font/font.ttf")
	d, _ := os.Getwd()

	//fontBytes, err = os.ReadFile(d + "/font/font.ttf")
	if err != nil {
		slog.Error("[PDF] can not find font:" + d + "/font/font.ttf")
	} else {
		slog.Info("[PDF] load font:" + d + "/font/font.ttf")
	}
}

// func readTempFile(templateFile string) (*io.ReadSeeker, error) {
// 	ex, err := Exist(templateFile)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if !ex {
// 		return nil, errors.New("templateFile not exist" + templateFile)
// 	}
// 	bs, err := os.ReadFile(templateFile)
// 	if err != nil {
// 		return nil, err
// 	}
// 	reader := bytes.NewReader(bs)
// 	rd := io.ReadSeeker(reader)
// 	return &rd, nil
// }

func readTempFileToBytes(templateFile string) ([]byte, error) {
	ex, err := Exist(templateFile)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, errors.New("templateFile not exist" + templateFile)
	}
	bs, err := os.ReadFile(templateFile)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func readTempFileBytes(templateFile []byte) (*io.ReadSeeker, error) {
	if len(templateFile) == 0 {
		return nil, errors.New("templateFile error length is 0")
	}
	reader := bytes.NewReader(templateFile)
	rd := io.ReadSeeker(reader)
	return &rd, nil
}

func AddKeywords(locations []*PDFSearchLocation, templateFile string, saveasFilepath string, useTempPageSize bool) (err error) {
	bytes, err := readTempFileToBytes(templateFile)
	if err != nil {
		//logger.Logger.Error("[PDF] readTempFile", err)
		return err
	}
	pg, err := doAddKeywordsBytes(locations, bytes, useTempPageSize)
	if err != nil {
		return err
	}
	return pg.WritePdf(saveasFilepath)
}

func New() (*gopdffork.GoPdf, error) {
	if len(fontBytes) == 0 {
		return nil, errors.New("[PDF] can not find font file")
	}
	pdf := &gopdffork.GoPdf{}
	pdf.Start(gopdffork.Config{PageSize: *gopdffork.PageSizeA4})
	err := pdf.AddTTFFontByReader("song", bytes.NewReader(fontBytes))
	return pdf, err
}

func Merge(gp *gopdffork.GoPdf, other *os.File) {
	importer := pdfimporter.NewImporter()
	rs := io.ReadSeeker(other)
	importer.SetSourceStream(&rs)
	num := importer.GetNumPages()
	pageSizes := importer.GetPageSizes()
	for i := 1; i <= num; i++ {
		tempW := pageSizes[i]["/MediaBox"]["w"]
		tempH := pageSizes[i]["/MediaBox"]["h"]
		gp.AddPageWithOption(gopdffork.PageOption{PageSize: &gopdffork.Rect{W: tempW, H: tempH}})
		tpl := gp.ImportPageStream(&rs, i, "/MediaBox")
		gp.UseImportedTemplate(tpl, 0, 0, tempW, tempH)
	}
}

func MergeImage(gp *gopdffork.GoPdf, imageFilePath string) error {
	if img, err := os.Open(imageFilePath); err != nil {
		return nil
	} else {
		if img1, _, err := image.Decode(img); err != nil {
			return err
		} else {
			iw := min(float64(img1.Bounds().Dx()), gopdffork.PageSizeA4.W)
			ih := min(float64(img1.Bounds().Dy()), gopdffork.PageSizeA4.H)
			rect := &gopdffork.Rect{W: iw, H: ih}
			if imgHolder, err := gopdffork.ImageHolderByPath(imageFilePath); err != nil {
				return err
			} else {
				gp.AddPageWithOption(gopdffork.PageOption{PageSize: gopdffork.PageSizeA4})
				gp.ImageByHolder(imgHolder, 0, 0, rect)
			}
		}
	}
	return nil
}

func AddKeywordsBytes(locations []*PDFSearchLocation, templateFile []byte, useTempPageSize bool) (bs []byte, err error) {
	pg, err := doAddKeywordsBytes(locations, templateFile, useTempPageSize)
	if err != nil {
		return templateFile, err
	}
	buffer := getByteBuffer()
	defer releaseByteBuffer(buffer)
	bs, err = pg.GetBytesPdfReturnErr()
	return
	// _, err = pg.WriteTo(buffer)
	// if err != nil {
	// 	return templateFile, nil
	// }
	// bs = make([]byte, buffer.Len())
	// buffer.Read(bs)
	// return bs, nil
}

func doAddKeywordsBytes(locations []*PDFSearchLocation, templateFile []byte, useTempPageSize bool) (gp *gopdffork.GoPdf, err error) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("[PDF] recover :", err)
		}
	}()
	pdf := &gopdffork.GoPdf{}
	pdf.Start(gopdffork.Config{PageSize: *gopdffork.PageSizeA4})
	if len(fontBytes) == 0 {
		return nil, errors.New("[PDF] can not find font file")
	}
	err = pdf.AddTTFFontByReader("song", bytes.NewReader(fontBytes))
	if err != nil {
		//logger.Logger.Error("[PDF] AddTTFFontByReader error", err)
		return
	}
	rd, err := readTempFileBytes(templateFile)
	if err != nil {
		//logger.Logger.Error("[PDF] readTempFile", err)
		return nil, err
	}

	importer := pdfimporter.NewImporter()
	importer.SetSourceStream(rd)
	num := importer.GetNumPages()
	pageSizes := importer.GetPageSizes()
	for i := 1; i <= num; i++ {
		tempW := gopdffork.PageSizeA4.W
		tempH := gopdffork.PageSizeA4.H
		if useTempPageSize {
			tempW = pageSizes[i]["/MediaBox"]["w"]
			tempH = pageSizes[i]["/MediaBox"]["h"]
			pdf.AddPageWithOption(gopdffork.PageOption{PageSize: &gopdffork.Rect{W: tempW, H: tempH}})
			tpl := pdf.ImportPageStream(rd, i, "/MediaBox")
			pdf.UseImportedTemplate(tpl, 0, 0, tempW, tempH)
		} else {
			pdf.AddPage()
			tpl := pdf.ImportPageStream(rd, i, "/MediaBox")
			pdf.UseImportedTemplate(tpl, 0, 0, tempW, tempH)
		}
		for _, sl := range locations {
			if sl.Page <= 0 {
				sl.Page = 1
			}
			if sl.Page == i {
				img := sl.Image
				if img != nil {
					var rect *gopdffork.Rect
					if img.Width > 0 && img.Height > 0 {
						rect = &gopdffork.Rect{W: img.Width, H: img.Height}
					}
					var imgHolder gopdffork.ImageHolder
					if len(img.FilePath) > 0 {
						imgHolder, err = gopdffork.ImageHolderByPath(img.FilePath)
					}
					if len(img.Base64) > 0 {
						var bs []byte
						bs, err = base64.StdEncoding.DecodeString(img.Base64)
						if err != nil {
							return nil, err
						}
						imgHolder, err = gopdffork.ImageHolderByBytes(bs)
					}
					if err != nil {
						return nil, err
					}
					if imgHolder != nil {
						err = pdf.ImageByHolder(imgHolder, sl.AbsX, tempH-sl.AbsY, rect)
						if err != nil {
							return nil, err
						}
					}
					continue
				}

				if sl.FontSize <= 0 {
					sl.FontSize = 10.0
				}
				err = pdf.SetFont("song", "", int(sl.FontSize))
				if err != nil {
					//logger.Logger.Error("[PDF] SetFont error", err)
					return nil, err
				}
				pdf.SetTextColor(sl.BaseColor.Red, sl.BaseColor.Green, sl.BaseColor.Blue)
				if sl.BaseColor.Alpha == 0 {
					sl.BaseColor.Alpha = 1
				}
				pdf.SetTransparency(gopdffork.Transparency{
					Alpha: sl.BaseColor.Alpha,
				})

				pdf.SetX(sl.AbsX)
				pdf.SetY(tempH - sl.AbsY)
				pdf.Text(sl.AddText)
			}
		}
	}
	return pdf, nil
}

// 水平页面居中
func AlignCenterInPageX(gp *gopdffork.GoPdf, pageSize *gopdffork.Rect, text string) (float64, error) {
	width, err := gp.MeasureTextWidth(text)
	if err != nil {
		return 0, nil
	}
	return (pageSize.W - width) / 2, nil
}
